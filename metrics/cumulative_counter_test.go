package metrics

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var tl = logrus.WithField("testing", true)

func TestAddWithMultipleMTS(t *testing.T) {
	rec := new(recordingTransport)
	env := NewEnvironment(rec)

	// we aren't going to start the reporter, just manually call it
	r := newIntervalReporter(time.Second, tl)
	env.reporter = r

	g := env.NewCumulativeCounter("something")
	g.Increment(nil)
	g.IncrementN(10, DimMap{"dim1": "value1", "dim2": 4})
	g.IncrementN(20, DimMap{"dim1": "value2", "dim2": 4})

	r.report()

	// we should have sent the metric
	if assert.Len(t, rec.metrics, 3) {
		var foundT1, foundT2D1, foundT2D2 bool

		for _, m := range rec.metrics {
			assert.Equal(t, "something", m.Name)
			switch len(m.Dims) {
			case 0:
				foundT1 = true
				assert.EqualValues(t, 1, m.Value)
			case 2:
				assert.EqualValues(t, 4, m.Dims["dim2"])
				switch m.Dims["dim1"] {
				case "value1":
					foundT2D1 = true
					assert.EqualValues(t, 10, m.Value)
				case "value2":
					foundT2D2 = true
					assert.EqualValues(t, 20, m.Value)
				default:
					assert.Fail(t, "Unexpected MTS")
				}
			default:
				assert.Fail(t, "Unexpected MTS")
			}
		}

		assert.True(t, foundT1)
		assert.True(t, foundT2D1)
		assert.True(t, foundT2D2)
	}
}
