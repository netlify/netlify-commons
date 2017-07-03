package transport

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/sfxclient"

	"github.com/netlify/netlify-commons/metrics"
)

type SFXConfig struct {
	AuthToken string `mapstructure:"auth_token"`
}

func NewSignalFXTransport(config *SFXConfig) (*SFXTransport, error) {
	sink := sfxclient.NewHTTPSink()
	sink.AuthToken = config.AuthToken

	return &SFXTransport{sink}, nil
}

type SFXTransport struct {
	sink *sfxclient.HTTPSink
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
	case metrics.TimerType:
		fallthrough
	case metrics.CumulativeType:
		p.MetricType = datapoint.Counter
	case metrics.CounterType:
		p.MetricType = datapoint.Count
	}

	return t.sink.AddDatapoints(context.Background(), []*datapoint.Datapoint{p})
}
