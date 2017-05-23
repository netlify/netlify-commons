package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncrement(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	g := env.NewGauge("something", nil)
	g.Increment(nil)

	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		assert.EqualValues(t, 1, m.Value)
	}
}

func TestDecrement(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	g := env.NewGauge("something", nil)
	g.Decrement(nil)

	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		assert.EqualValues(t, -1, m.Value)
	}
}

func TestSet(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	g := env.NewGauge("something", nil)
	g.Set(123, nil)

	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		assert.EqualValues(t, 123, m.Value)
	}
}
