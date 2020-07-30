package mongoclient

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

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

type Auth struct {
	Username string
	Password string
	Source   string
}

type Config struct {
	TLS         *nconf.TLSConfig
	Servers     []string
	ReplSetName string
	ConnTimeout time.Duration
	Auth        *Auth
}

// Connect connects to MongoDB
func Connect(config *Config, log *logrus.Entry) (*mongo.Client, error) {
	opts := options.Client().
		SetConnectTimeout(config.ConnTimeout).
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

	if config.Auth != nil {
		creds := options.Credential{
			Username:   config.Auth.Username,
			Password:   config.Auth.Password,
			AuthSource: config.Auth.Source,
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

	// Connect does not block for server discovery, so we should ping to ensure
	// we are actually connected
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnTimeout)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.WithError(err).Error("Failed to ping primary Mongo instance")
		return nil, err
	}

	log.Debug("Connected to MongoDB")
	return client, nil
}
