package stats

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/netlify/netlify-commons/metrics"
	"github.com/stretchr/testify/assert"
)

func TestReporting(t *testing.T) {
	var discovered int
	var foundT1, foundT2V1, foundT2V2 bool

	metrics.Trace(func(m *metrics.RawMetric) {
		discovered++
		switch m.Name {
		case "test.test1":
			foundT1 = true
			assert.EqualValues(t, 1, m.Value)
		case "test.test2":
			assert.Len(t, m.Dims, 2)
			assert.EqualValues(t, 4, m.Dims["dim2"])
			switch v := m.Dims["dim1"]; v {
			case "value1":
				foundT2V1 = true
				assert.EqualValues(t, 2, m.Value)
			case "value2":
				foundT2V2 = true
				assert.EqualValues(t, 10, m.Value)
			default:
				assert.Fail(t, "Unexpected value for dim1: %v", v)
			}
		default:
			assert.Fail(t, "Unexpected metric: "+m.Name)
		}
	})

	Increment("test1", nil)
	IncrementN("test2", 2, metrics.DimMap{"dim1": "value1", "dim2": 4})
	IncrementN("test2", 10, metrics.DimMap{"dim1": "value2", "dim2": 4})

	config := &Config{Interval: 1, Prefix: "test"}

	shutdown := ReportStats(config, logrus.WithField("test", true))

	<-time.After(time.Second + time.Millisecond*500)
	shutdown <- true

	assert.True(t, foundT1)
	assert.True(t, foundT2V1)
	assert.True(t, foundT2V2)
	assert.Equal(t, 3, discovered)
}
