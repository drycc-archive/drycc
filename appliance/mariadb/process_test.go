package mariadb

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/drycc/drycc/discoverd/client"
	"github.com/drycc/drycc/pkg/attempt"
	"github.com/drycc/drycc/pkg/sirenia/state"
	. "github.com/drycc/go-check"
	_ "github.com/go-sql-driver/mysql"
)

// Hook gocheck up to the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type MariaDBSuite struct{}

var _ = Suite(&MariaDBSuite{})

func (MariaDBSuite) TestSingletonPrimary(c *C) {
	p := NewProcess()
	p.ID = "node1"
	p.Singleton = true
	p.Password = "password"
	p.DataDir = c.MkDir()
	p.Port = "7500"
	p.ServerID = 1
	p.OpTimeout = 30 * time.Second
	err := p.Reconfigure(&state.Config{Role: state.RolePrimary})
	c.Assert(err, IsNil)

	err = p.Start()
	c.Assert(err, IsNil)
	defer p.Stop()

	conn := connect(c, p, "")
	_, err = conn.Exec("CREATE DATABASE test")
	conn.Close()
	c.Assert(err, IsNil)

	err = p.Stop()
	c.Assert(err, IsNil)

	// ensure that we can start a new instance from the same directory
	p = NewProcess()
	p.ID = "node1"
	p.Singleton = true
	p.Password = "password"
	p.DataDir = c.MkDir()
	p.Port = "7500"
	p.ServerID = 1
	p.OpTimeout = 30 * time.Second
	err = p.Reconfigure(&state.Config{Role: state.RolePrimary})
	c.Assert(err, IsNil)
	c.Assert(p.Start(), IsNil)
	defer p.Stop()

	conn = connect(c, p, "")
	_, err = conn.Exec("CREATE DATABASE foo")
	conn.Close()
	c.Assert(err, IsNil)

	err = p.Stop()
	c.Assert(err, IsNil)
}

func instance(p *Process) *discoverd.Instance {
	return &discoverd.Instance{
		ID:   p.ID,
		Addr: fmt.Sprintf("127.0.0.1:%d", MustAtoi(p.Port)),
		Meta: map[string]string{
			"MYSQL_ID": p.ID,
		},
	}
}

func connect(c *C, p *Process, database string) *sql.DB {
	dsn := DSN{
		Host:     fmt.Sprintf("127.0.0.1:%d", MustAtoi(p.Port)),
		User:     "root",
		Password: "",
		Database: database,
	}
	db, err := sql.Open("mysql", dsn.String())
	c.Assert(err, IsNil)
	return db
}

func Config(role state.Role, upstream, downstream *Process) *state.Config {
	c := &state.Config{Role: role}
	if upstream != nil {
		c.Upstream = instance(upstream)
	}
	if downstream != nil {
		c.Downstream = instance(downstream)
	}
	return c
}

var queryAttempts = attempt.Strategy{
	Min:   5,
	Total: 30 * time.Second,
	Delay: 200 * time.Millisecond,
}

func assertDownstream(c *C, db *sql.DB, n int) {
	var row struct {
		ServerID int64
		Host     string
		Port     int64
		MasterID int64
	}
	err := queryAttempts.Run(func() error {
		return db.QueryRow("SHOW SLAVE HOSTS").Scan(
			&row.ServerID,
			&row.Host,
			&row.Port,
			&row.MasterID,
		)
	})
	c.Assert(err, IsNil, Commentf("node%d", n))
	c.Assert(row.Host, Equals, fmt.Sprintf("node%d", n))
}

func waitRow(c *C, db *sql.DB, n int) {
	var res int64
	err := queryAttempts.Run(func() error {
		return db.QueryRow(fmt.Sprintf("SELECT id FROM test WHERE id = %d", n)).Scan(&res)
	})
	c.Assert(err, IsNil)
}

func createTable(c *C, db *sql.DB) {
	_, err := db.Exec("CREATE TABLE test (id bigint PRIMARY KEY)")
	c.Assert(err, IsNil)
	insertRow(c, db, 1)
}

func insertRow(c *C, db *sql.DB, n int) {
	_, err := db.Exec(fmt.Sprintf("INSERT INTO test (id) VALUES (%d)", n))
	c.Assert(err, IsNil)
}

func waitReadWrite(c *C, db *sql.DB) {
	// Check if the master has transitioned the database into read/write
	var readOnly string
	err := queryAttempts.Run(func() error {
		if err := db.QueryRow("SELECT @@read_only").Scan(&readOnly); err != nil {
			return err
		}
		if readOnly == "0" {
			return nil
		}
		return fmt.Errorf("database is read_only")
	})
	c.Assert(err, IsNil)

	// Even if the database is read/write a slave must be connected
	// for writes to be allowed.
	err = queryAttempts.Run(func() error {
		var discard interface{}
		var masterClients int
		err = db.QueryRow("SHOW STATUS LIKE 'rpl_semi_sync_master_clients'").Scan(&discard, &masterClients)
		if err != nil {
			return err
		}
		if masterClients > 0 {
			return nil
		}
		return fmt.Errorf("no connected slave")
	})
	c.Assert(err, IsNil)
}

var syncAttempts = attempt.Strategy{
	Min:   5,
	Total: 30 * time.Second,
	Delay: 200 * time.Millisecond,
}

func waitReplSync(c *C, p *Process, n int) {
	id := fmt.Sprintf("node%d", n)
	err := syncAttempts.Run(func() error {
		info, err := p.Info()
		if err != nil {
			return err
		}
		if info.SyncedDownstream == nil || info.SyncedDownstream.ID != id {
			return errors.New("downstream not synced")
		}
		return nil
	})
	c.Assert(err, IsNil, Commentf("up:%s down:%s", p.ID, id))
}

func (MariaDBSuite) TestIntegration_TwoNodeSync(c *C) {
	node1 := NewTestProcess(c, 1)
	node2 := NewTestProcess(c, 2)

	// Start a primary.
	err := node1.Reconfigure(Config(state.RolePrimary, nil, node2))
	c.Assert(err, IsNil)
	c.Assert(node1.Start(), IsNil)
	defer node1.Stop()

	srv1 := NewHTTPServer(c, node1)
	defer srv1.Close()

	// Start a sync
	err = node2.Reconfigure(Config(state.RoleSync, node1, nil))
	c.Assert(err, IsNil)
	c.Assert(node2.Start(), IsNil)
	defer node2.Stop()

	srv2 := NewHTTPServer(c, node2)
	defer srv2.Close()

	// check it catches up
	waitReplSync(c, node1, 2)

	// Write to the master.
	db1 := connect(c, node1, "mysql")
	defer db1.Close()
	createTable(c, db1)

	// Read from the slave.
	db2 := connect(c, node2, "mysql")
	defer db2.Close()
	waitRow(c, db2, 1)
}

func (MariaDBSuite) TestIntegration_FourNode(c *C) {
	node1 := NewTestProcess(c, 1)
	node2 := NewTestProcess(c, 2)
	node3 := NewTestProcess(c, 3)
	node4 := NewTestProcess(c, 4)

	// Start a primary
	c.Log("starting primary (node1)")
	err := node1.Reconfigure(Config(state.RolePrimary, nil, node2))
	c.Assert(err, IsNil)
	c.Assert(node1.Start(), IsNil)
	defer node1.Stop()

	srv1 := NewHTTPServer(c, node1)
	defer srv1.Close()

	// Connect to the primary and make sure it's read-only
	c.Log("checking primary (node1) is read-only")
	db1 := connect(c, node1, "mysql")
	defer db1.Close()

	// There should currently be no connected semi-sync peers
	var discard interface{}
	var masterClients int
	err = db1.QueryRow("SHOW STATUS LIKE 'rpl_semi_sync_master_clients'").Scan(&discard, &masterClients)
	c.Assert(err, IsNil)
	c.Assert(masterClients, Equals, 0)

	var masterStatus string
	err = db1.QueryRow("SHOW STATUS LIKE 'rpl_semi_sync_master_status'").Scan(&discard, &masterStatus)
	c.Assert(err, IsNil)
	c.Assert(masterStatus, Equals, "ON")

	// Start a sync
	c.Log("starting sync (node2)")
	err = node2.Reconfigure(Config(state.RoleSync, node1, node3))
	c.Assert(err, IsNil)
	c.Assert(node2.Start(), IsNil)
	defer node2.Stop()

	srv2 := NewHTTPServer(c, node2)
	defer srv2.Close()

	// check it catches up
	c.Log("waiting for primary (node1) to indicate downstream is sync (node2)")
	waitReplSync(c, node1, 2)

	// try to query primary until it comes up as read-write
	c.Log("waiting for primary (node1) to become read-write")
	waitReadWrite(c, db1)

	for _, n := range []*Process{node1, node2} {
		pos, err := n.XLogPosition()
		c.Assert(err, IsNil)
		c.Assert(pos, Not(Equals), "")
		c.Assert(pos, Not(Equals), "master-bin.000000/0")
	}

	// make sure the sync is listed as sync and semi-sync is configured
	c.Log("assert primary (node1) downstream is sync (node2)")
	assertDownstream(c, db1, 2)

	err = db1.QueryRow("SHOW STATUS LIKE 'rpl_semi_sync_master_clients'").Scan(&discard, &masterClients)
	c.Assert(err, IsNil)
	c.Assert(masterClients, Equals, 1)

	err = db1.QueryRow("SHOW STATUS LIKE 'rpl_semi_sync_master_status'").Scan(&discard, &masterStatus)
	c.Assert(err, IsNil)
	c.Assert(masterStatus, Equals, "ON")

	// create a table and a row
	c.Log("write to primary (node1)")
	createTable(c, db1)
	db1.Close()

	// query the sync and see the database
	c.Log("read from sync (node2)")
	db2 := connect(c, node2, "mysql")
	defer db2.Close()
	waitRow(c, db2, 1)

	// Start an async
	c.Log("starting async (node3)")
	err = node3.Reconfigure(Config(state.RoleAsync, node2, node4))
	c.Assert(err, IsNil)
	c.Assert(node3.Start(), IsNil)
	defer node3.Stop()

	srv3 := NewHTTPServer(c, node3)
	defer srv3.Close()

	// check it catches up
	c.Log("wait for sync (node2) to indicate downstream is async (node3)")
	waitReplSync(c, node2, 3)

	db3 := connect(c, node3, "mysql")
	defer db3.Close()

	// check that data replicated successfully
	c.Log("read from async (node3)")
	waitRow(c, db3, 1)

	c.Log("assert sync (node2) lists async (node3) as downstream")
	assertDownstream(c, db2, 3)

	// Start a second async
	c.Log("start second async (node4)")
	err = node4.Reconfigure(Config(state.RoleAsync, node3, nil))
	c.Assert(err, IsNil)
	c.Assert(node4.Start(), IsNil)
	defer node4.Stop()

	srv4 := NewHTTPServer(c, node4)
	defer srv4.Close()

	// check it catches up
	c.Log("waiting for async (node3) to indicate downstream is second async (node4)")
	waitReplSync(c, node3, 4)

	db4 := connect(c, node4, "mysql")
	defer db4.Close()

	// check that data replicated successfully
	c.Log("read from second async (node4)")
	waitRow(c, db4, 1)

	c.Log("assert async (node3) lists second async (node4) as downstream")
	assertDownstream(c, db3, 4)

	// promote node2 to primary
	c.Log("stop primary (node1)")
	c.Assert(node1.Stop(), IsNil)
	c.Log("promote sync (node2) to primary")
	err = node2.Reconfigure(Config(state.RolePrimary, nil, node3))
	c.Assert(err, IsNil)
	c.Log("promote async (node3) to sync")
	err = node3.Reconfigure(Config(state.RoleSync, node2, node4))
	c.Assert(err, IsNil)

	// wait for read-write transactions to come up
	c.Log("waiting for primary (node2) to indicate downstream is sync (node3)")
	waitReplSync(c, node2, 3)
	c.Log("waiting for primary (node2) to become read/write")
	waitReadWrite(c, db2)

	// check replication of each node
	c.Log("assert primary (node2) lists sync (node3) as downstream")
	assertDownstream(c, db2, 3)
	c.Log("assert primary (node2) lists sync (node3) as downstream")
	assertDownstream(c, db3, 4)

	// write to primary and ensure data propagates to followers
	c.Log("write to primary (node2)")
	insertRow(c, db2, 2)
	db2.Close()
	c.Log("read from sync (node3)")
	waitRow(c, db3, 2)
	c.Log("read from async (node4)")
	waitRow(c, db4, 2)

	// promote node3 to primary
	c.Log("stop primary (node2)")
	c.Assert(node2.Stop(), IsNil)
	c.Log("promote sync (node3) to primary")
	err = node3.Reconfigure(Config(state.RolePrimary, nil, node4))
	c.Assert(err, IsNil)
	c.Log("promote async (node4) to sync")
	err = node4.Reconfigure(Config(state.RoleSync, node3, nil))
	c.Assert(err, IsNil)

	// check replication
	c.Log("waiting for primary (node3) to indicate downstream is sync (node4)")
	waitReplSync(c, node3, 4)
	c.Log("waiting for primary (node3) to become read/write")
	waitReadWrite(c, db3)
	c.Log("assert primary (node4) lists sync (node4) as downstream")
	assertDownstream(c, db3, 4)
	insertRow(c, db3, 3)
}

func (MariaDBSuite) TestRemoveNodes(c *C) {
	node1 := NewTestProcess(c, 1)
	node2 := NewTestProcess(c, 2)
	node3 := NewTestProcess(c, 3)
	node4 := NewTestProcess(c, 4)

	// start a chain of four nodes
	err := node1.Reconfigure(Config(state.RolePrimary, nil, node2))
	c.Assert(err, IsNil)
	c.Assert(node1.Start(), IsNil)
	defer node1.Stop()

	srv1 := NewHTTPServer(c, node1)
	defer srv1.Close()

	err = node2.Reconfigure(Config(state.RoleSync, node1, nil))
	c.Assert(err, IsNil)
	c.Assert(node2.Start(), IsNil)
	defer node2.Stop()

	srv2 := NewHTTPServer(c, node2)
	defer srv2.Close()

	err = node3.Reconfigure(Config(state.RoleAsync, node2, nil))
	c.Assert(err, IsNil)
	c.Assert(node3.Start(), IsNil)
	defer node3.Stop()

	srv3 := NewHTTPServer(c, node3)
	defer srv3.Close()

	err = node4.Reconfigure(Config(state.RoleAsync, node3, nil))
	c.Assert(err, IsNil)
	c.Assert(node4.Start(), IsNil)
	defer node4.Stop()

	srv4 := NewHTTPServer(c, node4)
	defer srv4.Close()

	// wait for cluster to come up
	node1Conn := connect(c, node1, "mysql")
	defer node1Conn.Close()
	db4 := connect(c, node4, "mysql")
	defer db4.Close()
	waitReadWrite(c, node1Conn)
	createTable(c, node1Conn)
	waitRow(c, db4, 1)
	db4.Close()

	// remove first async
	c.Assert(node3.Stop(), IsNil)
	// reconfigure second async
	err = node4.Reconfigure(Config(state.RoleAsync, node2, nil))
	c.Assert(err, IsNil)

	// run query
	db4 = connect(c, node4, "mysql")
	defer db4.Close()
	insertRow(c, node1Conn, 2)
	waitRow(c, db4, 2)
	db4.Close()

	// remove sync and promote node4 to sync
	c.Assert(node2.Stop(), IsNil)
	c.Assert(node1.Reconfigure(Config(state.RolePrimary, nil, node4)), IsNil)
	c.Assert(node4.Reconfigure(Config(state.RoleSync, node1, nil)), IsNil)

	waitReadWrite(c, node1Conn)
	insertRow(c, node1Conn, 3)
	db4 = connect(c, node4, "mysql")
	defer db4.Close()
	waitRow(c, db4, 3)
}

// newPort represents the starting port when allocating new ports.
var newPort uint32 = 7500

func NewTestProcess(c *C, n uint32) *Process {
	p := NewProcess()
	p.ID = fmt.Sprintf("node%d", n)
	p.DataDir = c.MkDir()
	p.Port = strconv.Itoa(int(atomic.AddUint32(&newPort, 2)))
	p.Password = "password"
	p.ServerID = uint32(n)
	p.OpTimeout = 30 * time.Second
	p.Logger = p.Logger.New("id", p.ID, "port", p.Port)
	return p
}

// HTTPServer is a wrapper for http.Server that provides the ability to close the listener.
type HTTPServer struct {
	*http.Server
	ln net.Listener
}

// NewHTTPServer returns a new, running HTTP server attached to a process.
func NewHTTPServer(c *C, p *Process) *HTTPServer {
	h := NewHandler()
	h.Process = p

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", MustAtoi(p.Port)+1))
	c.Assert(err, IsNil)

	s := &HTTPServer{
		Server: &http.Server{
			Handler: h,
		},
		ln: ln,
	}
	go s.Serve(ln)

	return s
}

// Close closes the server's listener.
func (s *HTTPServer) Close() error { s.ln.Close(); return nil }

// MustAtoi converts a string into an integer. Panic on error.
func MustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
