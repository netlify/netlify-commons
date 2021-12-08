package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"

	"github.com/sirupsen/logrus"
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

func blocksContainsAny(blocks []*net.IPNet, ips []net.IP) bool {
	for _, block := range blocks {
		for _, ip := range ips {
			if block.Contains(ip) {
				return true
			}
		}
	}
	return false
}

func containsPrivateIP(ips []net.IP) bool {
	return blocksContainsAny(privateIPBlocks, ips)
}

type noLocalTransport struct {
	inner         http.RoundTripper
	errlog        logrus.FieldLogger
	allowedBlocks []*net.IPNet
}

func (no noLocalTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(req.Context())

	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			host, _, err := net.SplitHostPort(hostPort)
			if err != nil {
				cancel()
				no.errlog.WithError(err).Error("Cancelled request due to error in address parsing")
				return
			}

			ips, err := net.LookupIP(host)
			if err != nil || len(ips) == 0 {
				cancel()
				no.errlog.WithError(err).Error("Cancelled request due to error in host lookup")
				return
			}

			if blocksContainsAny(no.allowedBlocks, ips) {
				return
			}

			if containsPrivateIP(ips) {
				cancel()
				no.errlog.Error("Cancelled attempted request to ip in private range")
				return
			}
		},
	})

	req = req.WithContext(ctx)
	return no.inner.RoundTrip(req)
}

// SafeRoundtripper blocks requests to private ip ranges
// Deprecated: use SafeDial instead
func SafeRoundtripper(trans http.RoundTripper, log logrus.FieldLogger, allowedBlocks ...*net.IPNet) http.RoundTripper {
	if trans == nil {
		trans = http.DefaultTransport
	}

	ret := &noLocalTransport{
		inner:         trans,
		errlog:        log.WithField("transport", "local_blocker"),
		allowedBlocks: allowedBlocks,
	}

	return ret
}

// SafeHTTPClient blocks requests to private ip ranges
// Deprecated: use SafeDial instead
func SafeHTTPClient(client *http.Client, log logrus.FieldLogger, allowedBlocks ...*net.IPNet) *http.Client {
	client.Transport = SafeRoundtripper(client.Transport, log, allowedBlocks...)

	return client
}

type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// SafeDial wraps a *net.Dialer and restricts connections to private ip ranges.
func SafeDial(dialer *net.Dialer, allowedBlocks ...*net.IPNet) DialFunc {
	d := &safeDialer{dialer: dialer, allowedBlocks: allowedBlocks}
	return d.dialContext
}

type safeDialer struct {
	allowedBlocks []*net.IPNet
	dialer        *net.Dialer
}

func (d *safeDialer) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	addr := conn.RemoteAddr().String()

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("safe dialer: invalid address: %w", err)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		_ = conn.Close()
		return nil, fmt.Errorf("safe dialer: invalid ip: %v", host)
	}

	for _, block := range d.allowedBlocks {
		if block.Contains(ip) {
			return conn, nil
		}
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			_ = conn.Close()
			return nil, errors.New("safe dialer: private ip not allowed")
		}
	}

	return conn, nil
}
