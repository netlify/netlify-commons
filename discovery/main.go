package discovery

import (
	"fmt"
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

func DiscoverServersAsStrings(serviceDNS string) ([]string, error) {
	urls := []string{}

	endpoints, err := DiscoverEndpoints(serviceDNS)
	if err != nil {
		return urls, err
	}

	for _, endpoint := range endpoints {
		urls = append(urls, fmt.Sprintf("%s:%d", endpoint.Target, endpoint.Port))
	}

	return urls, nil
}
