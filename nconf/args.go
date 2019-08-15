package nconf

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type RootArgs struct {
	Prefix  string
	EnvFile string
}

func (args *RootArgs) Setup(config interface{}, version string) logrus.FieldLogger {
	// first load the logger
	logConfig := &struct {
		Log *LoggingConfig
	}{}
	if err := LoadFromEnv(args.Prefix, args.EnvFile, logConfig); err != nil {
		logrus.WithError(err).Fatal("Failed to load the logging configuration")
	}

	log, err := ConfigureLogging(logConfig.Log)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create the logger")
	}
	if version == "" {
		version = "unknown"
	}
	log = log.WithField("version", version)

	if config != nil {
		// second load the config for this project
		if err := LoadFromEnv(args.Prefix, args.EnvFile, config); err != nil {
			log.WithError(err).Fatal("Failed to load the config object")
		}
		log.Debug("Loaded configuration")
	}
	return log
}

func (args *RootArgs) ConfigFlag() *pflag.Flag {
	return &pflag.Flag{
		Name:      "config",
		Shorthand: "c",
		Usage:     "A .env file to load configuration from",
		Value:     newStringValue("", &args.EnvFile),
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
