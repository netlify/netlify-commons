package discovery

import (
	"net"
)

type Endpoint struct {
	Name string
	Port uint16
}

func DiscoverEndpoints(serviceDNS string) ([]Endpoint, error) {
	endpoints := []Endpoint{}
	_, remotes, err := net.LookupSRV("", "", serviceDNS)

	if err != nil {
		return endpoints, err
	}

	for _, n := range remotes {
		endpoints = append(endpoints, Endpoint{
			Name: n.Target,
			Port: n.Port,
		})
	}

	return endpoints, nil
}
