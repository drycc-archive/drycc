package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/drycc/drycc/appliance/mariadb"
	ct "github.com/drycc/drycc/controller/types"
	c "github.com/drycc/go-check"
	_ "github.com/go-sql-driver/mysql"
)

type MariaDBSuite struct {
	Helper
}

var _ = c.ConcurrentSuite(&MariaDBSuite{})

// Sirenia integration tests
var sireniaMariaDB = sireniaDatabase{
	appName:    "mariadb",
	serviceKey: "DRYCC_MYSQL",
	hostKey:    "MYSQL_HOST",
	initDb: func(t *c.C, r *ct.Release, d *sireniaDeploy) {
		dsn := &mariadb.DSN{
			Host:     fmt.Sprintf("leader.%s.discoverd", d.name) + ":3306",
			User:     "drycc",
			Password: r.Env["MYSQL_PWD"],
			Database: "mysql",
		}
		db, err := sql.Open("mysql", dsn.String())
		t.Assert(err, c.IsNil)
		defer db.Close()
		dbname := "deploy_test"
		_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE %s`, dbname))
		t.Assert(err, c.IsNil)
		_, err = db.Exec(fmt.Sprintf(`USE %s`, dbname))
		t.Assert(err, c.IsNil)
		_, err = db.Exec(`CREATE TABLE deploy_test (data TEXT)`)
		t.Assert(err, c.IsNil)
	},
	assertWriteable: func(t *c.C, r *ct.Release, d *sireniaDeploy) {
		dbname := "deploy_test"
		dsn := &mariadb.DSN{
			Host:     fmt.Sprintf("leader.%s.discoverd", d.name) + ":3306",
			User:     "drycc",
			Password: r.Env["MYSQL_PWD"],
			Database: dbname,
		}
		db, err := sql.Open("mysql", dsn.String())
		t.Assert(err, c.IsNil)
		defer db.Close()
		_, err = db.Exec(`INSERT INTO deploy_test (data) VALUES ('data')`)
		t.Assert(err, c.IsNil)
	},
}

func (s *MariaDBSuite) TestDeploySingleAsync(t *c.C) {
	testSireniaDeploy(s.controllerClient(t), s.discoverdClient(t), t, &sireniaDeploy{
		name:        "mariadb-single-async",
		db:          sireniaMariaDB,
		sireniaJobs: 3,
		webJobs:     2,
		expected:    testDeploySingleAsync,
	})
}

func (s *MariaDBSuite) TestDeployMultipleAsync(t *c.C) {
	testSireniaDeploy(s.controllerClient(t), s.discoverdClient(t), t, &sireniaDeploy{
		name:        "mariadb-multiple-async",
		db:          sireniaMariaDB,
		sireniaJobs: 5,
		webJobs:     2,
		expected:    testDeployMultipleAsync,
	})
}

func (s *MariaDBSuite) TestDumpRestore(t *c.C) {
	r := s.newGitRepo(t, "empty")
	t.Assert(r.drycc("create"), Succeeds)

	res := r.drycc("resource", "add", "mysql")
	t.Assert(res, Succeeds)
	id := strings.Split(res.Output, " ")[2]

	t.Assert(r.drycc("mysql", "console", "--", "-e",
		"CREATE TABLE T (F text); INSERT INTO T (F) VALUES ('abc')"), Succeeds)

	file := filepath.Join(t.MkDir(), "db.dump")
	t.Assert(r.drycc("mysql", "dump", "-f", file), Succeeds)
	t.Assert(r.drycc("mysql", "console", "--", "-e", "DROP TABLE T"), Succeeds)

	r.drycc("mysql", "restore", "-f", file)

	query := r.drycc("mysql", "console", "--", "-e", "SELECT * FROM T")
	t.Assert(query, SuccessfulOutputContains, "abc")

	t.Assert(r.drycc("resource", "remove", "mysql", id), Succeeds)
}
