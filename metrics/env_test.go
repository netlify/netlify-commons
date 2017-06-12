package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMetric(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	// create the metric
	sender := env.newMetric("something", CounterType, nil)
	sender.Value = 123
	err := sender.send(nil)
	assert.Nil(t, err)

	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		assert.Equal(t, "something", m.Name)
		assert.EqualValues(t, m.Value, 123)
		assert.Equal(t, m.Type, CounterType)
		assert.NotNil(t, m.Dims)
		assert.Len(t, m.Dims, 0)

		// validate counts
		checkCounters(t, 1, 0, 0, env)
	}
}

func TestSendWithTracer(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	called := false
	env.Tracer = func(m *RawMetric) {
		called = true
		if assert.NotNil(t, m) {
			assert.Equal(t, "something", m.Name)
			assert.EqualValues(t, m.Value, 123)
			assert.Equal(t, m.Type, CounterType)
			assert.NotNil(t, m.Dims)
			assert.Len(t, m.Dims, 0)
		}
	}

	// create the metric
	sender := env.newMetric("something", CounterType, nil)
	sender.Value = 123
	err := sender.send(nil)
	require.Nil(t, err)

	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		assert.Equal(t, "something", m.Name)
		assert.EqualValues(t, m.Value, 123)
		assert.Equal(t, m.Type, CounterType)
		assert.NotNil(t, m.Dims)
		assert.Len(t, m.Dims, 0)
	}
	assert.True(t, called)
	checkCounters(t, 1, 0, 0, env)
}

func TestSeparateEnv(t *testing.T) {
	rec1 := new(recordingTransport)
	rec2 := new(recordingTransport)

	e1 := NewEnvironment(rec1)
	e2 := NewEnvironment(rec2)

	require.NoError(t, e1.NewCounter("c1", nil).Count(nil))
	require.NoError(t, e2.NewCounter("c2", nil).Count(nil))

	assert.Len(t, rec1.metrics, 1)
	assert.Len(t, rec2.metrics, 1)

	assert.Equal(t, "c1", rec1.metrics[0].Name)
	assert.Equal(t, "c2", rec2.metrics[0].Name)
}

func TestNamespace(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	require.NoError(t, env.NewCounter("c1", nil).Count(nil))
	env.Namespace = "marp."
	require.NoError(t, env.NewCounter("c2", nil).Count(nil))

	assert.Len(t, rec.metrics, 2)
	assert.Equal(t, "c1", rec.metrics[0].Name)
	assert.Equal(t, "marp.c2", rec.metrics[1].Name)
}

func checkCounters(t *testing.T, counters, timers, gauges int, env *Environment) {
	assert.EqualValues(t, counters, env.countersSent)
	assert.EqualValues(t, timers, env.timersSent)
	assert.EqualValues(t, gauges, env.gaugesSent)
}

type recordingTransport struct {
	metrics []*RawMetric
	events  []*Event
}

func (t *recordingTransport) Publish(m *RawMetric) error {
	t.metrics = append(t.metrics, m)
	return nil
}
func (t *recordingTransport) PublishEvent(e *Event) error {
	t.events = append(t.events, e)
	return nil
}
