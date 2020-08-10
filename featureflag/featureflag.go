package featureflag

import (
	"time"

	"github.com/sirupsen/logrus"

	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
)

type Client interface {
	Enabled(string, string) bool
	Variation(string, string, string) string
}

type ldClient struct {
	*ld.LDClient
	log *logrus.Entry
}

func NewClientWithConfig(cfg *Config, logger *logrus.Entry) (Client, error) {
	return NewClient(cfg.Key, cfg.RequestTimeout, logger)
}

func NewClient(key string, reqTimeout time.Duration, logger *logrus.Entry) (Client, error) {
	config := ld.DefaultConfig
	if key == "" {
		config.Offline = true
	}

	inner, err := ld.MakeCustomClient(key, config, reqTimeout)
	return &ldClient{inner, logger}, err
}

func (c *ldClient) Enabled(key string, userID string) bool {
	res, err := c.BoolVariation(key, ld.NewUser(userID), false)
	if err != nil && c.log != nil {
		c.log.WithError(err).WithField("key", key).Errorf("Failed to load feature flag")
	}
	return res
}

func (c *ldClient) Variation(key, defaultVal, userID string) string {
	res, err := c.StringVariation(key, ld.NewUser(userID), defaultVal)
	if err != nil && c.log != nil {
		c.log.WithError(err).WithField("key", key).Errorf("Failed to load feature flag")
	}
	return res
}

type MockClient struct {
	BoolVars   map[string]bool
	StringVars map[string]string
}

func (c MockClient) Enabled(key, _ string) bool {
	return c.BoolVars[key]
}

func (c MockClient) Variation(key, defaultVal, _ string) string {
	res, ok := c.StringVars[key]
	if !ok {
		return defaultVal
	}
	return res
}
