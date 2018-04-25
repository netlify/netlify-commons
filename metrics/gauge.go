package metrics

import (
	"sync/atomic"
	"time"
)

// Gauge keeps a running measure of the value at that moment
type Gauge interface {
	Increment(DimMap)
	Decrement(DimMap)
	Set(int64, DimMap)
	SetTimestamp(time.Time)
}

type gauge struct {
	metric
}

func (e *Environment) NewGauge(name string, metricDims DimMap) Gauge {
	m := e.newMetric(name, GaugeType, metricDims)
	return &gauge{
		metric: *m,
	}
}

func (m *gauge) Increment(instanceDims DimMap) {
	val := atomic.AddInt64(&m.value, 1)
	m.send(instanceDims, val)
}

func (m *gauge) Decrement(instanceDims DimMap) {
	val := atomic.AddInt64(&m.value, -1)
	m.send(instanceDims, val)
}

func (m *gauge) Set(val int64, instanceDims DimMap) {
	atomic.SwapInt64(&m.value, val)
	m.send(instanceDims, val)
}
