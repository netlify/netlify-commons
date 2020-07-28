package mongo

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"

	"github.com/netlify/netlify-commons/nconf"
)

const (
	CollectionBlobs         = "blobs"
	CollectionResellers     = "resellers"
	CollectionUsers         = "users"
	CollectionSubscriptions = "bb_subscriptions"
	CollectionSites         = "projects"
)

type Config struct {
	TLS         *nconf.TLSConfig `mapstructure:"tls_conf"`
	DB          string           `mapstructure:"db"`
	Servers     []string         `mapstructure:"servers"`
	ReplSetName string           `mapstructure:"replset_name"`
	ConnTimeout int64            `mapstructure:"conn_timeout"`
}

type Auth struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Source   string `mapstructure:"source"`
}

// Connect connects to MongoDB
func Connect(config *Config, auth *Auth, log *logrus.Entry) (*mongo.Database, error) {
	opts := options.Client().
		SetConnectTimeout(time.Second * time.Duration(config.ConnTimeout)).
		SetReplicaSet(config.ReplSetName).
		SetHosts(config.Servers)

	if config.TLS != nil && config.TLS.Enabled {
		tlsLog := log.WithFields(logrus.Fields{
			"cert_file": config.TLS.CertFile,
			"key_file":  config.TLS.KeyFile,
			"ca_files":  strings.Join(config.TLS.CAFiles, ","),
		})

		tlsLog.Debug("Using TLS config")
		tlsConfig, err := config.TLS.TLSConfig()
		if err != nil {
			return nil, err
		}

		opts.SetTLSConfig(tlsConfig)
	} else {
		log.Debug("Skipping TLS config")
	}

	if auth.Enabled {
		creds := options.Credential{
			Username:   auth.Username,
			Password:   auth.Password,
			AuthSource: auth.Source,
		}

		opts.SetAuth(creds)
	}

	log.WithFields(logrus.Fields{
		"servers":     strings.Join(opts.Hosts, ","),
		"replica_set": config.ReplSetName,
	}).Debug("Dialing database")

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	log.WithField("db", config.DB).Debugf("Got session, Using database %s", config.DB)
	return client.Database(config.DB), nil
}
