package discovery

import (
	"net"
)

type Endpoint struct {
	Name string
	Port uint16
}

func DiscoverEndpoints(serviceDNS string) ([]*net.SRV, error) {
	_, remotes, err := net.LookupSRV("", "", serviceDNS)
	if err != nil {
		return nil, err
	}
	return remotes, nil
}
