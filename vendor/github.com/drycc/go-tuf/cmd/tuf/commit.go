package main

import (
	"github.com/drycc/go-docopt"
	"github.com/drycc/go-tuf"
)

func init() {
	register("commit", cmdCommit, `
usage: tuf commit

Commit staged files to the repository.
`)
}

func cmdCommit(args *docopt.Args, repo *tuf.Repo) error {
	return repo.Commit()
}
