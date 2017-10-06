package nconf

import (
	"os"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/pkg/errors"
	"github.com/shopify/logrus-bugsnag"
	"github.com/sirupsen/logrus"
)

type LoggingConfig struct {
	Level   string `mapstructure:"log_level" json:"log_level"`
	File    string `mapstructure:"log_file" json:"log_file"`
	BugSnag *BugSnagConfig
}

type BugSnagConfig struct {
	Environment string
	APIKey      string `envconfig:"api_key"`
}

func ConfigureLogging(config *LoggingConfig) (*logrus.Entry, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// always use the full timestamp
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	})

	// use a file if you want
	if config.File != "" {
		f, errOpen := os.OpenFile(config.File, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0664)
		if errOpen != nil {
			return nil, errOpen
		}
		logrus.SetOutput(f)
		logrus.Infof("Set output file to %s", config.File)
	}

	if config.Level != "" {
		level, err := logrus.ParseLevel(config.Level)
		if err != nil {
			return nil, err
		}
		logrus.SetLevel(level)
		logrus.Debug("Set log level to: " + logrus.GetLevel().String())
	}

	log := logrus.StandardLogger()
	if err := AddBugSnagHook(log, config.BugSnag); err != nil {
		return nil, errors.Wrap(err, "Failed to configure bugsnag")
	}

	return log.WithField("hostname", hostname), nil
}

func AddBugSnagHook(log *logrus.Logger, config *BugSnagConfig) error {
	if config == nil || config.APIKey == "" {
		return nil
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       config.APIKey,
		ReleaseStage: config.Environment,
	})
	hook, err := logrus_bugsnag.NewBugsnagHook()
	if err != nil {
		return err
	}
	log.Hooks.Add(hook)
	log.Debug("Added bugsnag hook")
	return nil
}
