package schedutil

import (
	"github.com/drycc/drycc/pkg/cluster"
	"github.com/drycc/drycc/pkg/random"
)

type HostSlice []*cluster.Host

func PickHost(hosts HostSlice) *cluster.Host {
	if len(hosts) == 0 {
		return nil
	}
	// Return a random pick
	return hosts[random.Math.Intn(len(hosts))]
}
