package mongo

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/netlify/netlify-commons/nconf"
	"github.com/sirupsen/logrus"
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

func Connect(config *Config, log *logrus.Entry) (*mgo.Database, error) {
	info := &mgo.DialInfo{
		Addrs:          config.Servers,
		ReplicaSetName: config.ReplSetName,
		Timeout:        time.Second * time.Duration(config.ConnTimeout),
	}

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

		info.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), tlsConfig)
		}
	} else {
		log.Debug("Skipping TLS config")
	}

	log.WithFields(logrus.Fields{
		"servers":     strings.Join(info.Addrs, ","),
		"replica_set": config.ReplSetName,
	}).Debug("Dialing database")

	sess, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, err
	}

	log.WithField("db", config.DB).Debugf("Got session, Using database %s", config.DB)
	return sess.DB(config.DB), nil
}
