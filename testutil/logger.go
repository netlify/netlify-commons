package testutil

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

// Opt is a function that will modify the logger used
type Opt func(l *logrus.Logger)

// TestLogger will build a logger that is useful for debugging
// it respects levels configured by the 'LOG_LEVEL' env var.
// It takes opt functions to modify the logger used
func TestLogger(t *testing.T, opts ...Opt) (*logrus.Entry, *test.Hook) {
	l := logrus.New()
	l.SetOutput(testLogWrapper{t})
	hook := test.NewLocal(l)
	l.SetReportCaller(true)
	if ll := os.Getenv("LOG_LEVEL"); ll != "" {
		level, err := logrus.ParseLevel(ll)
		if err != nil {
			t.Logf("Error parsing the log level env var (%s), defaulting to info", ll)
			level = logrus.InfoLevel
		}
		l.SetLevel(level)
	}

	for _, o := range opts {
		o(l)
	}

	return l.WithField("test", t.Name()), hook
}

type testLogWrapper struct {
	t *testing.T
}

func (w testLogWrapper) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

// OptSetLevel will override the env var used to configre the logger
func OptSetLevel(lvl logrus.Level) Opt {
	return func(l *logrus.Logger) {
		l.SetLevel(lvl)
	}
}

// OptReportCaller will override the reporting of the calling function info
func OptReportCaller(b bool) Opt {
	return func(l *logrus.Logger) {
		l.SetReportCaller(b)
	}
}
