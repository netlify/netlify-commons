package messaging

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats"
	"github.com/sirupsen/logrus"

	"github.com/rybit/nats_logrus_hook"

	"github.com/netlify/netlify-commons/discovery"
	"github.com/netlify/netlify-commons/tls"
)

type NatsConfig struct {
	TLS            *tls.Config `mapstructure:"tls_conf"`
	ServiceDNSName string      `mapstructure:"service_dnsname"`
	Servers        []string    `mapstructure:"servers"`
	LogsSubject    string      `mapstructure:"log_subject"`
}

type MetricsConfig struct {
	Subject    string                  `mapstructure:"subject"`
	Dimensions *map[string]interface{} `mapstructure:"default_dims"`
}

// ServerString will build the proper string for nats connect
func (config *NatsConfig) ServerString() string {
	return strings.Join(config.Servers, ",")
}

func (config *NatsConfig) Fields() logrus.Fields {
	f := logrus.Fields{
		"logs_subject": config.LogsSubject,
		"servers":      strings.Join(config.Servers, ","),
	}

	if config.TLS != nil {
		f["ca_files"] = strings.Join(config.TLS.CAFiles, ",")
		f["key_file"] = config.TLS.KeyFile
		f["cert_file"] = config.TLS.CertFile
	}

	return f
}

func ConfigureNatsConnection(config *NatsConfig, log *logrus.Entry) (*nats.Conn, error) {
	if config == nil {
		log.Debug("Skipping nats connection because there is no config")
		return nil, nil
	}

	nc, err := ConnectToNats(config, ErrorHandler(log))
	if err != nil {
		return nil, err
	}

	if config.LogsSubject != "" {
		logrus.AddHook(nhook.NewNatsHook(nc, config.LogsSubject))
		log.WithField("subject", config.LogsSubject).Debug("Configured nats hook for logrus")
	}

	return nc, nil
}

// ConnectToNats will do a TLS connection to the nats servers specified
func ConnectToNats(config *NatsConfig, errHandler nats.ErrHandler) (*nats.Conn, error) {
	serversString := config.ServerString()

	if config.ServiceDNSName != "" {
		natsUrls := []string{}

		endpoints, err := discovery.DiscoverEndpoints(config.ServiceDNSName)
		if err != nil {
			return nil, err
		}

		for _, endpoint := range endpoints {
			natsUrls = append(natsUrls, fmt.Sprintf("nats://%s:%d", endpoint.Name, endpoint.Port))
		}

		serversString = strings.Join(natsUrls, ",")
	}

	options := []nats.Option{}
	if config.TLS != nil {
		tlsConfig, err := config.TLS.TLSConfig()
		if err != nil {
			return nil, err
		}
		if tlsConfig != nil {
			options = append(options, nats.Secure(tlsConfig))
		}
	}

	if errHandler != nil {
		options = append(options, nats.ErrorHandler(errHandler))
	}

	return nats.Connect(serversString, options...)
}

func ErrorHandler(log *logrus.Entry) nats.ErrHandler {
	errLogger := log.WithField("component", "error-logger")
	return func(conn *nats.Conn, sub *nats.Subscription, err error) {
		errLogger.WithError(err).WithFields(logrus.Fields{
			"subject":     sub.Subject,
			"group":       sub.Queue,
			"conn_status": conn.Status(),
		}).Error("Error while consuming from " + sub.Subject)
	}
}
