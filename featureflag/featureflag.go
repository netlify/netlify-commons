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

	AllEnabledFlags(key string) []string
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
		config.SendEvents = false
	}

	config.Loggers.SetBaseLogger(wrapLogger(logger))

	if cfg.RelayHost != "" {
		config.BaseUri = cfg.RelayHost
		config.StreamUri = cfg.RelayHost
		config.EventsUri = cfg.RelayHost
	}

	if cfg.DisableEvents {
		config.SendEvents = false
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

func (c *ldClient) AllEnabledFlags(key string) []string {
	res := c.AllFlagsState(ld.NewUser(key), ld.DetailsOnlyForTrackedFlags)
	flagMap := res.ToValuesMap()

	var flags []string
	for flag, value := range flagMap {
		switch value.(type) {
		case bool:
			if value == true {
				flags = append(flags, flag)
			}
		}
	}

	return flags
}

func wrapLogger(logger logrus.FieldLogger) infoToDebugLogger {
	if logger == nil {
		l := logrus.New()
		l.SetOutput(ioutil.Discard)
		logger = l
	}

	return infoToDebugLogger{logger.WithField("component", "launch_darkly")}
}

type infoToDebugLogger struct {
	log logrus.FieldLogger
}

func (l infoToDebugLogger) Println(values ...interface{}) {
	l.log.Debugln(values...)
}
func (l infoToDebugLogger) Printf(format string, values ...interface{}) {
	l.log.Debugf(format, values...)
}
