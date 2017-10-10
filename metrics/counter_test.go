package metrics

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)
	env.ErrorHandler = failHandler(t)
	c := env.NewCounter("thingy", nil)
	c.Count(nil)

	if assert.Len(t, rec.metrics, 1) {
		assert.EqualValues(t, 1, rec.metrics[0].Value)
	}
}

func TestCountN(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)
	env.ErrorHandler = failHandler(t)

	c := env.NewCounter("thingy", nil)
	c.CountN(100, nil)

	if assert.Len(t, rec.metrics, 1) {
		assert.EqualValues(t, 100, rec.metrics[0].Value)
	}
}

func TestCountMultipleTimes(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)
	env.ErrorHandler = failHandler(t)

	ts := time.Now()

	c := env.NewCounter("thingy", nil)
	c.SetTimestamp(ts)
	c.Count(nil)

	c.SetTimestamp(time.Time{})
	c.Count(nil)

	c.Count(nil)

	if assert.Len(t, rec.metrics, 3) {
		assert.EqualValues(t, 1, rec.metrics[0].Value)
		// that we locked the time stamp
		assert.Equal(t, ts.UnixNano(), rec.metrics[0].Timestamp)

		// that we cleared the timestamp and time started again
		assert.True(t, rec.metrics[1].Timestamp > rec.metrics[0].Timestamp)
		assert.True(t, rec.metrics[2].Timestamp > rec.metrics[1].Timestamp)
	}
}
