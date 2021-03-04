package metriks

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/datadog"
	"github.com/pkg/errors"
)

const timerGranularity = time.Millisecond

// Init will initialize the internal metrics system with a Datadog statsd sink
func Init(serviceName string, conf Config) error {
	return InitTags(serviceName, conf, nil)
}

// InitTags behaves like Init but allows appending extra tags
func InitTags(serviceName string, conf Config, extraTags []string) error {
	if !conf.Enabled {
		return nil
	}

	sink, err := createDatadogSink(conf.StatsdAddr(), "", conf.Tags, extraTags)
	if err != nil {
		return err
	}

	return InitWithSink(serviceName, sink)
}

// InitWithURL will initialize using a URL to identify the sink type
//
// Examples:
//
//  InitWithURL("api", "datadog://187.32.21.12:8125/?hostname=foo.com&tag=env:production")
//
//  InitWithURL("api", "discard://nothing")
//
//  InitWithURL("api", "inmem://discarded/?interval=10s&duration=30s")
//
func InitWithURL(serviceName string, endpoint string) (metrics.MetricSink, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "invalid endpoint format")
	}

	hostname := u.Query().Get("hostname")
	if hostname == "" {
		h, _ := os.Hostname()
		hostname = h
	}

	var sink metrics.MetricSink
	switch u.Scheme {
	case "datadog":
		sink, err = createDatadogSink(u.Host, hostname, map[string]string{}, u.Query()["tag"])
	case "discard", "":
		sink = &metrics.BlackholeSink{}
	default:
		sink, err = metrics.NewMetricSinkFromURL(endpoint)
	}

	if err != nil {
		return nil, errors.Wrap(err, "error creating sink")
	}

	err = InitWithSink(serviceName, sink)
	return sink, err
}

// InitWithSink initializes the internal metrics system with custom sink
func InitWithSink(serviceName string, sink metrics.MetricSink) error {
	c := metrics.DefaultConfig(serviceName)
	c.EnableHostname = false
	c.EnableHostnameLabel = false
	c.EnableServiceLabel = false
	c.TimerGranularity = timerGranularity

	if _, err := metrics.NewGlobal(c, sink); err != nil {
		return err
	}
	return nil
}

func createDatadogSink(url string, name string, tags map[string]string, extraTags []string) (metrics.MetricSink, error) {
	sink, err := datadog.NewDogStatsdSink(url, name)
	if err != nil {
		return nil, err
	}

	var ddTags []string
	for k, v := range tags {
		ddTags = append(ddTags, fmt.Sprintf("%s:%s", k, v))
	}

	if extraTags != nil {
		for _, t := range extraTags {
			ddTags = append(ddTags, t)
		}
	}

	sink.SetTags(ddTags)
	if err := initDistribution(url, name, ddTags); err != nil {
		return nil, errors.Wrap(err, "failed to initialize the datadog statsd client")
	}
	return sink, nil
}

//
// Some simpler wrappers around go-metrics
//

// Labels builds a dynamic list of labels
func Labels(labels ...metrics.Label) []metrics.Label {
	return labels
}

// L returns a single label, kept short for conciseness
func L(name string, value string) metrics.Label {
	return metrics.Label{
		Name:  name,
		Value: value,
	}
}

// Inc increments a simple counter
//
// Example:
//
// metriks.Inc("publisher.errors", 1)
//
func Inc(name string, val int64, labels ...metrics.Label) {
	if len(labels) > 0 {
		metrics.IncrCounterWithLabels([]string{name}, float32(val), labels)
	} else {
		metrics.IncrCounter([]string{name}, float32(val))
	}
}

// IncLabels increments a counter with additional labels
//
// Example:
//
// metriks.IncLabels("publisher.errors", metriks.Labels(metriks.L("status_class", "4xx")), 1)
//
func IncLabels(name string, labels []metrics.Label, val int64) {
	Inc(name, val, labels...)
}

// MeasureSince records the time from start until the invocation of the function
// It is usually used with `defer` to record time of a function.
//
// Example:
//
// func getRows() ([]Row) {
//   defer metriks.MeasureSince("publisher-get-rows.time", time.Now())
//
//   query := "SELECT * FROM publisher"
//   return db.Execute(query)
// }
//
func MeasureSince(name string, start time.Time, labels ...metrics.Label) {
	if len(labels) > 0 {
		metrics.MeasureSinceWithLabels([]string{name}, start, labels)
	} else {
		metrics.MeasureSince([]string{name}, start)
	}
}

// MeasureSinceLabels is the same as MeasureSince, but with additional labels
func MeasureSinceLabels(name string, labels []metrics.Label, start time.Time) {
	MeasureSince(name, start, labels...)
}

// Sample records a float32 sample as part of a histogram. This will get histogram
// distribution metrics
//
// Example:
//
// metriks.Sample("publisher-payload-size", float32(len(payload)))
//
func Sample(name string, val float32, labels ...metrics.Label) {
	if len(labels) > 0 {
		metrics.AddSampleWithLabels([]string{name}, val, labels)
	} else {
		metrics.AddSample([]string{name}, val)
	}
}

// SampleLabels is the same as Sample but with additional labels
func SampleLabels(name string, labels []metrics.Label, val float32) {
	Sample(name, val, labels...)
}

// Gauge is used to report a single float32 value. It is most often used during a
// periodic update or timer to report the current size of a queue or how many
// connections are currently connected.
func Gauge(name string, val float32, labels ...metrics.Label) {
	if len(labels) > 0 {
		metrics.SetGaugeWithLabels([]string{name}, val, labels)
	} else {
		metrics.SetGauge([]string{name}, val)
	}
}

// GaugeLabels is the same as Gauge but with additional labels
func GaugeLabels(name string, labels []metrics.Label, val float32) {
	Gauge(name, val, labels...)
}

func (conf Config) StatsdAddr() string {
	return fmt.Sprintf("%s:%d", conf.Host, conf.Port)
}
