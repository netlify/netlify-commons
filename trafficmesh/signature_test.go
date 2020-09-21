package trafficmesh

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/require"
)

func TestSignatureDecoder(t *testing.T) {
	const url = "https://example.org/index.html"

	var makeReq = func() *http.Request {
		req := httptest.NewRequest(http.MethodGet, "https://192.0.2.1/index.html", nil)
		req.Host = "example.org"
		return req
	}

	var addToken = func(req *http.Request, secret, url string) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sid":      "1",
			"did":      "2",
			"url":      url,
			"remapped": true,
		})
		sig, err := token.SignedString([]byte(secret))
		require.NoError(t, err)
		req.Header.Set(signatureHeader, sig)
	}

	dec := NewSignatureDecoder("secret")

	t.Run("no token", func(t *testing.T) {
		req := makeReq()
		payload, err := dec.DecodeSignature(req)
		require.NoError(t, err)
		require.Nil(t, payload)
	})

	t.Run("valid token", func(t *testing.T) {
		req := makeReq()
		addToken(req, "secret", url)

		payload, err := dec.DecodeSignature(req)
		require.NoError(t, err)

		require.Equal(t, payload.SiteID, "1")
		require.Equal(t, payload.DeployID, "2")
	})

	t.Run("wrong secret", func(t *testing.T) {
		req := makeReq()
		addToken(req, "wrong-secret", url)

		_, err := dec.DecodeSignature(req)
		require.Error(t, err)
	})

	t.Run("wrong path", func(t *testing.T) {
		req := makeReq()
		addToken(req, "secret", "http://example.org/stuff")
		_, err := dec.DecodeSignature(req)
		require.Error(t, err)
	})

	t.Run("wrong host", func(t *testing.T) {
		req := makeReq()
		addToken(req, "secret", "http://example.net/index.html")
		_, err := dec.DecodeSignature(req)
		require.Error(t, err)
	})

	t.Run("remapped flag", func(t *testing.T) {
		req := makeReq()
		addToken(req, "secret", url)
		payload, err := dec.DecodeSignature(req)
		require.NoError(t, err)
		require.True(t, payload.Remapped)
	})
}
