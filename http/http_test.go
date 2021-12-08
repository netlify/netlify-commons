package http

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"216.58.194.206", false},
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"192.168.0.1", true},
		{"172.16.0.0", true},
		{"169.254.169.254", true},
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if ip == nil {
			require.Fail(t, "failed to parse IP")
		}
		assert.Equal(t, tt.expected, containsPrivateIP([]net.IP{ip}))
	}
}

func TestBlockList(t *testing.T) {
	t.Run("safe http client", func(t *testing.T) {
		client := SafeHTTPClient(&http.Client{}, logrus.New())
		testBlockList(t, client)
	})
	t.Run("safe dial", func(t *testing.T) {
		tr := http.Transport{
			DialContext: SafeDial(&net.Dialer{}),
		}
		client := &http.Client{Transport: &tr}
		testBlockList(t, client)
	})
}

func TestAllowList(t *testing.T) {
	_, local, err := net.ParseCIDR("127.0.0.1/8")
	require.NoError(t, err)

	t.Run("safe http client", func(t *testing.T) {
		client := SafeHTTPClient(&http.Client{}, logrus.New(), local)
		testAllowList(t, client)
	})
	t.Run("safe dial", func(t *testing.T) {
		tr := http.Transport{
			DialContext: SafeDial(&net.Dialer{}, local),
		}
		client := &http.Client{Transport: &tr}
		testAllowList(t, client)
	})
}

func testBlockList(t *testing.T, client *http.Client) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Done"))
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	// It allows accessing non-local addresses
	_, err = client.Get("https://google.com")
	require.Nil(t, err)

	// It blocks the local IP.
	_, err = client.Get(ts.URL)
	require.NotNil(t, err)

	// It blocks localhost.
	_, err = client.Get("http://localhost:" + tsURL.Port())
	require.NotNil(t, err)

	// It works when reusing pooled connections.
	for i := 0; i < 50; i++ {
		res, err := client.Get("http://localhost:" + tsURL.Port())
		assert.Nil(t, res)
		assert.NotNil(t, err)
	}
}

func testAllowList(t *testing.T, client *http.Client) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Done"))
	}))
	defer ts.Close()
	_, err := client.Get(ts.URL)
	assert.NoError(t, err)
}
