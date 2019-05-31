package nconf

import (
	"fmt"
	"strings"

	"github.com/netlify/netlify-commons/discovery"
	"github.com/sirupsen/logrus"
)

const (
	NatsAuthMethodUser  = "user"
	NatsAuthMethodToken = "token"
	NatsAuthMethodTLS   = "tls"
)

type NatsAuth struct {
	Method   string `mapstructure:"method"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Token    string `mapstructure:"token"`
}

type NatsConfig struct {
	TLS           *TLSConfig `mapstructure:"tls_conf"`
	DiscoveryName string     `mapstructure:"discovery_name" split_words:"true"`
	Servers       []string   `mapstructure:"servers"`
	Auth          NatsAuth   `mapstructure:"auth"`

	// for streaming
	ClusterID string `mapstructure:"cluster_id" split_words:"true"`
	ClientID  string `mapstructure:"client_id" split_words:"true"`
	StartPos  string `mapstructure:"start_pos" split_words:"true"`
}

func (c *NatsConfig) LoadServerNames() error {
	if c.DiscoveryName == "" {
		return nil
	}

	natsURLs := []string{}
	endpoints, err := discovery.DiscoverEndpoints(c.DiscoveryName)
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		natsURLs = append(natsURLs, fmt.Sprintf("nats://%s:%d", endpoint.Target, endpoint.Port))
	}

	c.Servers = natsURLs
	return nil
}

// ServerString will build the proper string for nats connect
func (config *NatsConfig) ServerString() string {
	return strings.Join(config.Servers, ",")
}

func (config *NatsConfig) Fields() logrus.Fields {
	f := logrus.Fields{
		"servers": strings.Join(config.Servers, ","),
	}

	if config.Auth.Method != "" {
		f["auth_method"] = config.Auth.Method
	}

	if config.TLS != nil {
		f["ca_files"] = strings.Join(config.TLS.CAFiles, ",")
		f["key_file"] = config.TLS.KeyFile
		f["cert_file"] = config.TLS.CertFile
	}

	if config.ClusterID != "" {
		f["client_id"] = config.ClientID
		f["cluster_id"] = config.ClusterID
	}

	return f
}
