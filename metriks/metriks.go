package metriks

import (
	"fmt"
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/datadog"
)

const timerGranularity = time.Millisecond

// Init will initialize the internal metrics system and build a Datadog statsd sink
func Init(serviceName string, conf Config) error {
	sink, err := datadog.NewDogStatsdSink(statsdAddr(conf), conf.Name)
	if err != nil {
		return err
	}

	var tags []string
	for k, v := range conf.Tags {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	sink.SetTags(tags)

	c := metrics.DefaultConfig(serviceName)
	c.TimerGranularity = timerGranularity

	if _, err := metrics.NewGlobal(c, sink); err != nil {
		return err
	}
	return nil
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
func Inc(name string, val int64) {
	metrics.IncrCounter([]string{name}, float32(val))
}

// IncLabels increments a counter with additional labels
//
// Example:
//
// metriks.IncLabels("publisher.errors", metriks.Labels(metriks.L("status_class", "4xx")), 1)
//
func IncLabels(name string, labels []metrics.Label, val int64) {
	metrics.IncrCounterWithLabels([]string{name}, float32(val), labels)
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
func MeasureSince(name string, start time.Time) {
	metrics.MeasureSince([]string{name}, start)
}

// MeasureSinceLabels is the same as MeasureSince, but with additional labels
func MeasureSinceLabels(name string, labels []metrics.Label, start time.Time) {
	metrics.MeasureSinceWithLabels([]string{name}, start, labels)
}

// Sample records a float32 sample as part of a histogram. This will get histogram
// distribution metrics
//
// Example:
//
// metriks.Sample("publisher-payload-size", float32(len(payload)))
//
func Sample(name string, val float32) {
	metrics.AddSample([]string{name}, val)
}

// SampleLabels is the same as Sample but with additional labels
func SampleLabels(name string, labels []metrics.Label, val float32) {
	metrics.AddSampleWithLabels([]string{name}, val, labels)
}

// Gauge is used to report a single float32 value. It is most often used during a
// periodic update or timer to report the current size of a queue or how many
// connections are currently connected.
func Gauge(name string, val float32) {
	metrics.SetGauge([]string{name}, val)
}

// GaugeLabels is the same as Gauge but with additional labels
func GaugeLabels(name string, labels []metrics.Label, val float32) {
	metrics.SetGaugeWithLabels([]string{name}, val, labels)
}

func statsdAddr(conf Config) string {
	return fmt.Sprintf("%s:%d", conf.Host, conf.Port)
}
