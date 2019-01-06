package main

import (
	"fmt"

	"github.com/drycc/go-docopt"
)

func init() {
	register("install", runInstaller, `usage: drycc install`)
}

func runInstaller(args *docopt.Args) error {
	fmt.Printf("DEPRECATED: `drycc install` has been deprecated.\nRefer to https://drycc.cc/docs/installation for current installation instructions.\nAn unsupported and unmaintained snapshot of the installer binaries at the time of deprecation is available at https://dl.drycc.cc/drycc-install-deprecated.tar.gz\n")
	return nil
}
