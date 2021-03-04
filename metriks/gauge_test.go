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

func TestGauge(t *testing.T) {
	pc, err := net.ListenPacket("udp", ":10000")
	require.NoError(t, err)
	defer pc.Close()

	g := newGauge("some_gauge", []metrics.Label{L("a", "b")}, time.Second)
	defer g.stop()

	var actualVals []string
	done := make(chan struct{})
	go func() {
		for len(actualVals) != 4 {
			buf := make([]byte, 1024)
			_, _, err := pc.ReadFrom(buf)
			require.NoError(t, err)
			for _, p := range strings.Split(string(bytes.Trim(buf, "\x00")), "\n") {
				if p != "" {
					actualVals = append(actualVals, p)
				}
			}
		}
		close(done)
	}()

	sink, err := metrics.NewStatsdSink(pc.LocalAddr().String())
	require.NoError(t, err)
	defer sink.Shutdown()

	cfg := metrics.DefaultConfig("test")
	cfg.EnableHostname = false
	cfg.EnableRuntimeMetrics = false
	metrics.NewGlobal(cfg, sink)

	assert.EqualValues(t, 1, g.Inc())
	assert.EqualValues(t, 0, g.Dec())
	assert.EqualValues(t, 0, g.Set(10))

	select {
	case <-done:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "failed to get a metric in time")
	}

	expectedValues := []string{
		// these values should be reported everytime we make a call
		"test.some_gauge.b:1.000000|g",
		"test.some_gauge.b:0.000000|g",
		"test.some_gauge.b:10.000000|g",
		// this value should be reported after an interval
		"test.some_gauge.b:10.000000|g",
	}
	assert.Equal(t, expectedValues, actualVals)
}
