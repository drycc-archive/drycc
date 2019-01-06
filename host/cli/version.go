package cli

import (
	"fmt"

	"github.com/drycc/drycc/pkg/version"
	"github.com/drycc/go-docopt"
)

func init() {
	Register("version", runVersion, `
usage: drycc-host version [--release]

Options:
	--release   Print the release version

Show current version.
`)
}

func runVersion(args *docopt.Args) {
	if args.Bool["--release"] {
		fmt.Println(version.Release())
	} else {
		fmt.Println(version.String())
	}
}
