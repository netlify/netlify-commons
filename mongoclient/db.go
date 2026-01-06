package mongoclient

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/sirupsen/logrus"
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
	AppName            string
	TLS                *TLSConfig
	Servers            []string
	ReplSetName        string
	ConnTimeout        time.Duration
	Auth               *Auth
	SecondaryPreferred bool
}

func ConnectWithOptions(log logrus.FieldLogger, replSet string, servers []string, opts ...Option) (*mongo.Client, error) {
	finalOpts := options.Client().
		SetConnectTimeout(time.Minute). // use a default
		SetReplicaSet(replSet).
		SetHosts(servers)
	for _, opt := range opts {
		if err := opt(finalOpts); err != nil {
			return nil, err
		}
	}
	log.WithFields(logrus.Fields{
		"servers":     strings.Join(finalOpts.Hosts, ","),
		"replica_set": finalOpts.ReplicaSet,
	}).Debug("Dialing database")

	client, err := mongo.Connect(context.Background(), finalOpts)
	if err != nil {
		return nil, err
	}

	// Connect does not block for server discovery, so we should ping to ensure
	// we are actually connected
	ctx, cancel := context.WithTimeout(context.Background(), *finalOpts.ConnectTimeout+time.Second)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.WithError(err).Error("Failed to ping primary Mongo instance")
		return nil, err
	}

	log.Debug("Connected to MongoDB")
	return client, nil
}

// Connect connects to MongoDB
func Connect(log logrus.FieldLogger, config *Config) (*mongo.Client, error) {
	opts := []Option{
		AppName(config.AppName),
	}
	if config.TLS != nil {
		opts = append(opts, TLSOption(log, *config.TLS))
	}
	if config.Auth != nil {
		opts = append(opts, AuthOption(*config.Auth))
	}
	if config.SecondaryPreferred {
		opts = append(opts, SecondaryPreferred())
	}

	return ConnectWithOptions(log, config.ReplSetName, config.Servers, opts...)
}
