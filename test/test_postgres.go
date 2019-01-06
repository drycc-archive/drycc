package main

import (
	"fmt"
	"path/filepath"
	"strings"

	ct "github.com/drycc/drycc/controller/types"
	"github.com/drycc/drycc/discoverd/client"
	"github.com/drycc/drycc/pkg/postgres"
	c "github.com/drycc/go-check"
)

type PostgresSuite struct {
	Helper
}

var _ = c.ConcurrentSuite(&PostgresSuite{})

// Check postgres config to avoid regressing on https://github.com/drycc/drycc/issues/101
func (s *PostgresSuite) TestSSLRenegotiationLimit(t *c.C) {
	query := drycc(t, "/", "-a", "controller", "pg", "psql", "--", "-c", "SHOW ssl_renegotiation_limit")
	t.Assert(query, SuccessfulOutputContains, "ssl_renegotiation_limit \n-------------------------\n 0\n(1 row)")
}

func (s *PostgresSuite) TestDumpRestore(t *c.C) {
	r := s.newGitRepo(t, "empty")
	t.Assert(r.drycc("create"), Succeeds)

	res := r.drycc("resource", "add", "postgres")
	t.Assert(res, Succeeds)
	id := strings.Split(res.Output, " ")[2]

	t.Assert(r.drycc("pg", "psql", "--", "-c",
		"CREATE table foos (data text); INSERT INTO foos (data) VALUES ('foobar')"), Succeeds)

	file := filepath.Join(t.MkDir(), "db.dump")
	t.Assert(r.drycc("pg", "dump", "-f", file), Succeeds)
	t.Assert(r.drycc("pg", "psql", "--", "-c", "DROP TABLE foos"), Succeeds)

	r.drycc("pg", "restore", "-f", file)

	query := r.drycc("pg", "psql", "--", "-c", "SELECT * FROM foos")
	t.Assert(query, SuccessfulOutputContains, "foobar")

	t.Assert(r.drycc("resource", "remove", "postgres", id), Succeeds)
}

var sireniaPostgres = sireniaDatabase{
	appName:    "postgres",
	serviceKey: "DRYCC_POSTGRES",
	hostKey:    "PGHOST",
	initDb: func(t *c.C, r *ct.Release, d *sireniaDeploy) {
		db := postgres.Wait(&postgres.Conf{
			Discoverd: discoverd.NewClientWithURL(fmt.Sprintf("http://%s:1111", routerIP)),
			Service:   d.name,
			User:      "drycc",
			Password:  r.Env["PGPASSWORD"],
			Database:  "postgres",
		}, nil)
		dbname := "deploy-test"
		t.Assert(db.Exec(fmt.Sprintf(`CREATE DATABASE "%s" WITH OWNER = "drycc"`, dbname)), c.IsNil)
		db.Close()
		db = postgres.Wait(&postgres.Conf{
			Discoverd: discoverd.NewClientWithURL(fmt.Sprintf("http://%s:1111", routerIP)),
			Service:   d.name,
			User:      "drycc",
			Password:  r.Env["PGPASSWORD"],
			Database:  dbname,
		}, nil)
		defer db.Close()
		t.Assert(db.Exec(`CREATE TABLE deploy_test ( data text)`), c.IsNil)
	},
	assertWriteable: func(t *c.C, r *ct.Release, d *sireniaDeploy) {
		dbname := "deploy-test"
		db := postgres.Wait(&postgres.Conf{
			Discoverd: discoverd.NewClientWithURL(fmt.Sprintf("http://%s:1111", routerIP)),
			Service:   d.name,
			User:      "drycc",
			Password:  r.Env["PGPASSWORD"],
			Database:  dbname,
		}, nil)
		defer db.Close()
		debug(t, "writing to postgres database")
		t.Assert(db.ExecRetry(`INSERT INTO deploy_test (data) VALUES ('data')`), c.IsNil)
	},
}

// Sirenia integration tests
func (s *PostgresSuite) TestDeploySingleAsync(t *c.C) {
	testSireniaDeploy(s.controllerClient(t), s.discoverdClient(t), t, &sireniaDeploy{
		name:        "postgres-single-async",
		db:          sireniaPostgres,
		sireniaJobs: 3,
		webJobs:     2,
		expected:    testDeploySingleAsync,
	})
}

func (s *PostgresSuite) TestDeployMultipleAsync(t *c.C) {
	testSireniaDeploy(s.controllerClient(t), s.discoverdClient(t), t, &sireniaDeploy{
		name:        "postgres-multiple-async",
		db:          sireniaPostgres,
		sireniaJobs: 5,
		webJobs:     2,
		expected:    testDeployMultipleAsync,
	})
}
