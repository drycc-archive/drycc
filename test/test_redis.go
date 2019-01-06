package main

import (
	"fmt"
	"path/filepath"
	"strings"

	c "github.com/drycc/go-check"
)

type RedisSuite struct {
	Helper
}

var _ = c.ConcurrentSuite(&RedisSuite{})

func (s *RedisSuite) TestRedisEnv(t *c.C) {
	a := s.newCliTestApp(t)

	// create a redis resource
	t.Assert(a.drycc("resource", "add", "redis"), Succeeds)

	// get the new release
	client := s.controllerClient(t)
	release, err := client.GetAppRelease(a.id)
	t.Assert(err, c.IsNil)

	// check that DRYCC_REDIS points to a Redis app
	service, ok := release.Env["DRYCC_REDIS"]
	if !ok {
		t.Fatal("missing DRYCC_REDIS")
	}
	redisApp, err := client.GetApp(service)
	t.Assert(err, c.IsNil)
	redisRelease, err := client.GetAppRelease(redisApp.ID)
	t.Assert(err, c.IsNil)
	t.Assert(redisRelease.Processes, c.HasLen, 1)
	redisProc, ok := redisRelease.Processes["redis"]
	if !ok {
		t.Fatal("missing redis process")
	}
	t.Assert(redisProc.Service, c.Equals, service)
	password, ok := redisRelease.Env["REDIS_PASSWORD"]
	if !ok {
		t.Fatal("missing REDIS_PASSWORD")
	}

	// check that the release has valid redis env vars
	expected := map[string]string{
		"REDIS_HOST":     fmt.Sprintf("leader.%s.discoverd", service),
		"REDIS_PORT":     "6379",
		"REDIS_PASSWORD": password,
		"REDIS_URL":      fmt.Sprintf("redis://:%s@leader.%s.discoverd:6379", password, service),
	}
	for key, val := range expected {
		actual, ok := release.Env[key]
		if !ok {
			t.Fatalf("env missing key %q", key)
		}
		if actual != val {
			t.Fatalf("expected %s to be %s, got %s", key, val, actual)
		}
	}
}

func (s *RedisSuite) TestDumpRestore(t *c.C) {
	a := s.newCliTestApp(t)

	res := a.drycc("resource", "add", "redis")
	t.Assert(res, Succeeds)
	id := strings.Split(res.Output, " ")[2]

	release, err := s.controllerClient(t).GetAppRelease(a.id)
	t.Assert(err, c.IsNil)

	t.Assert(release.Env["DRYCC_REDIS"], c.Not(c.Equals), "")
	a.waitForService(release.Env["DRYCC_REDIS"])

	t.Assert(a.drycc("redis", "redis-cli", "set", "foo", "bar"), Succeeds)

	file := filepath.Join(t.MkDir(), "dump.rdb")
	t.Assert(a.drycc("redis", "dump", "-f", file), Succeeds)
	t.Assert(a.drycc("redis", "redis-cli", "del", "foo"), Succeeds)

	a.drycc("redis", "restore", "-f", file)

	query := a.drycc("redis", "redis-cli", "get", "foo")
	t.Assert(query, SuccessfulOutputContains, "bar")

	t.Assert(a.drycc("resource", "remove", "redis", id), Succeeds)
}
