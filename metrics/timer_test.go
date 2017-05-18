package metrics

import (
	"testing"
	"time"

	"errors"

	"github.com/stretchr/testify/assert"
)

func TestTimeIt(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	timer := env.NewTimer("something", nil)
	start := timer.Start()
	<-time.After(time.Millisecond * 100)
	stop := time.Now()
	_, err := timer.Stop(nil)
	assert.Nil(t, err)

	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		measured := start.Add(time.Duration(m.Value))
		assert.WithinDuration(t, stop, measured, time.Millisecond*10)
	}
}

func TestTimeBlock(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	wasCalled := false
	env.timeBlock("something", DimMap{"pokemon": "pikachu"}, func() {
		wasCalled = true
	})

	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		assert.Equal(t, "pikachu", m.Dims["pokemon"])
		assert.NotZero(t, m.Value)
	}
	assert.True(t, wasCalled)
}

func TestTimeBlockErr(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	wasCalled := false
	madeErr := errors.New("garbage error")
	dur, err := env.timeBlockErr("something", DimMap{"pokemon": "pikachu"}, func() error {
		wasCalled = true
		return madeErr
	})

	assert.True(t, wasCalled)
	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		assert.Equal(t, madeErr, err)
		assert.Equal(t, 2, len(m.Dims))
		assert.Equal(t, true, m.Dims["had_error"])
		assert.Equal(t, "pikachu", m.Dims["pokemon"])

		assert.NotZero(t, m.Value)
		assert.NotZero(t, dur.Nanoseconds())
	}
}
