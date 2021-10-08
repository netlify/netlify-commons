package metriks

import (
	"bytes"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistribution(t *testing.T) {
	pc, err := net.ListenPacket("udp", "")
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
	assert.Equal(t, "testing.some_metric:12|d|#a:1,b:c", string(bytes.Trim(buf, "\x00")))
}

func TestDistributionRace(t *testing.T) {
	pc, err := net.ListenPacket("udp", "")
	require.NoError(t, err)
	defer pc.Close()

	go func() {
		for {
			buf := make([]byte, 1024)
			pc.ReadFrom(buf)
		}
	}()

	// set cap so concurrent callers of Distribution overwrite the same space
	permTags := make([]string, 1, 8)
	permTags[0] = "a:1"

	require.NoError(t, initDistribution(pc.LocalAddr().String(), "testing", permTags))
	SetDistributionErrorHandler(func(err error) {
		assert.NoError(t, err)
	})

	work := make(chan struct{})
	var wg sync.WaitGroup
	for n := 0; n < 1024; n++ {
		wg.Add(1)
		go func() {
			for range work {
				Distribution("some_metric", 12.0, L("b", "c"))
			}
			wg.Done()
		}()
	}

	for n := 0; n < 100_000; n++ {
		work <- struct{}{}
	}
	close(work)
	wg.Wait()
}
