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
	AuthToken string `mapstructure:"auth_token"`
	ReportSec int    `mapstructure:"report_sec"`
}

func NewSignalFXTransport(config *SFXConfig) (*SFXTransport, error) {
	sink := sfxclient.NewHTTPSink()
	sink.AuthToken = config.AuthToken

	t := &SFXTransport{
		sink,
		map[string]map[string]*sfxclient.RollingBucket{},
		time.Duration(config.ReportSec) * time.Second,
		new(sync.Mutex),
	}

	if t.reportDelay > 0 {
		scheduler := sfxclient.NewScheduler()
		scheduler.ReportingDelay(t.reportDelay)
		scheduler.Sink = sink
		scheduler.AddCallback(t)
		go scheduler.Schedule(context.Background())
	}

	return t, nil
}

type SFXTransport struct {
	sink          *sfxclient.HTTPSink
	timingBuckets map[string]map[string]*sfxclient.RollingBucket
	reportDelay   time.Duration
	statLock      *sync.Mutex
}

func (t *SFXTransport) Publish(m *metrics.RawMetric) error {
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
			return fmt.Errorf("Unsupported type for dimension '%s': '%v'", k, reflect.TypeOf(v))
		}
		p.Dimensions[k] = asStr
	}

	switch m.Type {
	case metrics.GaugeType:
		p.MetricType = datapoint.Gauge
	case metrics.CumulativeType:
		p.MetricType = datapoint.Counter
	case metrics.TimerType:
		if t.reportDelay > 0 {
			if err := t.recordTimer(m, p.Dimensions); err != nil {
				return errors.Wrap(err, "error recording timer to histogram")
			}
		}
		p.MetricType = datapoint.Count
	case metrics.CounterType:
		p.MetricType = datapoint.Count
	}

	return t.sink.AddDatapoints(context.Background(), []*datapoint.Datapoint{p})
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
	for _, dimMap := range t.timingBuckets {
		for _, bucket := range dimMap {
			pts = append(pts, bucket.Datapoints()...)
		}
	}
	return pts
}
