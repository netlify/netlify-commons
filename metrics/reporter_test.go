package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReporter(t *testing.T) {
	r := newIntervalReporter(500*time.Millisecond, tl)

	r.start()
	<-time.After(time.Second)
	r.stop()

	select {
	case <-time.After(time.Second):
		assert.Fail(t, "Failed to shutdown in time")
	case <-r.shutdown:
	}

	assert.Equal(t, stopped, r.shutdownFlag)
}
