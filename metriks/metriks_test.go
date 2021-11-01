package metriks

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/datadog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/nettest"
)

func TestMetriksInit(t *testing.T) {
	err := InitWithSink("foo", &metrics.BlackholeSink{})
	require.NoError(t, err)

	config := Config{
		Host: "127.0.0.1",
		Port: 8125,
		Tags: nil,
	}
	err = Init("foo", config)
	require.NoError(t, err)
}

func TestDatadogSink(t *testing.T) {
	l, err := nettest.NewLocalPacketListener("udp")
	require.NoError(t, err)
	defer l.Close()

	endpoint := fmt.Sprintf("datadog://%s?namespace=edge_state&tag=app:edge-state&tag=env:test",
		l.LocalAddr().String())
	sink, err := InitWithURL("test", endpoint)
	require.NoError(t, err)
	require.IsType(t, &datadog.DogStatsdSink{}, sink)

	Inc("test_counter", 1)

	expectedMsg := "test.test_counter:1|c|#app:edge-state,env:test,service:test"

	var readBytes int
	buf := make([]byte, 512)
	readBytes, _, err = l.ReadFrom(buf)
	require.NoError(t, err)
	require.True(t, readBytes > 0)

	require.True(t, bytes.Equal(buf[0:readBytes], []byte(expectedMsg)))
}

func TestDiscardSink(t *testing.T) {
	sink, err := InitWithURL("test", "discard://")
	require.NoError(t, err)
	require.IsType(t, &metrics.BlackholeSink{}, sink)

	sink, err = InitWithURL("test", "")
	require.NoError(t, err)
	require.IsType(t, &metrics.BlackholeSink{}, sink)
}

func TestInMemorySink(t *testing.T) {
	sink, err := InitWithURL("test", "inmem://?interval=1s&retain=2s")
	require.NoError(t, err)
	require.IsType(t, &metrics.InmemSink{}, sink)

	Inc("test_counter", 1)

	met := sink.(*metrics.InmemSink).Data()
	require.Len(t, met, 1)
	require.Len(t, met[0].Counters, 1)
	require.Contains(t, met[0].Counters, "test.test_counter")
	require.Equal(t, "test.test_counter", met[0].Counters["test.test_counter"].Name)
	require.Equal(t, 1, met[0].Counters["test.test_counter"].Count)
}

func TestIncWithLabels(t *testing.T) {
	sink, err := InitWithURL("test", "inmem://?interval=1s&retain=2s")
	require.NoError(t, err)
	require.IsType(t, &metrics.InmemSink{}, sink)

	Inc("test_counter", 1, L("tag", "value"), L("tag2", "value2"))
	met := sink.(*metrics.InmemSink).Data()
	require.Len(t, met, 1)

	require.Len(t, met[0].Counters, 1)
	var incr metrics.SampledValue
	for _, v := range met[0].Counters {
		incr = v
		break
	}

	assert.Len(t, incr.Labels, 2)
	for _, l := range incr.Labels {
		switch l.Name {
		case "tag":
			assert.Equal(t, "value", l.Value)
		case "tag2":
			assert.Equal(t, "value2", l.Value)
		default:
			assert.Fail(t, "unexpected label value")
		}
	}
}
