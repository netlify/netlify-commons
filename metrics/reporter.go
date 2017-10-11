package metrics

import (
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	stopped uint32 = 0
	running uint32 = 1
)

type reporter interface {
	start()
	stop()
	register(*cumulativeCounter)
}

type noopReporter struct{}

func (*noopReporter) start()                      {}
func (*noopReporter) stop()                       {}
func (*noopReporter) register(*cumulativeCounter) {}

func newIntervalReporter(interval time.Duration, log *logrus.Entry) *intervalReporter {
	return &intervalReporter{
		counters:     []*cumulativeCounter{},
		log:          log,
		interval:     interval,
		shutdown:     make(chan bool),
		shutdownFlag: stopped,
	}
}

type intervalReporter struct {
	counters []*cumulativeCounter

	log      *logrus.Entry
	interval time.Duration

	shutdown     chan bool
	shutdownFlag uint32
}

func (r *intervalReporter) start() {
	if atomic.SwapUint32(&r.shutdownFlag, running) == running {
		// going from running -> running
		return
	}

	go func() {
		if r.log != nil {
			r.log.WithFields(logrus.Fields{
				"interval": r.interval,
			}).Infof("Starting to report stats every %s", r.interval.String())
		}

		ticks := time.Tick(r.interval)
		for {
			select {
			case <-r.shutdown:
				if r.log != nil {
					r.log.Info("Shutting down")
				}
				return
			case <-ticks:
				go r.report()
			}
		}
	}()
}

func (r *intervalReporter) stop() {
	if atomic.SwapUint32(&r.shutdownFlag, stopped) == stopped {
		// going from stopped -> stopped
		return
	}
	close(r.shutdown)
}

func (r *intervalReporter) register(c *cumulativeCounter) {
	r.counters = append(r.counters, c)
}

func (r *intervalReporter) report() {
	if len(r.counters) == 0 {
		return
	}

	results := make(map[string][]map[string]interface{})

	for _, c := range r.counters {
		for _, m := range c.series() {
			m.send(nil)
			res := map[string]interface{}{"dims": m.Dims, "value": m.Value}
			results[c.Name] = append(results[c.Name], res)
		}
	}

	if r.log != nil && len(results) > 0 {
		data, err := json.Marshal(&results)
		if err != nil {
			r.log.WithError(err).Warn("Failed to marshal stats results")
		} else {
			r.log.Infof(string(data))
		}
	}
}
