package main

import (
	"github.com/drycc/go-docopt"
)

func main() {
	usage := `drycc-release generates Drycc releases.

Usage:
  drycc-release status <commit>
  drycc-release vagrant <url> <checksum> <version> <provider>
`
	args, _ := docopt.Parse(usage, nil, true, "", false)

	switch {
	case args.Bool["status"]:
		status(args)
	case args.Bool["vagrant"]:
		vagrant(args)
	}
}
