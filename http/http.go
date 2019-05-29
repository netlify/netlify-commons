package http

import (
	"context"
	"errors"
	"net"
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
