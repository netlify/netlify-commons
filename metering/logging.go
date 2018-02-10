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

func init() {
	initFromEnv()
}

var global *MeteringLog
var errorLogger *logrus.Entry = logrus.WithField("component", "metering_errors")

type MeteringLog struct {
	writelock *sync.Mutex
	encoder   *json.Encoder
	buffer    *bufio.Writer
}

type MeteringEvent struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

func initFromEnv() {
	var err error
	fname := os.Getenv("METERING_FILENAME")
	if fname != "" {
		global, err = NewMeteringLog(fname)
		if err != nil {
			panic(errors.Wrapf(err, "Failed to open file '%s' which was specified by the env var 'METERING_FILENAME'", fname))
		}
	} else {
		global = DefaultMeteringLog()
	}
}

func Global() *MeteringLog {
	return global
}

func NewMeteringLog(filename string) (*MeteringLog, error) {
	ml := &MeteringLog{
		writelock: new(sync.Mutex),
	}
	err := ml.setOutputFile(filename)
	if err != nil {
		return nil, err
	}

	return ml, nil
}

func DefaultMeteringLog() *MeteringLog {
	return &MeteringLog{
		writelock: new(sync.Mutex),
		buffer:    bufio.NewWriter(os.Stdout),
		encoder:   json.NewEncoder(os.Stdout),
	}
}
func SetErrorLogger(log *logrus.Entry) {
	if log != nil {
		errorLogger = log
	}
}

func SetMeteringFile(filename string) error { return global.setOutputFile(filename) }
func (ml *MeteringLog) setOutputFile(filename string) error {
	open, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	buf := bufio.NewWriter(open)
	ml.buffer = buf
	ml.encoder = json.NewEncoder(buf)
	return nil
}

func Flush() {
	if err := global.Flush(); err != nil {
		errorLogger.WithError(err).Warn("Failed to flush buffer")
	}
}
func (ml *MeteringLog) Flush() error {
	ml.writelock.Lock()
	defer ml.writelock.Unlock()
	return ml.buffer.Flush()
}

func RecordEvent(event string, data map[string]interface{}) {
	if err := global.RecordEvent(event, data); err != nil {
		errorLogger.WithError(err).WithField("event", event).Warn("Failed to encode message")
	}
}
func (ml *MeteringLog) RecordEvent(event string, data map[string]interface{}) error {
	ml.writelock.Lock()
	defer ml.writelock.Unlock()

	return ml.encoder.Encode(&MeteringEvent{
		Event:     event,
		Data:      data,
		Timestamp: time.Now().UnixNano(),
	})
}
