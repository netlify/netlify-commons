package metriks

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/armon/go-metrics"
	"github.com/netlify/netlify-commons/util"
)

const (
	defaultGaugeDuration = time.Second * 10
)

// PersistentGauge will report on an interval the value to the metrics collector.
// Every call to the methods to modify the value immediately report, but if we
// don't have a change inside the window (default 10s) after the last report
// we will report the current value.
type PersistentGauge struct {
	name  string
	value int32
	tags  []metrics.Label

	ticker *time.Ticker
	cancel context.CancelFunc
	dur    time.Duration
}

// Set will replace the value with a new one, it returns the old value
func (g *PersistentGauge) Set(value int32) int32 {
	v := atomic.SwapInt32(&g.value, value)
	g.report(value)
	return v
}

// Inc will +1 to the current value and return the new value
func (g *PersistentGauge) Inc() int32 {
	v := atomic.AddInt32(&g.value, 1)
	g.report(v)
	return v
}

// Dec will -1 to the current value and return the new value
func (g *PersistentGauge) Dec() int32 {
	v := atomic.AddInt32(&g.value, -1)
	g.report(v)
	return v
}

func (g *PersistentGauge) report(v int32) {
	Gauge(g.name, float32(v), g.tags...)
	g.ticker.Reset(g.dur)
}

func (g *PersistentGauge) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-g.ticker.C:
			g.report(g.value)
		}
	}
}

// Stop will make the gauge stop reporting. Any calls to Inc/Set/Dec will still report
// to the metrics collector
func (g *PersistentGauge) Stop() {
	g.cancel()
}

// NewPersistentGauge will create and start a PersistentGauge that reports the current value every 10s
func NewPersistentGauge(name string, tags ...metrics.Label) *PersistentGauge {
	return NewPersistentGaugeWithDuration(name, defaultGaugeDuration, tags...)
}

// NewPersistentGaugeWithDuration will create and start a PersistentGauge that reports the current value every period
func NewPersistentGaugeWithDuration(name string, dur time.Duration, tags ...metrics.Label) *PersistentGauge {
	ctx, cancel := context.WithCancel(context.Background())
	g := PersistentGauge{
		name:   name,
		tags:   tags,
		ticker: time.NewTicker(dur),
		cancel: cancel,
		dur:    dur,
	}
	go g.start(ctx)
	return &g
}

// ScheduledGauge will call the provided method after a duration
// it will then report that value to the metrics system
type ScheduledGauge struct {
	util.ScheduledExecutor
}

// NewScheduledGauge will create an start a ScheduledGauge that reports the value every 10s
func NewScheduledGauge(name string, cb func() int32, tags ...metrics.Label) ScheduledGauge {
	return NewScheduledGaugeWithDuration(name, defaultGaugeDuration, cb, tags...)
}

// NewScheduledGaugeWithDuration will create an start a ScheduledGauge that reports the value every period
func NewScheduledGaugeWithDuration(name string, dur time.Duration, cb func() int32, tags ...metrics.Label) ScheduledGauge {
	g := ScheduledGauge{
		ScheduledExecutor: util.NewScheduledExecutor(dur, func() {
			v := cb()
			Gauge(name, float32(v), tags...)
			fmt.Println("sent the value?")
		}),
	}
	g.Start()
	return g
}
