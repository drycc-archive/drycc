package main

import (
	"fmt"
	"strings"

	"github.com/drycc/drycc/controller/client"
	"github.com/drycc/drycc/controller/types"
	"github.com/drycc/go-docopt"
)

func init() {
	register("meta", runMeta, `
usage: drycc meta
       drycc meta set <var>=<val>...
       drycc meta unset <var>...

Manage metadata for an application.

Examples:

	$ drycc meta
	KEY  VALUE
	foo  bar

	$ drycc meta set foo=baz bar=qux

	$ drycc meta
	KEY  VALUE
	foo  baz
	bar  qux

	$ drycc meta unset foo

	$ drycc meta
	KEY  VALUE
	bar  qux
`)
}

func runMeta(args *docopt.Args, client controller.Client) error {
	app, err := client.GetApp(mustApp())
	if err != nil {
		return err
	}

	if args.Bool["set"] {
		return runMetaSet(app, args, client)
	} else if args.Bool["unset"] {
		return runMetaUnset(app, args, client)
	} else {
		return runMetaGet(app, args, client)
	}
}

func runMetaGet(app *types.App, args *docopt.Args, client controller.Client) error {
	w := tabWriter()
	defer w.Flush()
	listRec(w, "KEY", "VALUE")
	for k, v := range app.Meta {
		listRec(w, k, v)
	}
	return nil
}

func runMetaSet(app *types.App, args *docopt.Args, client controller.Client) error {
	pairs := args.All["<var>=<val>"].([]string)
	if app.Meta == nil {
		app.Meta = make(map[string]string, len(pairs))
	}
	for _, s := range pairs {
		v := strings.SplitN(s, "=", 2)
		if len(v) != 2 {
			return fmt.Errorf("invalid var format: %q", s)
		}
		app.Meta[v[0]] = v[1]
	}
	return client.UpdateAppMeta(app)
}

func runMetaUnset(app *types.App, args *docopt.Args, client controller.Client) error {
	vars := args.All["<var>"].([]string)
	for _, s := range vars {
		delete(app.Meta, s)
	}
	return client.UpdateAppMeta(app)
}
