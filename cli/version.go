package main

import (
	"fmt"

	"github.com/drycc/drycc/pkg/version"
)

func init() {
	register("version", runVersion, `
usage: drycc version

Show drycc version string.
`)
}

func runVersion() {
	fmt.Println(version.String())
}
