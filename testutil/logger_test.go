package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerWrapper(t *testing.T) {
	tl, h := TestLogger(t)
	tl.Info("this is a test")
	assert.Len(t, h.Entries, 1)
	assert.Equal(t, "this is a test", h.Entries[0].Message)
	assert.Equal(t, "TestLoggerWrapper", h.Entries[0].Data["test"])
}

func TestLogWrapperLevel(t *testing.T) {
	testCases := []struct {
		desc     string
		expected []string
	}{
		{desc: "DEBUG", expected: []string{"debug", "info", "warn", "error"}},
		{desc: "NONSENSE", expected: []string{"info", "warn", "error"}},
		{desc: "INFO", expected: []string{"info", "warn", "error"}},
		{desc: "ERROR", expected: []string{"error"}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tC.desc)
			tl, h := TestLogger(t)
			tl.Debug("debug")
			tl.Info("info")
			tl.Warn("warn")
			tl.Error("error")

			logged := []string{}
			for _, entry := range h.Entries {
				logged = append(logged, entry.Message)
			}
			assert.Equal(t, tC.expected, logged)
		})
	}
}
