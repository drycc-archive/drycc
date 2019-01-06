package utils

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	logagg "github.com/drycc/drycc/logaggregator/types"
	"github.com/drycc/drycc/pkg/syslog/rfc5424"
)

func ParseMessage(data []byte) (*rfc5424.Message, *HostCursor, error) {
	msg, err := rfc5424.Parse(data)
	if err != nil {
		return nil, nil, err
	}
	c, err := ParseHostCursor(msg)
	return msg, c, err
}

func ParseHostCursor(msg *rfc5424.Message) (*HostCursor, error) {
	sd, err := rfc5424.ParseStructuredData(msg.StructuredData)
	if err != nil {
		return nil, err
	}
	if sd == nil || !bytes.Equal(sd.ID, []byte("drycc")) || len(sd.Params) == 0 {
		return nil, errors.New("missing structured data")
	}
	var c *HostCursor
	for _, p := range sd.Params {
		if !bytes.Equal(p.Name, []byte("seq")) {
			continue
		}
		seq, err := strconv.ParseUint(string(p.Value), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing seq: %s", err)
		}
		c = &HostCursor{msg.Timestamp, seq}
		break
	}
	if c == nil {
		return nil, errors.New("missing seq structured data")
	}
	return c, nil
}

type HostCursor struct {
	Time time.Time `json:"time"`
	Seq  uint64    `json:"seq"`
}

func (c HostCursor) After(other HostCursor) bool {
	return c.Time.After(other.Time) || (c.Time.Equal(other.Time) && c.Seq > other.Seq)
}

func StreamType(msg *rfc5424.Message) logagg.StreamType {
	switch logagg.MsgID(msg.MsgID) {
	case logagg.MsgIDStdout:
		return logagg.StreamTypeStdout
	case logagg.MsgIDStderr:
		return logagg.StreamTypeStderr
	case logagg.MsgIDInit:
		return logagg.StreamTypeInit
	default:
		return logagg.StreamTypeUnknown
	}
}
