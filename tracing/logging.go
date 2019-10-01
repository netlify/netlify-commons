package tracing

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	logKey = contextKey("nf-log-key")
)

func requestLogger(r *http.Request, log logrus.FieldLogger) logrus.FieldLogger {
	if r.Header.Get(HeaderNFDebugLogging) != "" {
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)

		if entry, ok := log.(*logrus.Entry); ok {
			log = logger.WithFields(entry.Data)
		}
	}

	reqID := RequestID(r)

	log = log.WithFields(logrus.Fields{
		"request_id": reqID,
	})
	return log
}

func GetLoggerFromContext(ctx context.Context) logrus.FieldLogger {
	entry := GetFromContext(ctx)
	if entry == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return entry.FieldLogger
}

func GetLogger(r *http.Request) logrus.FieldLogger {
	entry := GetLoggerFromContext(r.Context())
	if entry == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return entry
}

// SetLogField will add the field to this log line and every one following
func SetLogField(r *http.Request, key string, value interface{}) logrus.FieldLogger {
	entry := GetTracer(r)
	if entry == nil {
		return logrus.StandardLogger().WithField(key, value)
	}
	return entry.SetLogField(key, value)
}

// SetLogFields will add the fields to this log line and every one following
func SetLogFields(r *http.Request, fields logrus.Fields) logrus.FieldLogger {
	entry := GetTracer(r)
	if entry == nil {
		return logrus.StandardLogger().WithFields(fields)
	}

	return entry.SetLogFields(fields)
}

// SetFinalField will add a field to the canonical line created at in Finish. It will add
// it to this line, but not every log line in between
func SetFinalField(r *http.Request, key string, value interface{}) logrus.FieldLogger {
	entry := GetTracer(r)
	if entry == nil {
		return logrus.StandardLogger().WithField(key, value)
	}

	entry.FinalFields[key] = value
	return entry.FieldLogger.WithField(key, value)
}
