package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendingEvent(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	env.Namespace = "test."
	calledTracer := false
	env.EventTracer = func(event *Event) {
		calledTracer = true
	}

	e := env.NewEvent("test_event", DimMap{"thing": 1}, DimMap{"sha": "ajkldfs"})
	require.NoError(t, e.Record())

	// make sure we called through
	assert.True(t, calledTracer)
	assert.Len(t, rec.events, 1)
	recEvent := rec.events[0]
	assert.Equal(t, "test.test_event", recEvent.Name)
	assert.Len(t, recEvent.Dims, 1)
	assert.EqualValues(t, 1, recEvent.Dims["thing"])
	assert.Len(t, recEvent.Props, 1)
	assert.EqualValues(t, "ajkldfs", recEvent.Props["sha"])
}
