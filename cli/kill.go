package main

import (
	"errors"
	"log"

	"github.com/drycc/drycc/controller/client"
	"github.com/drycc/go-docopt"
)

func init() {
	register("kill", runKill, `
usage: drycc kill <job>...

Kill running jobs.`)
}

func runKill(args *docopt.Args, client controller.Client) error {
	success := true
	for _, job := range args.All["<job>"].([]string) {
		if err := client.DeleteJob(mustApp(), job); err != nil {
			success = false
			log.Printf("ERROR: could not kill job %s: %s\n", job, err)
			continue
		}
		log.Printf("Job %s killed.", job)
	}
	if !success {
		return errors.New("Could not kill all jobs.")
	}
	return nil
}
