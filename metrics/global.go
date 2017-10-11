package metrics

import (
	"sync"

	"github.com/sirupsen/logrus"

	"time"
)

var globalEnv = NewEnvironment(NoopTransport{})
var initLock = sync.Mutex{}

// Init will setup the global context
func Init(trans Transport) {
	initLock.Lock()
	defer initLock.Unlock()

	globalEnv = NewEnvironment(trans)
}

func GlobalEnv() *Environment {
	initLock.Lock()
	ge := globalEnv
	initLock.Unlock()
	return ge
}

func SetErrorHandler(eh func(m *RawMetric, err error)) {
	globalEnv.ErrorHandler = eh
}

func StartReportingCumulativeCounters(interval time.Duration, log *logrus.Entry) {
	globalEnv.StartReportingCumulativeCounters(interval, log)
}

// AddDimension will let you store a dimension in the global space
func AddDimension(key string, value interface{}) {
	globalEnv.AddDimension(key, value)
}

// NewCounter creates a named counter with these dimensions
func NewCounter(name string, metricDims DimMap) Counter {
	return globalEnv.NewCounter(name, metricDims)
}

// NewCumulativeCounter creates a named cumulative counter
func NewCumulativeCounter(name string) CumulativeCounter {
	return globalEnv.NewCumulativeCounter(name)
}

// NewGauge creates a named gauge with these dimensions
func NewGauge(name string, metricDims DimMap) Gauge {
	return globalEnv.NewGauge(name, metricDims)
}

// NewTimer creates a named timer with these dimensions
func NewTimer(name string, metricDims DimMap) Timer {
	timer := globalEnv.NewTimer(name, metricDims)
	timer.Start()
	return timer
}

// TimeBlock will just time the block provided
func TimeBlock(name string, metricDims DimMap, f func()) time.Duration {
	return globalEnv.timeBlock(name, metricDims, f)
}

// TimeBlockErr will run the function and publish the time it took.
// It will add the dimension 'had_error' and return the error from the internal function
func TimeBlockErr(name string, metricDims DimMap, f func() error) (time.Duration, error) {
	return globalEnv.timeBlockErr(name, metricDims, f)
}

func Trace(tracer func(m *RawMetric)) {
	globalEnv.Tracer = tracer
}

func Namespace(val string) {
	globalEnv.Namespace = val
}

func Count(name string, metricDims DimMap) {
	globalEnv.NewCounter(name, nil).Count(metricDims)
}

func CountN(name string, val int64, metricDims DimMap) {
	globalEnv.NewCounter(name, nil).CountN(val, metricDims)
}
