package metriks

import (
	"testing"

	"github.com/armon/go-metrics"
	"github.com/stretchr/testify/require"
)

func TestMetriksInit(t *testing.T) {
	err := InitWithSink("foo", &metrics.BlackholeSink{})
	require.NoError(t, err)

	config := Config{
		Host: "127.0.0.1",
		Port: 8125,
		Name: "dev-test",
		Tags: nil,
	}
	err = Init("foo", config)
	require.NoError(t, err)
}
