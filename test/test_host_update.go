package main

import (
	"encoding/json"
	"net/http"
	"syscall"
	"time"

	ct "github.com/drycc/drycc/controller/types"
	"github.com/drycc/drycc/host/types"
	logaggc "github.com/drycc/drycc/logaggregator/client"
	logagg "github.com/drycc/drycc/logaggregator/types"
	"github.com/drycc/drycc/pkg/cluster"
	"github.com/drycc/drycc/pkg/dialer"
	"github.com/drycc/drycc/pkg/exec"
	c "github.com/drycc/go-check"
)

type HostUpdateSuite struct {
	Helper
}

var _ = c.ConcurrentSuite(&HostUpdateSuite{})

func (s *HostUpdateSuite) TestUpdateLogs(t *c.C) {
	x := s.bootCluster(t, 1)
	defer x.Destroy()

	hosts, err := x.cluster.Hosts()
	t.Assert(err, c.IsNil)
	t.Assert(hosts, c.HasLen, 1)
	hostClient := hosts[0]

	app := &ct.App{Name: "partial-logger"}
	t.Assert(x.controller.CreateApp(app), c.IsNil)

	// start partial logger job
	debug(t, "starting partial logger job")
	cmd := exec.JobUsingHost(
		hostClient,
		s.createArtifactWithClient(t, "test-apps", x.controller),
		&host.Job{
			Config: host.ContainerConfig{Args: []string{"/bin/partial-logger"}},
			Metadata: map[string]string{
				"drycc-controller.app": app.ID,
			},
		},
	)
	t.Assert(cmd.Start(), c.IsNil)
	defer cmd.Kill()

	// wait for partial line
	debug(t, "waiting for partial line")
	_, err = x.discoverd.Instances("partial-logger", 10*time.Second)
	t.Assert(err, c.IsNil)

	// update drycc-host using the same flags
	debug(t, "updating drycc-host")
	status, err := hostClient.GetStatus()
	t.Assert(err, c.IsNil)
	_, err = hostClient.UpdateWithShutdownDelay(
		"/usr/local/bin/drycc-host",
		30*time.Second,
		append([]string{"daemon"}, status.Flags...)...,
	)
	t.Assert(err, c.IsNil)

	// stream the log
	debug(t, "getting the app log")
	log, err := x.controller.GetAppLog(app.ID, &logagg.LogOpts{Follow: true})
	t.Assert(err, c.IsNil)
	defer log.Close()

	msgs := make(chan *logaggc.Message)
	go func() {
		defer close(msgs)
		dec := json.NewDecoder(log)
		for {
			var msg logaggc.Message
			if err := dec.Decode(&msg); err != nil {
				debugf(t, "error decoding message: %s", err)
				return
			}
			debugf(t, "got message: %+v", msg)
			msgs <- &msg
		}
	}()

	// give the new drycc-host daemon time to connect to the job before
	// signalling it to finish logging
	time.Sleep(5 * time.Second)

	// finish logging using a new cluster client to avoid reusing the TCP
	// connection to the host which has shut down
	debug(t, "signalling job to finish logging")
	hostClient = cluster.NewHost(
		hostClient.ID(),
		hostClient.Addr(),
		&http.Client{Transport: &http.Transport{Dial: dialer.Retry.Dial}},
		hostClient.Tags(),
	)
	t.Assert(hostClient.SignalJob(cmd.Job.ID, int(syscall.SIGUSR1)), c.IsNil)

	// check we get a single log line
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				t.Fatal("error getting log")
			}
			if msg.Stream == "stdout" {
				t.Assert(msg.Msg, c.Equals, "hello world")
				return
			}
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for log")
		}
	}
}
