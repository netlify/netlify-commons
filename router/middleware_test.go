package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/netlify/netlify-commons/tracing"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckAuth(t *testing.T) {
	validKey := "testkey"
	invalidKey := "nopekey"
	emptyKey := ""

	makeRequest := func(req *http.Request) *httptest.ResponseRecorder {
		r := New(logrus.WithField("test", "CheckAuth"))
		r.Use(CheckAuth(validKey))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}

	t.Run("valid key", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validKey))
		rsp := makeRequest(req)
		assert.Equal(t, http.StatusOK, rsp.Code)
	})
	t.Run("lower case bearer", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", validKey))
		rsp := makeRequest(req)
		assert.Equal(t, http.StatusOK, rsp.Code)
	})

	t.Run("invalid key", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", invalidKey))
		rsp := makeRequest(req)
		assert.Equal(t, http.StatusUnauthorized, rsp.Code)
	})
	t.Run("no header", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		rsp := makeRequest(req)
		assert.Equal(t, http.StatusUnauthorized, rsp.Code)
	})
	t.Run("empty key", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", emptyKey))
		rsp := makeRequest(req)
		assert.Equal(t, http.StatusUnauthorized, rsp.Code)
	})
	t.Run("invalid Authorization value", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("what even is this %s", invalidKey))
		rsp := makeRequest(req)
		assert.Equal(t, http.StatusUnauthorized, rsp.Code)
	})

}

func TestRecoveryLogging(t *testing.T) {
	logger, hook := test.NewNullLogger()

	mw := Recoverer(logger)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://doesntmatter.com", nil)
	req.Header.Set(tracing.HeaderRequestUUID, "123456")

	// this should be captured by the recorder
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("because I should"))
	}))

	handler.ServeHTTP(rec, req)
	require.NotEmpty(t, hook.AllEntries)
	var lineID int
	for _, e := range hook.AllEntries() {
		assert.Equal(t, "123456", e.Data["request_id"], "missing the request_id: %v", e.Data)
		assert.Equal(t, lineID, e.Data["trace_line"], "trace_line isn't in order: %v", e.Data)
		lineID++
	}
}
