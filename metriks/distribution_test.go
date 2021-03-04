package metriks

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistribution(t *testing.T) {
	pc, err := net.ListenPacket("udp", ":10000")
	require.NoError(t, err)
	defer pc.Close()

	msgs := make(chan []byte)
	go func() {
		buf := make([]byte, 1024)
		_, _, err := pc.ReadFrom(buf)
		require.NoError(t, err)
		msgs <- buf
		close(msgs)
	}()

	require.NoError(t, initDistribution(pc.LocalAddr().String(), "testing", []string{"a:1"}))
	SetDistributionErrorHandler(func(err error) {
		assert.NoError(t, err)
	})

	Distribution("some_metric", 12.0, L("b", "c"))

	buf := <-msgs
	assert.Equal(t, "some_metric.some_metric:12|d|#a:1,b:c", string(bytes.Trim(buf, "\x00")))
}
