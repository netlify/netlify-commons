package metriks

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/armon/go-metrics"
	"github.com/netlify/netlify-commons/util"
)

const (
	defaultGaugeDuration = time.Second * 5
)

// PersistentGauge will report on an interval the value to the metrics collector.
//
type PersistentGauge struct {
	name  string
	value int32
	tags  []metrics.Label

	ticker *time.Ticker
	cancel context.CancelFunc
	dur    time.Duration
	donec  chan struct{}
}

// Set will replace the value with a new one, it returns the old value
func (g *PersistentGauge) Set(value int32) int32 {
	return atomic.SwapInt32(&g.value, value)
}

// Inc will +1 to the current value and return the new value
func (g *PersistentGauge) Inc() int32 {
	return atomic.AddInt32(&g.value, 1)
}

// Dec will -1 to the current value and return the new value
func (g *PersistentGauge) Dec() int32 {
	return atomic.AddInt32(&g.value, -1)
}

func (g *PersistentGauge) report(v int32) {
	Gauge(g.name, float32(v), g.tags...)
	g.ticker.Reset(g.dur)
}

func (g *PersistentGauge) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(g.donec)
			return
		case <-g.ticker.C:
			g.report(atomic.LoadInt32(&g.value))
		}
	}
}

// Stop will make the gauge stop reporting. Any calls to Inc/Set/Dec will still report
// to the metrics collector
func (g *PersistentGauge) Stop() {
	g.cancel()
	<-g.donec
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
		donec:  make(chan struct{}),
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
		}),
	}
	g.Start()
	return g
}
