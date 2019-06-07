package messaging

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/nats-io/go-nats"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/netlify/netlify-commons/nconf"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var silent *logrus.Entry

func init() {
	l := logrus.New()
	l.Out = ioutil.Discard
	silent = logrus.NewEntry(l)
}

func ConfigureNatsConnection(config *nconf.NatsConfig, log logrus.FieldLogger) (*nats.Conn, error) {
	if log == nil {
		log = silent
	}

	if config == nil {
		log.Debug("Skipping nats connection because there is no config")
		return nil, nil
	}

	if !config.TLS.Enabled {
		log.Warn("Connection to NATS servers is not using TLS")
	}

	if err := config.LoadServerNames(); err != nil {
		return nil, errors.Wrap(err, "Failed to discover new servers")
	}

	log.WithFields(config.Fields()).Info("Going to connect to nats servers")
	nc, err := ConnectToNats(config, ErrorHandler(log), nats.MaxReconnects(-1))
	if err != nil {
		return nil, err
	}

	return nc, nil
}

func ConnectToNats(config *nconf.NatsConfig, opts ...nats.Option) (*nats.Conn, error) {
	tlsConfig, err := config.TLS.TLSConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to configure TLS")
	}

	// If TLS is enabled, make the connection to NATS secure
	if config.TLS.Enabled {
		opts = append(opts, NatsRootCAs(tlsConfig))
	}

	switch strings.ToLower(config.Auth.Method) {
	case nconf.NatsAuthMethodUser:
		opts = append(opts, nats.UserInfo(config.Auth.User, config.Auth.Password))
	case nconf.NatsAuthMethodToken:
		opts = append(opts, nats.Token(config.Auth.Token))
	case nconf.NatsAuthMethodTLS:
		// if using TLS auth, make sure the client certificate is loaded
		if tlsConfig == nil || len(tlsConfig.Certificates) == 0 {
			return nil, fmt.Errorf("TLS auth method is configured but no certificate was loaded")
		}
		opts = append(opts, nats.Secure(tlsConfig))
	default:
		return nil, fmt.Errorf("Invalid auth method: '%s'", config.Auth.Method)
	}

	return nats.Connect(config.ServerString(), opts...)
}

func ConfigureNatsStreaming(config *nconf.NatsConfig, log logrus.FieldLogger) (stan.Conn, error) {
	// connect to the underlying instance
	nc, err := ConfigureNatsConnection(config, log)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to underlying NATS servers")
	}

	conn, err := ConnectToNatsStreaming(nc, config, log)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to NATS streaming")
	}

	return conn, nil
}

func ConnectToNatsStreaming(nc *nats.Conn, config *nconf.NatsConfig, log logrus.FieldLogger) (stan.Conn, error) {
	if config.ClusterID == "" {
		return nil, errors.New("Must provide a cluster ID to connect to streaming nats")
	}
	if config.ClientID == "" {
		config.ClientID = fmt.Sprintf("generated-%d", time.Now().Nanosecond())
		log.WithField("client_id", config.ClientID).Debug("No client ID specified, generating a random one")
	}

	// connect to the actual instance
	log.WithFields(config.Fields()).Debugf("Connecting to nats streaming cluster %s", config.ClusterID)
	return stan.Connect(config.ClusterID, config.ClientID, stan.NatsConn(nc))
}

func ErrorHandler(log logrus.FieldLogger) nats.Option {
	errLogger := log.WithField("component", "error-logger")
	handler := func(conn *nats.Conn, sub *nats.Subscription, natsErr error) {
		err := natsErr

		l := errLogger.WithFields(logrus.Fields{
			"subject":     sub.Subject,
			"group":       sub.Queue,
			"conn_status": conn.Status(),
		})

		if err == nats.ErrSlowConsumer {
			pendingMsgs, _, perr := sub.Pending()
			if perr != nil {
				err = perr
			} else {
				l = l.WithField("pending_messages", pendingMsgs)
			}
		}

		l.WithError(err).Error("Error while consuming from " + sub.Subject)
	}
	return nats.ErrorHandler(handler)
}

// NatsRootCAs is a NATS helper option to provide the RootCAs pool from a tls.Config struct. If Secure is
// not already set this will set it as well.
func NatsRootCAs(tlsConf *tls.Config) nats.Option {
	return func(o *nats.Options) error {
		if tlsConf.RootCAs == nil {
			return fmt.Errorf("nats: the RootCAs pool from the given tls.Config is nil")
		}

		if o.TLSConfig == nil {
			o.TLSConfig = &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: tlsConf.InsecureSkipVerify,
			}
		}

		o.TLSConfig.RootCAs = tlsConf.RootCAs
		o.Secure = true

		return nil
	}
}
