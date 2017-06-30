package stats

import (
	"crypto/sha256"
	"fmt"
	"strings"
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
}

var statLock sync.Mutex

// this maps the name -> {sha -> value}
// meaning that you increment the metric time series
var mtsMap = map[string]map[string]int64{}

// this is a map of sha -> map[string]string so that we can reverse it later
var dimMap = map[string]metrics.DimMap{}

// ReportStats starts reporting stats on the configured time interval.
func ReportStats(config *Config, log *logrus.Entry) chan bool {
	shutdown := make(chan bool, 1)
	if config == nil || config.Interval == 0 {
		if log != nil {
			log.Debug("Skipping stats reporting because it is configured off")
		}
		return shutdown
	}

	go func() {
		if log != nil {
			log.WithFields(logrus.Fields{
				"interval":      config.Interval,
				"metric_prefix": config.Prefix,
			}).Infof("Starting to report stats every %d seconds", config.Interval)
		}

		ticks := time.Tick(time.Duration(config.Interval) * time.Second)
		for {
			select {
			case <-shutdown:
				if log != nil {
					log.Info("Shutting down")
				}
			case <-ticks:
				go report(log, config.Prefix)
			}
		}
	}()

	return shutdown
}

func report(log *logrus.Entry, prefix string) {
	results := make(map[string][]map[string]interface{})

	statLock.Lock()
	for k, series := range mtsMap {
		for sha, val := range series {
			dims := dimMap[sha]
			name := prefix
			if name != "" && !strings.HasSuffix(name, ".") {
				name += "."
			}
			name += k
			metrics.CountN(name, val, dims)

			resMap := map[string]interface{}{
				"value": val,
				"dims":  dims,
			}
			results[name] = append(results[name], resMap)
		}
	}
	statLock.Unlock()

	if log != nil {
		if len(results) > 0 {
			data, err := json.Marshal(&results)
			if err != nil {
				log.WithError(err).Warn("Failed to marshal stats results")
			} else {
				log.Infof(string(data))
			}
		}
	}
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
