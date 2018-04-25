package metrics

import (
	"sync"
	"time"
)

const (
	// TimerType is the type for timers
	TimerType MetricType = "timer"
	// CounterType is the type for timers
	CounterType MetricType = "counter"
	// GaugeType is the type for gauges
	GaugeType MetricType = "gauge"
	// CumulativeType is the type for cumulative counters
	CumulativeType MetricType = "cumulative"
)

// MetricType describes what type the metric is
type MetricType string

type RawMetric struct {
	Name      string     `json:"name"`
	Type      MetricType `json:"type"`
	Value     int64      `json:"value"`
	Dims      DimMap     `json:"dimensions"`
	Timestamp int64      `json:"timestamp"`
}

type metric struct {
	Name      string     `json:"name"`
	Type      MetricType `json:"type"`
	Dims      DimMap     `json:"dimensions"`
	Timestamp int64      `json:"timestamp"`

	value   int64
	dimlock *sync.RWMutex
	env     *Environment
}

func (m *metric) SetTimestamp(t time.Time) {
	if t.IsZero() {
		m.Timestamp = 0
	} else {
		m.Timestamp = t.UnixNano()
	}
}

// AddDimension will add this dimension with locking
func (m *metric) AddDimension(key string, value interface{}) {
	m.dimlock.Lock()
	defer m.dimlock.Unlock()
	m.Dims[key] = value
}

func (m *metric) send(instanceDims DimMap, val int64) {
	metricToSend := &RawMetric{
		Type:      m.Type,
		Value:     val,
		Name:      m.Name,
		Timestamp: m.Timestamp,
		Dims:      DimMap{},
	}

	// global
	m.env.dimlock.RLock()
	addAll(metricToSend.Dims, m.env.globalDims)
	m.env.dimlock.RUnlock()

	// metric
	m.dimlock.RLock()
	addAll(metricToSend.Dims, m.Dims)
	m.dimlock.RUnlock()

	// instance
	addAll(metricToSend.Dims, instanceDims)

	if metricToSend.Timestamp == 0 {
		metricToSend.Timestamp = time.Now().UnixNano()
	}

	m.env.send(metricToSend)
}

// DimMap is a map of dimensions
type DimMap map[string]interface{}
