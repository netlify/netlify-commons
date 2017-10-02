package discovery

import (
	"fmt"

	"github.com/miekg/dns"
)

type Endpoint struct {
	Name string
	Port uint16
}

func DiscoverEndpoints(serviceDNS string) ([]Endpoint, error) {
	endpoints := []Endpoint{}
	qType := dns.StringToType["SRV"]
	serviceDNS = dns.Fqdn(serviceDNS)

	client := &dns.Client{}
	msg := &dns.Msg{}

	msg.SetQuestion(serviceDNS, qType)

	dnsserver, err := getDNSServer()
	if err != nil {
		return endpoints, err
	}

	response, _, err := client.Exchange(msg, dnsserver)
	if err != nil {
		return endpoints, err
	}

	if msg.Id != response.Id {
		return nil, fmt.Errorf("DNS ID mismatch, request: %d, response: %d", msg.Id, response.Id)
	}

	for _, v := range response.Answer {
		if srv, ok := v.(*dns.SRV); ok {
			endpoints = append(endpoints, Endpoint{
				Name: srv.Target,
				Port: srv.Port,
			})
		}
	}

	return endpoints, nil
}

func getDNSServer() (string, error) {
	servers, err := parseResolvConf()
	return servers[0], err
}

func parseResolvConf() ([]string, error) {
	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return []string{}, err
	}

	return conf.Servers, nil
}
