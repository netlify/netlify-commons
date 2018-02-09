package metering

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var encoder *json.Encoder
var meteringBuffer *bufio.Writer
var logger *logrus.Entry
var writelock sync.Mutex

func init() {
	initFromEnv()
}

func initFromEnv() {
	fname := os.Getenv("METERING_FILENAME")
	if fname != "" {
		err := SetMeteringFile(fname)
		if err != nil {
			panic(errors.Wrapf(err, "Failed to open file '%s' which was specified by the env var 'METERING_FILENAME'", fname))
		}
	} else {
		meteringBuffer = bufio.NewWriter(os.Stdout)
		encoder = json.NewEncoder(meteringBuffer)
	}

	logger = logrus.WithField("component", "metering")
}

type MeteringEvent struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

func SetMeteringLogger(log *logrus.Entry) {
	if log != nil {
		logger = log
	}
}

func SetMeteringFile(filename string) error {
	open, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	meteringBuffer = bufio.NewWriter(open)
	encoder = json.NewEncoder(meteringBuffer)
	return nil
}

func Flush() {
	writelock.Lock()
	err := meteringBuffer.Flush()
	if err != nil {
		logger.WithError(err).Warn("Failed to flush metering logger")
	}
	writelock.Unlock()
}

func RecordEvent(event string, data map[string]interface{}) {
	writelock.Lock()
	err := encoder.Encode(&MeteringEvent{
		Event:     event,
		Data:      data,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		logger.WithError(err).WithField("event", event).Warn("Failed to encode message")
	}
	writelock.Unlock()
}
