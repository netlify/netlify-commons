package stats

import (
	"sync"
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/rybit/nats_metrics"
)

type Config struct {
	Interval int    `mapstructure:"report_sec"`
	Prefix   string `mpastructure:"prefix"`
}

var statLock sync.Mutex
var stats = make(map[string]int64)

func ReportStats(config *Config, log *logrus.Entry) {
	if config == nil || config.Interval == 0 {
		log.Debug("Skipping stats reporting because it is configured off")
		return
	}

	fields := logrus.Fields{}

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

func Increment(key string) {
	go func() {
		statLock.Lock()
		defer statLock.Unlock()
		val, ok := stats[key]
		if !ok {
			val = 0
		}
		stats[key] = val + 1
	}()
}
