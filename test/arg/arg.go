package arg

import (
	"flag"

	"github.com/drycc/drycc/test/cluster"
)

type Args struct {
	BootConfig       cluster.BootConfig
	CLI              string
	DryccHost        string
	RootFS           string
	RouterIP         string
	Build            bool
	Debug            bool
	Stream           bool
	Kill             bool
	BuildRootFS      bool
	AssetsDir        string
	Run              string
	Gist             bool
	ClusterAPI       string
	Concurrency      int
	ConcurrentBuilds int
	Interactive      bool
}

func Parse() *Args {
	args := &Args{BootConfig: cluster.BootConfig{}}

	flag.StringVar(&args.BootConfig.Kernel, "kernel", "rootfs/vmlinuz", "path to the Linux binary")
	flag.StringVar(&args.BootConfig.BackupsDir, "backups-dir", "", "directory containing historic cluster backups")
	flag.StringVar(&args.RootFS, "rootfs", "rootfs/rootfs.img", "filesystem image to use with QEMU")
	flag.StringVar(&args.CLI, "cli", "drycc", "path to drycc-cli binary")
	flag.StringVar(&args.DryccHost, "drycc-host", "drycc-host", "path to drycc-host binary")
	flag.StringVar(&args.RouterIP, "router-ip", "127.0.0.1", "IP address of the router")
	flag.StringVar(&args.AssetsDir, "assets", "runner/assets", "path to the runner assets dir")
	flag.StringVar(&args.Run, "run", "", "regular expression selecting which tests and/or suites to run")
	flag.StringVar(&args.ClusterAPI, "cluster-api", "", "cluster-api endpoint for adding and removing hosts")
	flag.BoolVar(&args.Build, "build", true, "build Drycc")
	flag.BoolVar(&args.Debug, "debug", false, "enable debug output")
	flag.BoolVar(&args.Stream, "stream", false, "stream debug output (implies --debug)")
	flag.BoolVar(&args.Kill, "kill", true, "kill the cluster after running the tests")
	flag.BoolVar(&args.BuildRootFS, "build-rootfs", false, "just build the rootfs (leaving it behind for future use) without running tests")
	flag.BoolVar(&args.Gist, "gist", false, "upload debug info to a gist")
	flag.BoolVar(&args.Interactive, "interactive", false, "start an interactive bash shell when cluster tests fail")
	flag.IntVar(&args.Concurrency, "concurrency", 5, "max number of concurrent tests")
	flag.IntVar(&args.ConcurrentBuilds, "concurrent-builds", 5, "max number of concurrent builds")
	flag.Parse()

	return args
}
