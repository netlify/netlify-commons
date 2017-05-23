package stats

import (
	"sync"
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/netlify/netlify-commons/metrics"
)

type Config struct {
	Interval int    `mapstructure:"report_sec"`
	Prefix   string `mapstructure:"prefix"`
	Subject  string `mapstructure:"subject"`
}

var statLock sync.Mutex
var stats = make(map[string]int64)

func ReportStats(config *Config, log *logrus.Entry) {
	if config == nil || config.Interval == 0 {
		log.Debug("Skipping stats reporting because it is configured off")
		return
	}

	fields := logrus.Fields{}
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
				if bs, err := json.Marshal(&stats); err == nil {
					log.WithFields(fields).Infof(string(bs))
				}

				for k, v := range stats {
					name := config.Prefix
					if name != "" {
						name += "."
					}
					name += k

					metrics.NewGauge(name, nil).Set(v, nil)
				}
				statLock.Unlock()
			}()
		}
	}()
}

func Decrement(key string) {
	IncrementN(key, -1)
}

func DecrementN(key string, n int64) {
	IncrementN(key, -1*n)
}

func IncrementN(key string, n int64) {
	go func() {
		statLock.Lock()
		defer statLock.Unlock()
		val, ok := stats[key]
		if !ok {
			val = 0
		}
		stats[key] = val + n
	}()
}

func Increment(key string) {
	IncrementN(key, 1)
}
