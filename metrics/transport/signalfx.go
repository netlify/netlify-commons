package transport

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/sfxclient"

	"github.com/netlify/netlify-commons/metrics"
)

type SFXConfig struct {
	AuthToken            string `mapstructure:"auth_token"`
	ReportSec            int    `mapstructure:"report_sec"`
	DisableTimerCounters bool   `mapstructure:"disable_timer_counters"`
}

func NewSignalFXTransport(config *SFXConfig) (*SFXTransport, error) {
	sink := sfxclient.NewHTTPSink()
	sink.AuthToken = config.AuthToken

	if config.ReportSec <= 0 {
		return nil, fmt.Errorf("Reporting interval must be greater than zero")
	}

	t := &SFXTransport{
		sink,
		map[string]map[string]*sfxclient.RollingBucket{},
		time.Duration(config.ReportSec) * time.Second,
		new(sync.Mutex),
		config.DisableTimerCounters,
		[]*datapoint.Datapoint{},
	}

	scheduler := sfxclient.NewScheduler()
	scheduler.ReportingDelay(t.reportDelay)
	scheduler.Sink = sink
	scheduler.AddCallback(t)
	go scheduler.Schedule(context.Background())

	return t, nil
}

type SFXTransport struct {
	sink                 *sfxclient.HTTPSink
	timingBuckets        map[string]map[string]*sfxclient.RollingBucket
	reportDelay          time.Duration
	statLock             *sync.Mutex
	disableTimerCounters bool
	queue                []*datapoint.Datapoint
}

func (t *SFXTransport) Queue(m *metrics.RawMetric) error {
	p, err := t.newDatapoint(m)
	if err != nil {
		return err
	}

	if m.Type == metrics.TimerType {
		if err := t.recordTimer(m, p.Dimensions); err != nil {
			return errors.Wrap(err, "error recording timer to histogram")
		}
		if t.disableTimerCounters {
			return nil
		}
	}

	t.statLock.Lock()
	defer t.statLock.Unlock()

	t.queue = append(t.queue, p)
	return nil
}

func (t *SFXTransport) Publish(m *metrics.RawMetric) error {
	p, err := t.newDatapoint(m)
	if err != nil {
		return err
	}

	if m.Type == metrics.TimerType {
		if err := t.recordTimer(m, p.Dimensions); err != nil {
			return errors.Wrap(err, "error recording timer to histogram")
		}
		if t.disableTimerCounters {
			return nil
		}
	}

	return t.sink.AddDatapoints(context.Background(), []*datapoint.Datapoint{p})
}

func (t *SFXTransport) newDatapoint(m *metrics.RawMetric) (*datapoint.Datapoint, error) {
	p := &datapoint.Datapoint{
		Metric:     m.Name,
		Dimensions: map[string]string{},
		Value:      datapoint.NewIntValue(m.Value),
		Timestamp:  time.Unix(0, m.Timestamp),
	}

	for k, v := range m.Dims {
		var asStr string
		switch v.(type) {
		case int, int64, int32:
			asStr = fmt.Sprintf("%d", v)
		case bool:
			asStr = fmt.Sprintf("%t", v)
		case float32, float64:
			asStr = fmt.Sprintf("%f", v)
		case string:
			asStr = v.(string)
		default:
			return nil, fmt.Errorf("Unsupported type for dimension '%s': '%v'", k, reflect.TypeOf(v))
		}
		p.Dimensions[k] = asStr
	}

	switch m.Type {
	case metrics.GaugeType:
		p.MetricType = datapoint.Gauge
	case metrics.CumulativeType:
		p.MetricType = datapoint.Counter
	case metrics.TimerType:
		p.MetricType = datapoint.Count
	case metrics.CounterType:
		p.MetricType = datapoint.Count
	}

	return p, nil
}

func (t *SFXTransport) recordTimer(m *metrics.RawMetric, formattedDims map[string]string) error {
	dims := m.Dims
	if dims == nil {
		dims = metrics.DimMap{}
	}
	sha, err := metrics.HashDims(dims)
	if err != nil {
		return err
	}

	t.statLock.Lock()
	defer t.statLock.Unlock()

	dimMap, ok := t.timingBuckets[m.Name]
	if !ok {
		dimMap = make(map[string]*sfxclient.RollingBucket)
		t.timingBuckets[m.Name] = dimMap
	}
	bucket, ok := dimMap[sha]
	if !ok {
		bucket = sfxclient.NewRollingBucket(m.Name, formattedDims)
		bucket.Quantiles = []float64{.5, .9, .99}
		bucket.BucketWidth = t.reportDelay
		t.timingBuckets[m.Name][sha] = bucket
	}

	bucket.AddAt(float64(m.Value), time.Unix(0, m.Timestamp))
	return nil
}

func (t *SFXTransport) Datapoints() []*datapoint.Datapoint {
	t.statLock.Lock()
	defer t.statLock.Unlock()

	pts := []*datapoint.Datapoint{}

	for _, p := range t.queue {
		pts = append(pts, p)
	}
	t.queue = []*datapoint.Datapoint{}

	for _, dimMap := range t.timingBuckets {
		for _, bucket := range dimMap {
			dps := bucket.Datapoints()
			// do not add points if they are just .count, .sum, .sumsquare for 0 datapoints
			if len(dps) == 3 {
				continue
			}
			pts = append(pts, dps...)
		}
	}
	return pts
}
