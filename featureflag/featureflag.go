package featureflag

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"

	"gopkg.in/launchdarkly/go-sdk-common.v1/ldvalue"
	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
	"gopkg.in/launchdarkly/go-server-sdk.v4/ldlog"
)

type Client interface {
	Enabled(key, userID string, attrs ...Attr) bool
	EnabledUser(key string, user ld.User) bool

	Variation(key, defaultVal, userID string, attrs ...Attr) string
	VariationUser(key string, defaultVal string, user ld.User) string

	AllEnabledFlags(key string) []string
	AllEnabledFlagsUser(key string, user ld.User) []string
}

type ldClient struct {
	*ld.LDClient
	log          logrus.FieldLogger
	defaultAttrs []Attr
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

	configureLogger(&config.Loggers, logger)

	if cfg.RelayHost != "" {
		config.BaseUri = cfg.RelayHost
		config.StreamUri = cfg.RelayHost
		config.EventsUri = cfg.RelayHost
	}

	if cfg.DisableEvents {
		config.SendEvents = false
	}

	inner, err := ld.MakeCustomClient(cfg.Key, config, cfg.RequestTimeout.Duration)
	if err != nil {
		logger.WithError(err).Error("Unable to construct LD client")
	}

	var defaultAttrs []Attr
	for k, v := range cfg.DefaultAttrs {
		defaultAttrs = append(defaultAttrs, StringAttr(k, v))
	}
	return &ldClient{inner, logger, defaultAttrs}, err
}

func (c *ldClient) Enabled(key string, userID string, attrs ...Attr) bool {
	return c.EnabledUser(key, c.userWithAttrs(userID, attrs))
}

func (c *ldClient) EnabledUser(key string, user ld.User) bool {
	res, err := c.BoolVariation(key, user, false)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	return res
}

func (c *ldClient) Variation(key, defaultVal, userID string, attrs ...Attr) string {
	return c.VariationUser(key, defaultVal, c.userWithAttrs(userID, attrs))
}

func (c *ldClient) VariationUser(key string, defaultVal string, user ld.User) string {
	res, err := c.StringVariation(key, user, defaultVal)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	return res
}

func (c *ldClient) AllEnabledFlags(key string) []string {
	return c.AllEnabledFlagsUser(key, ld.NewUser(key))
}

func (c *ldClient) AllEnabledFlagsUser(key string, user ld.User) []string {
	res := c.AllFlagsState(user, ld.DetailsOnlyForTrackedFlags)
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

func (c *ldClient) userWithAttrs(id string, attrs []Attr) ld.User {
	b := ld.NewUserBuilder(id)
	for _, attr := range c.defaultAttrs {
		b.Custom(attr.Name, attr.Value)
	}
	for _, attr := range attrs {
		b.Custom(attr.Name, attr.Value)
	}
	return b.Build()
}

type Attr struct {
	Name  string
	Value ldvalue.Value
}

func StringAttr(name, value string) Attr {
	return Attr{Name: name, Value: ldvalue.String(value)}
}

func configureLogger(ldLogger *ldlog.Loggers, log logrus.FieldLogger) {
	if log == nil {
		l := logrus.New()
		l.SetOutput(ioutil.Discard)
		log = l
	}
	log = log.WithField("component", "launch_darkly")

	ldLogger.SetBaseLoggerForLevel(ldlog.Debug, &wrapLog{log.Debugln, log.Debugf})
	ldLogger.SetBaseLoggerForLevel(ldlog.Info, &wrapLog{log.Infoln, log.Infof})
	ldLogger.SetBaseLoggerForLevel(ldlog.Warn, &wrapLog{log.Warnln, log.Warnf})
	ldLogger.SetBaseLoggerForLevel(ldlog.Error, &wrapLog{log.Errorln, log.Errorf})
}

type wrapLog struct {
	println func(values ...interface{})
	printf  func(format string, values ...interface{})
}

func (l *wrapLog) Println(values ...interface{}) {
	l.println(values...)
}

func (l *wrapLog) Printf(format string, values ...interface{}) {
	l.printf(format, values...)
}
