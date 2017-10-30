package nconf

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/netlify/netlify-commons/messaging"
	"github.com/netlify/netlify-commons/metrics"
	"github.com/netlify/netlify-commons/metrics/transport"
	"github.com/netlify/netlify-commons/tls"
)

type MetricsConfig struct {
	DataDog *struct {
		APIKey string `mapstructure:"api_key" envconfig:"API_KEY"`
		AppKey string `mapstructure:"app_key" split_words:"true"`
	} `mapstructure:"datadog"`

	SFXToken string `mapstructure:"sfx_token" envconfig:"SFX_TOKEN"`

	Nats *NatsConfig `mapstructure:"nats"`

	Namespace  string            `mapstructure:"namespace"`
	Dimensions map[string]string `mapstructure:"default_dims"`

	// for reporting cumulative counters on an interval
	ReportSec int `mapstructure:"report_sec" split_words:"true"`
}

type NatsConfig struct {
	TLS     *tls.Config `mapstructure:"tls_conf"`
	Servers []string    `mapstructure:"servers"`
	Subject string      `mapstructure:"subject"`
}

func ConfigureMetrics(mconf *MetricsConfig, log *logrus.Entry) error {
	if mconf == nil {
		log.Info("Skipping configuring metrics lib - no config specified")
		return nil
	}

	var err error
	ports := []metrics.Transport{}
	ports, err = appendNatsConfig(ports, mconf, log)
	if err != nil {
		return err
	}
	ports, err = appendDatadogConfig(ports, mconf, log)
	if err != nil {
		return err
	}
	ports, err = appendSignalFXConfig(ports, mconf, log)
	if err != nil {
		return err
	}

	if len(ports) > 0 {
		log.Infof("Configuring metrics with %d transports", len(ports))
		if len(ports) == 1 {
			metrics.Init(ports[0])
		} else {
			metrics.Init(transport.NewBroadcastTransport(ports))
		}
	}

	for k, v := range mconf.Dimensions {
		metrics.AddDimension(k, v)
	}
	metrics.Namespace(mconf.Namespace)

	metrics.StartReportingCumulativeCounters(
		time.Duration(mconf.ReportSec)*time.Second,
		log.WithField("component", "stats_report"),
	)

	return nil
}

func appendNatsConfig(ports []metrics.Transport, mconf *MetricsConfig, log *logrus.Entry) ([]metrics.Transport, error) {
	if mconf.Nats == nil {
		return ports, nil
	}

	if len(mconf.Nats.Servers) == 0 || mconf.Nats.Subject == "" {
		return ports, nil
	}

	log.Info("Configuring NATS transport for metrics")
	natsconf := &messaging.NatsConfig{
		TLS:     mconf.Nats.TLS,
		Servers: mconf.Nats.Servers,
	}
	nc, err := messaging.ConfigureNatsConnection(natsconf, log)
	if err != nil {
		log.WithError(err).Warn("Failed to setup nats connection")
		return nil, err
	}

	return append(ports, transport.NewNatsTransport(mconf.Nats.Subject, nc)), nil
}

func appendDatadogConfig(ports []metrics.Transport, mconf *MetricsConfig, log *logrus.Entry) ([]metrics.Transport, error) {
	if mconf.DataDog == nil {
		return ports, nil
	}

	if mconf.DataDog.APIKey == "" || mconf.DataDog.AppKey == "" {
		return ports, nil
	}

	log.Info("Configuring DataDog transport for metrics")
	t, err := transport.NewDataDogTransport(mconf.DataDog.APIKey, mconf.DataDog.AppKey)
	if err != nil {
		return nil, err
	}
	return append(ports, t), nil

}

func appendSignalFXConfig(ports []metrics.Transport, mconf *MetricsConfig, log *logrus.Entry) ([]metrics.Transport, error) {
	if mconf.SFXToken == "" {
		return ports, nil
	}
	log.Info("Configuring SignalFX transport for metrics")
	t, err := transport.NewSignalFXTransport(&transport.SFXConfig{AuthToken: mconf.SFXToken, ReportSec: mconf.ReportSec})
	if err != nil {
		return nil, err
	}
	return append(ports, t), nil
}
