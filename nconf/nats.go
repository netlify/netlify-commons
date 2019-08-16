package nconf

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"

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

type NatsClientConfig struct {
	NatsConfig
	Subject string `mapstructure:"command_subject"`
	Group   string `mapstructure:"command_group"`

	// StartAt will configure where the client should resume the stream:
	// - `all`: all the messages available
	// - `last`: from where the client left off
	// - `new`: all new messages for the client
	// - `first`: from the first message available (default)
	// - other: if it isn't one of the above fields, it will try and parse the param as a go duration (e.g. 30s, 1h)
	StartAt string `mapstructure:"start_at"`
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
func (c *NatsConfig) ServerString() string {
	return strings.Join(c.Servers, ",")
}

func (c *NatsConfig) Fields() logrus.Fields {
	f := logrus.Fields{
		"servers": strings.Join(c.Servers, ","),
	}

	if c.Auth.Method != "" {
		f["auth_method"] = c.Auth.Method
	}

	if c.TLS != nil {
		f["ca_files"] = strings.Join(c.TLS.CAFiles, ",")
		f["key_file"] = c.TLS.KeyFile
		f["cert_file"] = c.TLS.CertFile
	}

	if c.ClusterID != "" {
		f["client_id"] = c.ClientID
		f["cluster_id"] = c.ClusterID
	}

	return f
}

func (c *NatsConfig) StartPoint() (stan.SubscriptionOption, error) {
	switch v := strings.ToLower(c.StartPos); v {
	case "all":
		return stan.DeliverAllAvailable(), nil
	case "last":
		return stan.StartWithLastReceived(), nil
	case "new":
		return stan.StartAt(pb.StartPosition_NewOnly), nil
	case "", "first":
		return stan.StartAt(pb.StartPosition_First), nil
	default:
		dur, err := time.ParseDuration(v)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse field as a duration")
		}
		return stan.StartAtTimeDelta(dur), nil
	}
}
