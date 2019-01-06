package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/docker/go-units"
	ct "github.com/drycc/drycc/controller/types"
	"github.com/drycc/drycc/host/downloader"
	"github.com/drycc/drycc/host/volume"
	"github.com/drycc/drycc/host/volume/manager"
	"github.com/drycc/drycc/host/volume/zfs"
	"github.com/drycc/drycc/pkg/tufconfig"
	"github.com/drycc/drycc/pkg/tufutil"
	"github.com/drycc/drycc/pkg/version"
	"github.com/drycc/go-docopt"
	tuf "github.com/drycc/go-tuf/client"
	"github.com/inconshreveable/log15"
)

func init() {
	Register("download", runDownload, `
usage: drycc-host download [--repository=<uri>] [--tuf-db=<path>] [--config-dir=<dir>] [--bin-dir=<dir>] [--volpath=<path>]

Options:
  -r --repository=<uri>    TUF repository URI [default: https://dl.drycc.cc/tuf]
  -t --tuf-db=<path>       local TUF file [default: /etc/drycc/tuf.db]
  -c --config-dir=<dir>    config directory [default: /etc/drycc]
  -b --bin-dir=<dir>       binary directory [default: /usr/local/bin]
  -v --volpath=<path>      directory to create volumes in [default: /var/lib/drycc/volumes]

Download container images and Drycc binaries from a TUF repository.

Set DRYCC_VERSION to download an explicit version.`)
}

func runDownload(args *docopt.Args) error {
	log := log15.New()

	log.Info("initializing ZFS volumes")
	volPath := args.String["--volpath"]
	volDB := filepath.Join(volPath, "volumes.bolt")
	volMan := volumemanager.New(volDB, log, func() (volume.Provider, error) {
		return zfs.NewProvider(&zfs.ProviderConfig{
			DatasetName: zfs.DefaultDatasetName,
			Make:        zfs.DefaultMakeDev(volPath, log),
			WorkingDir:  filepath.Join(volPath, "zfs"),
		})
	})
	if err := volMan.OpenDB(); err != nil {
		log.Error("error opening volume database, make sure drycc-host is not running", "err", err)
		return err
	}

	// create a TUF client and update it
	log.Info("initializing TUF client")
	tufDB := args.String["--tuf-db"]
	local, err := tuf.FileLocalStore(tufDB)
	if err != nil {
		log.Error("error creating local TUF client", "err", err)
		return err
	}
	remote, err := tuf.HTTPRemoteStore(args.String["--repository"], tufHTTPOpts("downloader"))
	if err != nil {
		log.Error("error creating remote TUF client", "err", err)
		return err
	}
	client := tuf.NewClient(local, remote)
	if err := updateTUFClient(client); err != nil {
		log.Error("error updating TUF client", "err", err)
		return err
	}

	configDir := args.String["--config-dir"]

	requestedVersion := os.Getenv("DRYCC_VERSION")
	if requestedVersion == "" {
		requestedVersion, err = getChannelVersion(configDir, client, log)
		if err != nil {
			return err
		}
	}
	log.Info(fmt.Sprintf("downloading components with version %s", requestedVersion))

	d := downloader.New(client, volMan, requestedVersion)

	binDir := args.String["--bin-dir"]
	log.Info(fmt.Sprintf("downloading binaries to %s", binDir))
	if _, err := d.DownloadBinaries(binDir); err != nil {
		log.Error("error downloading binaries", "err", err)
		return err
	}

	// use the requested version of drycc-host to download the images as
	// the format changed in v20161106
	if version.Release() != requestedVersion {
		log.Info(fmt.Sprintf("executing %s drycc-host binary", requestedVersion))
		binPath := filepath.Join(binDir, "drycc-host")
		argv := append([]string{binPath}, os.Args[1:]...)
		return syscall.Exec(binPath, argv, os.Environ())
	}

	log.Info("downloading images")
	ch := make(chan *ct.ImagePullInfo)
	go func() {
		for info := range ch {
			switch info.Type {
			case ct.ImagePullTypeImage:
				log.Info(fmt.Sprintf("pulling %s image", info.Name))
			case ct.ImagePullTypeLayer:
				log.Info(fmt.Sprintf("pulling %s layer %s (%s)",
					info.Name, info.Layer.ID, units.BytesSize(float64(info.Layer.Length))))
			}
		}
	}()
	if err := d.DownloadImages(configDir, ch); err != nil {
		log.Error("error downloading images", "err", err)
		return err
	}

	log.Info(fmt.Sprintf("downloading config to %s", configDir))
	if _, err := d.DownloadConfig(configDir); err != nil {
		log.Error("error downloading config", "err", err)
		return err
	}

	log.Info("download complete")
	return nil
}

func tufHTTPOpts(name string) *tuf.HTTPRemoteOptions {
	return &tuf.HTTPRemoteOptions{
		UserAgent: fmt.Sprintf("drycc-host/%s %s-%s %s", version.String(), runtime.GOOS, runtime.GOARCH, name),
		Retries:   tufutil.DefaultHTTPRetries,
	}
}

// updateTUFClient updates the given client, initializing and re-running the
// update if ErrNoRootKeys is returned.
func updateTUFClient(client *tuf.Client) error {
	_, err := client.Update()
	if err == nil || tuf.IsLatestSnapshot(err) {
		return nil
	}
	if err == tuf.ErrNoRootKeys {
		if err := client.Init(tufconfig.RootKeys, len(tufconfig.RootKeys)); err != nil {
			return err
		}
		return updateTUFClient(client)
	}
	return err
}

// getChannelVersion reads the locally configured release channel from
// <configDir>/channel.txt then gets the latest version for that channel
// using the TUF client.
func getChannelVersion(configDir string, client *tuf.Client, log log15.Logger) (string, error) {
	log.Info("getting configured release channel")
	data, err := ioutil.ReadFile(filepath.Join(configDir, "channel.txt"))
	if err != nil {
		log.Error("error getting configured release channel", "err", err)
		return "", err
	}
	channel := strings.TrimSpace(string(data))

	log.Info(fmt.Sprintf("determining latest version of %s release channel", channel))
	version, err := tufutil.DownloadString(client, path.Join("channels", channel))
	if err != nil {
		log.Error("error determining latest version", "err", err)
		return "", err
	}
	version = strings.TrimSpace(version)
	log.Info(fmt.Sprintf("latest %s version is %s", channel, version))
	return version, nil
}
