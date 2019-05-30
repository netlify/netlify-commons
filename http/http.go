package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
)

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"100.64.0.0/10",  // RFC6598
		"172.16.0.0/12",  // RFC1918
		"192.0.0.0/24",   // RFC6890
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func isLocalAddress(addr string) bool {
	ip := net.ParseIP(addr)
	return isPrivateIP(ip)
}

// SafeDialContext exchanges a DialContext for a SafeDialContext that will never dial a reserved IP range
func SafeDialContext(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if isLocalAddress(addr) {
			return nil, errors.New("Connection to local network address denied")
		}

		return dialContext(ctx, network, addr)
	}

}

// transport is an http.RoundTripper that keeps track of the in-flight
// request and implements hooks to report HTTP tracing events.
type tracingTransport struct {
	trace     *httptrace.ClientTrace
	transport http.RoundTripper
	req       *http.Request
	cancel    context.CancelFunc
}

// RoundTrip wraps http.DefaultTransport.RoundTrip to keep track
// of the current request.
func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(httptrace.WithClientTrace(req.Context(), t.trace))
	t.req = req.WithContext(ctx)
	t.cancel = cancel
	return t.transport.RoundTrip(t.req)
}

// GotConn prints whether the connection has been used previously
// for the current request.
func (t *tracingTransport) DNSDone(info httptrace.DNSDoneInfo) {
	fmt.Printf("Got dns info: %v\n", info)
	for _, addr := range info.Addrs {
		fmt.Printf("Checking addr: %v\n", addr)
		if isPrivateIP(addr.IP) {
			fmt.Printf("Got private IP %v\n", addr.IP)
			t.cancel()
		}
	}
}

func SafeHttpClient(client *http.Client) *http.Client {
	transport := &tracingTransport{transport: client.Transport}
	fmt.Printf("Transport in client is: %v\n", transport.transport)
	if transport.transport == nil {
		transport.transport = http.DefaultTransport
	}
	transport.trace = &httptrace.ClientTrace{DNSDone: transport.DNSDone}
	client.Transport = transport

	return client
}
