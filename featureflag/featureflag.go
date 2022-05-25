package featureflag

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldlog"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	ld "gopkg.in/launchdarkly/go-server-sdk.v5"
	"gopkg.in/launchdarkly/go-server-sdk.v5/interfaces"
	"gopkg.in/launchdarkly/go-server-sdk.v5/interfaces/flagstate"
	"gopkg.in/launchdarkly/go-server-sdk.v5/ldcomponents"
)

type Client interface {
	Enabled(key, userID string, attrs ...Attr) bool
	EnabledUser(key string, user lduser.User) bool

	Variation(key, defaultVal, userID string, attrs ...Attr) string
	VariationUser(key string, defaultVal string, user lduser.User) string

	Int(key string, defaultVal int, userID string, attrs ...Attr) int
	IntUser(key string, defaultVal int, user lduser.User) int

	AllEnabledFlags(key string) []string
	AllEnabledFlagsUser(key string, user lduser.User) []string
}

type ldClient struct {
	*ld.LDClient
	log          logrus.FieldLogger
	defaultAttrs []Attr
}

var _ Client = &ldClient{}

func NewClient(cfg *Config, logger logrus.FieldLogger) (Client, error) {
	config := ld.Config{}

	if !cfg.Enabled {
		config.Offline = true
	}

	if cfg.updateProcessorFactory != nil {
		config.DataSource = cfg.updateProcessorFactory
		config.Events = ldcomponents.NoEvents()
	}

	config.Logging = configureLogger(logger)

	if cfg.RelayHost != "" {
		config.ServiceEndpoints = ldcomponents.RelayProxyEndpoints(cfg.RelayHost)
	}

	if cfg.DisableEvents {
		config.Events = ldcomponents.NoEvents()
	}

	inner, err := ld.MakeCustomClient(cfg.Key, config, cfg.RequestTimeout.Duration)
	if err != nil {
		logger.WithError(err).Error("Unable to construct LD client")
	}

	var defaultAttrs []Attr
	for k, v := range cfg.DefaultUserAttrs {
		defaultAttrs = append(defaultAttrs, StringAttr(k, v))
	}
	return &ldClient{inner, logger, defaultAttrs}, err
}

func (c *ldClient) Enabled(key string, userID string, attrs ...Attr) bool {
	return c.EnabledUser(key, c.userWithAttrs(userID, attrs))
}

func (c *ldClient) EnabledUser(key string, user lduser.User) bool {
	res, err := c.BoolVariation(key, user, false)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	return res
}

func (c *ldClient) Variation(key, defaultVal, userID string, attrs ...Attr) string {
	return c.VariationUser(key, defaultVal, c.userWithAttrs(userID, attrs))
}

func (c *ldClient) VariationUser(key string, defaultVal string, user lduser.User) string {
	res, err := c.StringVariation(key, user, defaultVal)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	return res
}

func (c *ldClient) Int(key string, defaultValue int, userID string, attrs ...Attr) int {
	return c.IntUser(key, defaultValue, c.userWithAttrs(userID, attrs))
}

func (c *ldClient) IntUser(key string, defaultVal int, user lduser.User) int {
	res, err := c.IntVariation(key, user, defaultVal)
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Failed to load feature flag")
	}
	// DefaultValue will be returned if IntVariation returns an error
	return res
}

func (c *ldClient) AllEnabledFlags(key string) []string {
	return c.AllEnabledFlagsUser(key, lduser.NewUser(key))
}

func (c *ldClient) AllEnabledFlagsUser(key string, user lduser.User) []string {
	res := c.AllFlagsState(user, flagstate.OptionDetailsOnlyForTrackedFlags())
	flagMap := res.ToValuesMap()

	var flags []string
	for flag, value := range flagMap {
		if value.BoolValue() {
			flags = append(flags, flag)
		}
	}

	return flags
}

func (c *ldClient) userWithAttrs(id string, attrs []Attr) lduser.User {
	b := lduser.NewUserBuilder(id)
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

func configureLogger(log logrus.FieldLogger) interfaces.LoggingConfigurationFactory {
	if log == nil {
		l := logrus.New()
		l.SetOutput(ioutil.Discard)
		log = l
	}
	log = log.WithField("component", "launch_darkly")

	return &logCreator{log: log}
}

type logCreator struct {
	log logrus.FieldLogger
}

func (c *logCreator) CreateLoggingConfiguration(b interfaces.BasicConfiguration) (interfaces.LoggingConfiguration, error) {
	logger := ldlog.NewDefaultLoggers()
	logger.SetBaseLoggerForLevel(ldlog.Debug, &wrapLog{c.log.Debugln, c.log.Debugf})
	logger.SetBaseLoggerForLevel(ldlog.Info, &wrapLog{c.log.Infoln, c.log.Infof})
	logger.SetBaseLoggerForLevel(ldlog.Warn, &wrapLog{c.log.Warnln, c.log.Warnf})
	logger.SetBaseLoggerForLevel(ldlog.Error, &wrapLog{c.log.Errorln, c.log.Errorf})
	return ldcomponents.Logging().Loggers(logger).CreateLoggingConfiguration(b)
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
