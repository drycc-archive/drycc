package tufconfig

import (
	"encoding/json"

	"github.com/drycc/go-tuf/data"
)

var (
	// these constants are overridden at build time (see builder/go-wrapper.sh)
	RootKeysJSON = `[{"keytype":"ed25519","keyval":{"public":"6cfda23aa48f530aebd5b9c01030d06d02f25876b5508d681675270027af4731"}}]`
	Repository   = "https://dl.drycc.cc/tuf"
)

var RootKeys []*data.Key

func init() {
	if err := json.Unmarshal([]byte(RootKeysJSON), &RootKeys); err != nil {
		panic("error decoding root keys")
	}
}
