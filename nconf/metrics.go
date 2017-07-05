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
		APIKey string `mapstructure:"api_key"`
		AppKey string `mapstructure:"app_key"`
	} `mapstructure:"datadog"`

	SFXToken string `mapstructure:"sfx_token"`

	Nats *struct {
		TLS     *tls.Config `mapstructure:"tls_conf"`
		Servers []string    `mapstructure:"servers"`
		Subject string      `mapstructure:"subject"`
	} `mapstructure:"nats"`

	Namespace  string                 `mapstructure:"namespace"`
	Dimensions map[string]interface{} `mapstructure:"default_dims"`

	// for reporting cumulative counters on an interval
	ReportSec int `mapstructure:"report_sec"`
}

func ConfigureMetrics(mconf *MetricsConfig, log *logrus.Entry) error {
	if mconf == nil {
		log.Info("Skipping configuring metrics lib - no config specified")
		return nil
	}

	ports := []metrics.Transport{}
	if mconf.Nats != nil {
		log.Info("Configuring NATS transport for metrics")
		natsconf := &messaging.NatsConfig{
			TLS:     mconf.Nats.TLS,
			Servers: mconf.Nats.Servers,
		}
		nc, err := messaging.ConnectToNats(natsconf, messaging.ErrorHandler(log))
		if err != nil {
			log.WithError(err).Warn("Failed to setup nats connection")
			return err
		}

		ports = append(ports, transport.NewNatsTransport(mconf.Nats.Subject, nc))
	}

	if mconf.DataDog != nil {
		log.Info("Configuring DataDog transport for metrics")
		t, err := transport.NewDataDogTransport(mconf.DataDog.APIKey, mconf.DataDog.AppKey)
		if err != nil {
			return err
		}
		ports = append(ports, t)
	}

	if mconf.SFXToken != "" {
		log.Info("Configuring SignalFX transport for metrics")
		t, err := transport.NewSignalFXTransport(&transport.SFXConfig{AuthToken: mconf.SFXToken})
		if err != nil {
			return err
		}
		ports = append(ports, t)
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
