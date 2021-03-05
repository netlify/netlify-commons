package metriks

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistentGauge(t *testing.T) {
	g := NewPersistentGaugeWithDuration("some_gauge", time.Second, L("a", "b"))
	defer g.Stop()

	res := setupStatsDSink(t)

	assert.EqualValues(t, 1, g.Inc())
	assert.EqualValues(t, 0, g.Dec())
	assert.EqualValues(t, 0, g.Set(10))

	expectedValues := []string{
		// these values should be reported everytime we make a call
		"test.some_gauge.b:1.000000|g",
		"test.some_gauge.b:0.000000|g",
		"test.some_gauge.b:10.000000|g",
		// this value should be reported after an interval
		"test.some_gauge.b:10.000000|g",
	}

	for i := 0; i < 4; i++ {
		select {
		case msg := <-res:
			assert.Equal(t, expectedValues[i], msg)
		case <-time.After(time.Second * 10):
			assert.Fail(t, "failed to get a metric in time")
		}
	}
}

func TestScheduledGauge(t *testing.T) {
	var callCount int32
	cb := func() int32 {
		t.Log("here")
		callCount++
		return callCount
	}

	g := NewScheduledGaugeWithDuration("some_gauge", time.Second, cb, L("a", "b"))
	defer g.Stop()

	res := setupStatsDSink(t)

	expectedValues := []string{
		"test.some_gauge.b:1.000000|g",
		"test.some_gauge.b:2.000000|g",
	}
	for i := 0; i < 2; i++ {
		select {
		case msg := <-res:
			t.Log("got", msg)
			assert.Equal(t, expectedValues[i], msg)
		case <-time.After(time.Second * 20):
			require.Fail(t, "failed to get a metric in time")
		}
	}
}

func setupStatsDSink(t *testing.T) <-chan string {
	pc, err := net.ListenPacket("udp", ":10000")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pc.Close()) })

	sink, err := metrics.NewStatsdSink(pc.LocalAddr().String())
	require.NoError(t, err)
	t.Cleanup(sink.Shutdown)

	cfg := metrics.DefaultConfig("test")
	cfg.EnableHostname = false
	cfg.EnableRuntimeMetrics = false
	_, err = metrics.NewGlobal(cfg, sink)
	require.NoError(t, err)

	return readValues(t, pc)
}

func readValues(t *testing.T, pc net.PacketConn) <-chan string {
	res := make(chan string)
	go func() {
		for {
			buf := make([]byte, 512)
			_, _, err := pc.ReadFrom(buf)
			if err != nil {
				close(res)
				return
			}
			for _, p := range strings.Split(string(bytes.Trim(buf, "\x00")), "\n") {
				t.Log("got msg off socket", p)
				if p != "" {
					res <- p
				}
			}
		}
	}()
	return res
}
