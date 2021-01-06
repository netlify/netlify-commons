package nconf

import (
	"fmt"
	"strings"

	"github.com/netlify/netlify-commons/featureflag"
	"github.com/netlify/netlify-commons/metriks"
	"github.com/netlify/netlify-commons/tracing"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RootArgs struct {
	Prefix     string
	ConfigFile string
}

type RootConfig struct {
	Log         LoggingConfig
	BugSnag     *BugSnagConfig
	Metrics     metriks.Config
	Tracing     tracing.Config
	FeatureFlag featureflag.Config
}

func (args *RootArgs) Setup(config interface{}, serviceName, version string) (logrus.FieldLogger, error) {
	rootConfig, err := args.loadDefaultConfig()
	if err != nil {
		return nil, err
	}

	log, err := ConfigureLogging(rootConfig.Log)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create the logger")
	}
	if version == "" {
		version = "unknown"
	}
	log = log.WithField("version", version)

	if err := SetupBugSnag(rootConfig.BugSnag, version); err != nil {
		return nil, errors.Wrap(err, "Failed to configure bugsnag")
	}

	if err := metriks.Init(serviceName, rootConfig.Metrics); err != nil {
		return nil, errors.Wrap(err, "Failed to configure metrics")
	}

	// Handles the 'enabled' flag itself
	tracing.Configure(&rootConfig.Tracing, log, serviceName)

	if err := featureflag.Init(rootConfig.FeatureFlag, log); err != nil {
		return nil, errors.Wrap(err, "Failed to configure featureflags")
	}

	if err := sendDatadogEvents(rootConfig.Metrics, serviceName, version); err != nil {
		log.WithError(err).Error("Failed to send the startup events to datadog")
	}

	if config != nil {
		// second load the config for this project
		if err := args.load(config); err != nil {
			return log, errors.Wrap(err, "Failed to load the config object")
		}
		log.Debug("Loaded configuration")
	}
	return log, nil
}

func (args *RootArgs) load(cfg interface{}) error {
	loader := func(cfg interface{}) error {
		return LoadFromEnv(args.Prefix, args.ConfigFile, cfg)
	}
	if !strings.HasSuffix(args.ConfigFile, ".env") {
		loader = func(cfg interface{}) error {
			return LoadConfigFromFile(args.ConfigFile, cfg)
		}
	}
	return loader(cfg)
}

func (args *RootArgs) MustSetup(config interface{}, serviceName, version string) logrus.FieldLogger {
	logger, err := args.Setup(config, serviceName, version)
	if err != nil {
		if logger != nil {
			logger.WithError(err).Fatal("Failed to setup configuration")
		} else {
			panic(fmt.Sprintf("Failed to setup configuration: %s", err.Error()))
		}
	}

	return logger
}

func (args *RootArgs) loadDefaultConfig() (*RootConfig, error) {
	c := &RootConfig{
		Log: DefaultLoggingConfig(),
	}

	if err := args.load(c); err != nil {
		return nil, errors.Wrap(err, "Failed to load the logging configuration")
	}
	return c, nil
}

func (args *RootArgs) AddFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().AddFlag(args.ConfigFlag())
	cmd.Flags().AddFlag(args.PrefixFlag())
	return cmd
}

func (args *RootArgs) ConfigFlag() *pflag.Flag {
	return &pflag.Flag{
		Name:      "config",
		Shorthand: "c",
		Usage:     "A file to load configuration from, supported formats are env, json, and yaml",
		Value:     newStringValue("", &args.ConfigFile),
	}
}

func (args *RootArgs) PrefixFlag() *pflag.Flag {
	return &pflag.Flag{
		Name:      "prefix",
		Shorthand: "p",
		Usage:     "A prefix to search for when looking for env vars",
		Value:     newStringValue("", &args.Prefix),
	}
}

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}
func (s *stringValue) Type() string   { return "string" }
func (s *stringValue) String() string { return string(*s) }
