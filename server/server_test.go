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
	apiDef := APIFunc(
		func(r router.Router) error {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) error {
				return nil
			})
			return nil
		},
		func() {
		},
		APIInfo{
			Name:    t.Name(),
			Version: "",
		},
	)

	cfg := testConfig()
	svr, err := New(tl(t), cfg, apiDef)
	require.NoError(t, err)

	testSvr := httptest.NewServer(svr.svr.Handler)
	defer testSvr.Close()

	rsp, err := http.Get(testSvr.URL + cfg.HealthPath)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rsp.StatusCode)
}

func TestServerVersioning(t *testing.T) {
	apiDef := APIFunc(
		func(r router.Router) error {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) error {
				return nil
			})
			return nil
		},
		func() {
		},
		APIInfo{
			Name:    "testing",
			Version: "",
		},
	)
	cfg := testConfig()
	t.Run("with-no-version", func(t *testing.T) {
		svr, err := New(tl(t), cfg, apiDef)
		require.NoError(t, err)
		testSvr := httptest.NewServer(svr.svr.Handler)
		defer testSvr.Close()
		rsp, err := http.Get(testSvr.URL + cfg.HealthPath)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rsp.StatusCode)
		assert.Equal(t, "unknown", rsp.Header.Get("X-Nf-Testing-Version"))
	})

	apiDef = APIFunc(
		func(r router.Router) error {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) error {
				return nil
			})
			return nil
		},
		func() {
		},
		APIInfo{
			Name:    "testing",
			Version: "123",
		},
	)

	t.Run("with-version", func(t *testing.T) {
		svr, err := New(tl(t), cfg, apiDef)
		require.NoError(t, err)
		testSvr := httptest.NewServer(svr.svr.Handler)
		defer testSvr.Close()
		rsp, err := http.Get(testSvr.URL + cfg.HealthPath)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rsp.StatusCode)
		assert.Equal(t, "123", rsp.Header.Get("X-Nf-testing-version"))
	})

}

type testAPICustomHealth struct{}

func (a *testAPICustomHealth) Start(r router.Router) error {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})
	return nil
}

func (a *testAPICustomHealth) Stop() {}

func (a *testAPICustomHealth) Healthy(w http.ResponseWriter, r *http.Request) error {
	return router.InternalServerError("healthcheck failed")
}
func (a *testAPICustomHealth) Info() APIInfo {
	return APIInfo{"testing", ""}
}

func TestServerCustomHealth(t *testing.T) {
	apiDef := new(testAPICustomHealth)

	cfg := testConfig()
	svr, err := New(tl(t), cfg, apiDef)
	require.NoError(t, err)

	testSvr := httptest.NewServer(svr.svr.Handler)
	defer testSvr.Close()

	rsp, err := http.Get(testSvr.URL + cfg.HealthPath)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
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

func TestServerAddr(t *testing.T) {
	apiDef := new(testAPICustomHealth)
	cfg := testConfig()
	svr, err := New(tl(t), cfg, apiDef)
	require.NoError(t, err)
	require.Equal(t, svr.svr.Addr, ":9090")
	cfg.Host = "127.0.0.1"
	svrWithHost, err := New(tl(t), cfg, apiDef)
	require.NoError(t, err)
	require.Equal(t, svrWithHost.svr.Addr, "127.0.0.1:9090")
}
