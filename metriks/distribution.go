package metriks

import (
	"fmt"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/armon/go-metrics"
)

var (
	distributionFunc         = func(name string, value float64, tags ...metrics.Label) {}
	distributionErrorHandler = func(_ error) {}
)

func initDistribution(url string, serviceName string, permTags []string) error {
	statsd, err := statsd.New(url)
	if err != nil {
		return err
	}

	distributionFunc = func(name string, value float64, tags ...metrics.Label) {
		ddtags := append(permTags, convertTags(tags)...)
		name = fmt.Sprintf("%s.%s", serviceName, strings.ReplaceAll(name, "-", "_"))

		err := statsd.Distribution(name, float64(value), ddtags, 1)
		if err != nil {
			distributionErrorHandler(err)
		}
	}

	return nil
}

// SetDistributionErrorHandler will set the global error handler. It will be invoked
// anytime that the statsd call produces an error
func SetDistributionErrorHandler(f func(error)) {
	distributionErrorHandler = f
}

// Distribution will report the value as a distribution metric to datadog.
// it only makes sense when you're using the datadog sink
func Distribution(name string, value float64, tags ...metrics.Label) {
	distributionFunc(name, value, tags...)
}

func convertTags(incoming []metrics.Label) []string {
	tags := []string{}
	for _, v := range incoming {
		tags = append(tags, fmt.Sprintf("%s:%s", v.Name, v.Value))
	}
	return tags

}
