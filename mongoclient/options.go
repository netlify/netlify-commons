package mongoclient

import (
	"strings"

	"github.com/netlify/netlify-commons/nconf"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Option func(opt *options.ClientOptions) error

func TLSOption(log logrus.FieldLogger, config nconf.TLSConfig) Option {
	return func(opts *options.ClientOptions) error {
		if !config.Enabled {
			log.Debug("Skipping TLS config")
			return nil
		}

		log.WithFields(logrus.Fields{
			"cert_file": config.CertFile,
			"key_file":  config.KeyFile,
			"ca_files":  strings.Join(config.CAFiles, ","),
		}).Debug("Using TLS config")

		tlsConfig, err := config.TLSConfig()
		if err != nil {
			return err
		}

		opts.SetTLSConfig(tlsConfig)
		return nil
	}
}

func AuthOption(config Auth) Option {
	return func(opts *options.ClientOptions) error {
		creds := options.Credential{
			Username:   config.Username,
			Password:   config.Password,
			AuthSource: config.Source,
		}

		opts.SetAuth(creds)
		return nil
	}
}

func AppName(name string) Option {
	return func(opts *options.ClientOptions) error {
		if name != "" {
			opts.SetAppName(name)
		}
		return nil
	}
}

func SecondaryPreferred() Option {
	return func(opts *options.ClientOptions) error {
		opts.SetReadPreference(readpref.SecondaryPreferred())
		return nil
	}
}
