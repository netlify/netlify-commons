package featureflag

import (
	"github.com/sirupsen/logrus"

	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
)

type Client interface {
	Enabled(key, userID string) bool
	Variation(key, defaultVal, userID string) string
}

type ldClient struct {
	*ld.LDClient
	log logrus.FieldLogger
}

var _ Client = &ldClient{}

func NewClient(cfg *Config, logger logrus.FieldLogger) (Client, error) {
	config := ld.DefaultConfig

	if !cfg.Enabled {
		config.Offline = true
	}

	if logger == nil {
		logger = logrus.New()
	}

	inner, err := ld.MakeCustomClient(cfg.Key, config, cfg.RequestTimeout)
	if err != nil {
		logger.WithError(err).Error("Unable to construct LD client")
		return nil, err
	}

	return &ldClient{inner, logger}, nil
}

func (c *ldClient) Enabled(key string, userID string) bool {
	res, err := c.BoolVariation(key, ld.NewUser(userID), false)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	return res
}

func (c *ldClient) Variation(key, defaultVal, userID string) string {
	res, err := c.StringVariation(key, ld.NewUser(userID), defaultVal)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	return res
}
