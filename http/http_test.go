package http

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type countServer struct {
	count int
}

func (c *countServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestSafeHTTPClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Done"))
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	client := SafeHTTPClient(&http.Client{}, logrus.New())

	// It blocks the local IP
	_, err = client.Get(ts.URL)
	assert.NotNil(t, err)

	// It blocks localhost
	_, err = client.Get("http://localhost:" + tsURL.Port())
	assert.NotNil(t, err)

	// It succeeds when the local IP range used by the testserver is removed from
	// the blacklist.
	ipNet := unshiftMatch(net.ParseIP(tsURL.Hostname()))
	defer func() {
		privateIPBlocks = append(privateIPBlocks, ipNet)
	}()

	_, err = client.Get(ts.URL)
	assert.Nil(t, err)
}

func unshiftMatch(ip net.IP) *net.IPNet {
	for i, ipNet := range privateIPBlocks {
		if ipNet.Contains(ip) {
			privateIPBlocks = append(privateIPBlocks[:i], privateIPBlocks[i+1:]...)
			return ipNet
		}
	}
	return nil
}
