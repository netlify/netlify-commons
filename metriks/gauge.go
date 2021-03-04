package metriks

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/armon/go-metrics"
)

const (
	defaultGaugeDuration = time.Second * 10
)

type PersistentGauge struct {
	name  string
	value int32
	tags  []metrics.Label

	ticker *time.Ticker
	cancel context.CancelFunc
	dur    time.Duration
}

func (g *PersistentGauge) Set(value int32) int32 {
	v := atomic.SwapInt32(&g.value, value)
	g.report(value)
	return v
}

func (g *PersistentGauge) Inc() int32 {
	v := atomic.AddInt32(&g.value, 1)
	g.report(v)
	return v
}
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

func (g *PersistentGauge) stop() {
	g.cancel()
}

func NewGauge(name string, tags ...metrics.Label) *PersistentGauge {
	return newGauge(name, tags, defaultGaugeDuration)
}

func newGauge(name string, tags []metrics.Label, dur time.Duration) *PersistentGauge {
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
