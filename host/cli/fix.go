package cli

import (
	"github.com/drycc/drycc/host/fixer"
)

func init() {
	Register("fix", (&fixer.ClusterFixer{}).Run, `
usage: drycc-host fix [options]

Attempts to fix a broken cluster by starting missing jobs.

Options:
    -n, --min-hosts=<n>  minimum expected number of hosts (required)
	--peer-ips=<iplist>  list of host IPs (required if discoverd is down)
`)
}
