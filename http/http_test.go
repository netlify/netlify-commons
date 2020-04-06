package http

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, tt.expected, isPrivateIP(ip))
	}
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

	// It blocks the local IP.
	_, err = client.Get(ts.URL)
	assert.NotNil(t, err)

	// It blocks localhost.
	_, err = client.Get("http://localhost:" + tsURL.Port())
	assert.NotNil(t, err)

	// It succeeds when the local IP range used by the testserver is removed from
	// the blacklist.
	ipNet := popMatchingBlock(net.ParseIP(tsURL.Hostname()))
	_, err = client.Get(ts.URL)
	assert.Nil(t, err)
	privateIPBlocks = append(privateIPBlocks, ipNet)

	// It allows whitelisting for local development.
	client = SafeHTTPClient(&http.Client{}, logrus.New(), ipNet)
	_, err = client.Get(ts.URL)
	assert.Nil(t, err)
}

func popMatchingBlock(ip net.IP) *net.IPNet {
	for i, ipNet := range privateIPBlocks {
		if ipNet.Contains(ip) {
			privateIPBlocks = append(privateIPBlocks[:i], privateIPBlocks[i+1:]...)
			return ipNet
		}
	}
	return nil
}

type testReuseTransport struct {
	reusedConn bool
}

func (r *testReuseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			r.reusedConn = info.Reused
		},
	})

	req = req.WithContext(ctx)
	return http.DefaultTransport.RoundTrip(req)
}

func TestSafeHTTPClientReuse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Done"))
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	tr := &testReuseTransport{}
	client := SafeHTTPClient(&http.Client{
		Transport: tr,
	}, logrus.New())

	for !tr.reusedConn {
		_, err = client.Get("http://localhost:" + tsURL.Port())
		assert.NotNil(t, err)
	}
}

func TestSafeHTTPClientReuse2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Done"))
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	client := SafeHTTPClient(&http.Client{}, logrus.New())
	var reusedConn bool
	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			reusedConn = info.Reused
		},
	}

	req, err := http.NewRequest("GET", "http://localhost:"+tsURL.Port(), nil)
	assert.Nil(t, err)
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	for i := 0; i < 5000 && !reusedConn; i++ {
		_, err := client.Do(req)
		assert.NotNil(t, err)
	}
}
