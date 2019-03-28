package tracing

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	logKey = contextKey("nf-log-key")

	HeaderNFDebugLogging = "X-NF-Debug-Logging"
)

type structuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func getEntry(r *http.Request) *structuredLoggerEntry {
	val := r.Context().Value(logKey)
	if val == nil {
		return nil
	}
	entry, ok := val.(*structuredLoggerEntry)
	if ok {
		return entry
	}
	return nil
}

func GetLogger(r *http.Request) logrus.FieldLogger {
	entry := getEntry(r)
	if entry == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return entry.Logger
}

func SetLogField(r *http.Request, key string, value interface{}) logrus.FieldLogger {
	entry := getEntry(r)
	if entry == nil {
		return logrus.StandardLogger().WithField(key, value)
	}

	entry.Logger = entry.Logger.WithField(key, value)
	return entry.Logger
}

func SetLogFields(r *http.Request, fields logrus.Fields) logrus.FieldLogger {
	entry := getEntry(r)
	if entry == nil {
		return logrus.StandardLogger().WithFields(fields)
	}

	entry.Logger = entry.Logger.WithFields(fields)
	return entry.Logger
}
