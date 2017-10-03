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
	RawMetric

	dimlock *sync.Mutex
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
func (m *metric) AddDimension(key string, value interface{}) *metric {
	m.dimlock.Lock()
	defer m.dimlock.Unlock()
	m.Dims[key] = value
	return m
}

func (m *metric) send(instanceDims DimMap) {
	metricToSend := &RawMetric{
		Type:      m.Type,
		Value:     m.Value,
		Name:      m.Name,
		Timestamp: m.Timestamp,
		Dims:      DimMap{},
	}

	// global
	m.env.dimlock.Lock()
	addAll(metricToSend.Dims, m.env.globalDims)
	m.env.dimlock.Unlock()

	// metric
	m.dimlock.Lock()
	addAll(metricToSend.Dims, m.Dims)
	m.dimlock.Unlock()

	// instance
	addAll(metricToSend.Dims, instanceDims)

	if metricToSend.Timestamp == 0 {
		metricToSend.Timestamp = time.Now().UnixNano()
	}

	m.env.send(metricToSend)
}

// DimMap is a map of dimensions
type DimMap map[string]interface{}
