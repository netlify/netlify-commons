package stats

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/netlify/netlify-commons/metrics"
)

// Config contains all the setting used for statistic reporting
type Config struct {
	Interval int    `mapstructure:"report_sec"`
	Prefix   string `mapstructure:"prefix"`
	Subject  string `mapstructure:"subject"`
}

var statLock sync.Mutex

// this maps the name -> {sha -> value}
// meaning that you increment the metric time series
var mtsMap = map[string]map[string]int64{}

// this is a map of sha -> map[string]string so that we can reverse it later
var dimMap = map[string]metrics.DimMap{}

// ReportStats starts reporting stats on the configured time interval.
func ReportStats(config *Config, log *logrus.Entry) {
	if config == nil || config.Interval == 0 {
		log.Debug("Skipping stats reporting because it is configured off")
		return
	}

	if config.Subject != "" {
		metrics.NewCounter("", nil)
	}

	go func() {
		log.WithFields(logrus.Fields{
			"interval":      config.Interval,
			"metric_prefix": config.Prefix,
		}).Infof("Starting to report stats every %d seconds", config.Interval)
		ticks := time.Tick(time.Duration(config.Interval) * time.Second)
		for range ticks {
			go func() {
				statLock.Lock()
				for k, series := range mtsMap {
					for sha, val := range series {
						dims := dimMap[sha]
						name := config.Prefix
						if name != "" {
							name += "."
						}
						name += k
						metrics.NewCounter(name, dims).CountN(val, nil)
						go func(n string, v int64, d metrics.DimMap) {
							d["value"] = v
							log.WithFields(logrus.Fields(d)).Infof("%s = %d", n, v)
						}(name, val, dims)
					}
				}

				statLock.Unlock()
			}()
		}
	}()
}

// Decrement reduces the metric specified by 1
func Decrement(key string, dims metrics.DimMap) {
	IncrementN(key, -1, dims)
}

// DecrementN reduces the metric by n
func DecrementN(key string, n int64, dims metrics.DimMap) {
	IncrementN(key, -n, dims)
}

// IncrementN increases the metric by n
func IncrementN(key string, n int64, dims metrics.DimMap) {
	go func() {
		statLock.Lock()
		defer statLock.Unlock()

		if dims == nil {
			dims = metrics.DimMap{}
		}
		sha := hashDims(dims)

		series, seriesExists := mtsMap[key]
		if !seriesExists {
			series = make(map[string]int64)
		}
		val, valExists := series[sha]
		if !valExists {
			val = 0
		}
		series[sha] = val + n
		mtsMap[key] = series

		if _, ok := dimMap[sha]; !ok {
			dimMap[sha] = dims
		}
	}()
}

// Increment increases the metric by 1
func Increment(key string, dims metrics.DimMap) {
	IncrementN(key, 1, dims)
}

func hashDims(dims metrics.DimMap) string {
	if dims == nil {
		return ""
	}

	data, err := json.Marshal(&dims)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", sha256.Sum256(data))
}
