package transport

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/nats-io/gnatsd/test"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"

	"github.com/netlify/netlify-commons/metrics"
)

var nc *nats.Conn

func TestMain(m *testing.M) {
	server := test.RunDefaultServer()
	defer server.Shutdown()

	conn, err := nats.Connect("nats://" + server.Addr().String())
	if err != nil {
		panic(err)
	}

	nc = conn
	os.Exit(m.Run())
}

func TestCanWriteAndRead(t *testing.T) {
	nt := NewNatsTransport("test-subject", nc)

	rx := make(chan struct{})
	var msg *nats.Msg
	nc.Subscribe("test-subject", func(m *nats.Msg) {
		msg = m
		close(rx)
	})
	ts := time.Now()
	in := &metrics.RawMetric{
		Name:  "some-metric",
		Type:  metrics.CounterType,
		Value: 123,
		Dims: metrics.DimMap{
			"somenum":    1,
			"somebool":   true,
			"somestring": "string",
		},
		Timestamp: ts.UnixNano(),
	}
	nt.Publish(in)

	select {
	case <-rx:
	case <-time.After(time.Second):
		assert.Fail(t, "Failed to get message in time")
	}

	// validate the message itself
	out := new(metrics.RawMetric)
	assert.NoError(t, json.Unmarshal(msg.Data, out))

	assert.EqualValues(t, in.Name, out.Name)
	assert.EqualValues(t, in.Value, out.Value)
	assert.EqualValues(t, in.Timestamp, out.Timestamp)
	assert.EqualValues(t, in.Type, out.Type)
	assert.Len(t, out.Dims, 3)
	assert.EqualValues(t, 1, out.Dims["somenum"])
	assert.EqualValues(t, true, out.Dims["somebool"])
	assert.EqualValues(t, "string", out.Dims["somestring"])
}
