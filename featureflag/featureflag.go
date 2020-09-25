package featureflag

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"

	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
)

type Client interface {
	Enabled(key, userID string) bool
	EnabledUser(key string, user ld.User) bool

	Variation(key, defaultVal, userID string) string
	VariationUser(key string, defaultVal string, user ld.User) string

	AllEnabledFlags(userID string) []string
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

	if cfg.updateProcessorFactory != nil {
		config.UpdateProcessorFactory = cfg.updateProcessorFactory
	}

	if logger == nil {
		logger = noopLogger()
	}

	inner, err := ld.MakeCustomClient(cfg.Key, config, cfg.RequestTimeout)
	if err != nil {
		logger.WithError(err).Error("Unable to construct LD client")
	}
	return &ldClient{inner, logger}, err
}

func (c *ldClient) Enabled(key string, userID string) bool {
	return c.EnabledUser(key, ld.NewUser(userID))
}

func (c *ldClient) EnabledUser(key string, user ld.User) bool {
	res, err := c.BoolVariation(key, user, false)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	return res
}

func (c *ldClient) Variation(key, defaultVal, userID string) string {
	return c.VariationUser(key, defaultVal, ld.NewUser(userID))
}

func (c *ldClient) VariationUser(key string, defaultVal string, user ld.User) string {
	res, err := c.StringVariation(key, user, defaultVal)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	return res
}

func (c *ldClient) AllEnabledFlags(userId string) []string {
	// Ask launch darkly for all the flags for the user, return ld.FeatureFlagsState
	res := c.AllFlagsState(ld.NewUser(userId), ld.DetailsOnlyForTrackedFlags)
	flagMap := res.ToValuesMap()
	for key, value := range flagMap {
		// get the "true" flags
	}
	return []string{}
}

func noopLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(ioutil.Discard)
	return l
}
