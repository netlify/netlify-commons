package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDimensionalOverride(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	env.AddDimension("global-val", 12)
	env.AddDimension("metric-overide", "global-level")
	env.AddDimension("instance-overide", "global-level")
	sender := env.newMetric("thing.one", CounterType, DimMap{
		"metric-val":       456,
		"metric-overide":   "metric-level",
		"instance-overide": "metric-level",
	})

	sender.send(DimMap{
		"instance-overide": "instance-level",
		"instance-val":     789,
	})

	if assert.Len(t, rec.metrics, 1) {
		m := rec.metrics[0]
		assert.EqualValues(t, 5, len(m.Dims))
		assert.EqualValues(t, 12, m.Dims["global-val"])
		assert.EqualValues(t, "metric-level", m.Dims["metric-overide"])
		assert.EqualValues(t, "instance-level", m.Dims["instance-overide"])
		assert.EqualValues(t, 456, m.Dims["metric-val"])
		assert.EqualValues(t, 789, m.Dims["instance-val"])
		assert.NotEqual(t, 0, m.Timestamp)
	}
}

func TestSettingTimestamp(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	ts := time.Now()
	m := env.newMetric("thing.one", CounterType, nil)
	m.SetTimestamp(ts)
	m.send(nil)

	m.SetTimestamp(time.Time{})
	m.send(nil)

	if assert.Len(t, rec.metrics, 2) {
		m := rec.metrics[0]
		assert.Equal(t, ts.UnixNano(), m.Timestamp)

		m = rec.metrics[1]
		assert.True(t, ts.UnixNano() < m.Timestamp)
	}
}
