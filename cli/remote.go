package main

import (
	"log"
	"os/exec"

	"github.com/drycc/drycc/controller/client"
	"github.com/drycc/go-docopt"
)

func init() {
	register("remote", runRemote, `
usage: drycc remote add [<remote>] [-y]

Create a git remote that allows deploying the application via git.
If a name for the remote is not provided 'drycc' will be used.

Note that the -a <app> option must be given so the remote to add is known.

Options:
	-y, --yes              Skip the confirmation prompt if the git remote already exists.

Examples:

	$ drycc -a turkeys-stupefy-perry remote add
	Created remote drycc with url https://git.dev.local.drycc.cc/turkeys-stupefy-perry.git

	$ drycc -a turkeys-stupefy-perry remote add staging
	Created remote staging with url https://git.dev.local.drycc.cc/turkeys-stupefy-perry.git
`)
}

func runRemote(args *docopt.Args, client controller.Client) error {
	app, err := client.GetApp(mustApp())
	if err != nil {
		return err
	}

	remote := args.String["<remote>"]
	if remote == "" {
		remote = "drycc"
	}

	if !inGitRepo() {
		log.Print("Must be executed within a git repository.")
		return nil
	}

	if !args.Bool["--yes"] {
		update, err := promptReplaceRemote(remote)
		if err != nil {
			return err
		}
		if update == false {
			return nil
		}
	}

	// Register git remote
	url := gitURL(clusterConf, app.Name)
	exec.Command("git", "remote", "remove", remote).Run()
	exec.Command("git", "remote", "add", "--", remote, url).Run()

	log.Printf("Created remote %s with url %s.", remote, url)
	return nil
}
