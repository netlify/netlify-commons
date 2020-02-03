package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/netlify/netlify-commons/router"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	if ll := os.Getenv("LOG_LEVEL"); strings.ToLower(ll) == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func TestServerHealth(t *testing.T) {
	apiDef := APIFunc(func(r router.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) *router.HTTPError {
			return nil
		})
	})

	cfg := testConfig()
	svr, err := New(tl(t), "testing", cfg, apiDef)
	require.NoError(t, err)

	testSvr := httptest.NewServer(svr.svr.Handler)
	defer testSvr.Close()

	rsp, err := http.Get(testSvr.URL + cfg.HealthPath)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rsp.StatusCode)
}

func tl(t *testing.T) *logrus.Entry {
	return logrus.WithField("test", t.Name())
}

func testConfig() Config {
	return Config{
		HealthPath: "/health",
		Port:       9090,
	}
}
