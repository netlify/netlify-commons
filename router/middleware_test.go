package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
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
