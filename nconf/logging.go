package nconf

import (
	"os"
	"time"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/pkg/errors"
	"github.com/shopify/logrus-bugsnag"
	"github.com/sirupsen/logrus"
)

type LoggingConfig struct {
	Level            string `mapstructure:"log_level" json:"log_level"`
	File             string `mapstructure:"log_file" json:"log_file"`
	DisableColors    bool   `mapstructure:"disable_colors" json:"disable_colors"`
	QuoteEmptyFields bool   `mapstructure:"quote_empty_fields" json:"quote_empty_fields"`
	TSFormat         string `mapstructure:"ts_format" json:"ts_format"`
	BugSnag          *BugSnagConfig
	Fields           map[string]interface{} `mapstructure:"fields" json:"fields"`
}

type BugSnagConfig struct {
	Environment string
	APIKey      string `envconfig:"api_key"`
}

func ConfigureLogging(config *LoggingConfig) (*logrus.Entry, error) {
	tsFormat := time.RFC3339Nano
	if config.TSFormat != "" {
		tsFormat = config.TSFormat
	}
	// always use the full timestamp
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
		TimestampFormat:  tsFormat,
		DisableColors:    config.DisableColors,
		QuoteEmptyFields: config.QuoteEmptyFields,
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

	if err := AddBugSnagHook(config.BugSnag); err != nil {
		return nil, errors.Wrap(err, "Failed to configure bugsnag")
	}

	f := logrus.Fields{}
	for k, v := range config.Fields {
		f[k] = v
	}

	return logrus.WithFields(f), nil
}

func AddBugSnagHook(config *BugSnagConfig) error {
	if config == nil || config.APIKey == "" {
		return nil
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       config.APIKey,
		ReleaseStage: config.Environment,
		PanicHandler: func() {}, // this is to disable panic handling. The lib was forking and restarting the process (causing races)
	})
	hook, err := logrus_bugsnag.NewBugsnagHook()
	if err != nil {
		return err
	}
	logrus.AddHook(hook)
	logrus.Debug("Added bugsnag hook")
	return nil
}
