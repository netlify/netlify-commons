package nconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/netlify/netlify-commons/featureflag"
	"github.com/netlify/netlify-commons/metriks"
	"github.com/netlify/netlify-commons/tracing"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// ErrUnknownConfigFormat indicates the extension of the config file is not supported as a config source
type ErrUnknownConfigFormat struct {
	ext string
}

func (e *ErrUnknownConfigFormat) Error() string {
	return fmt.Sprintf("Unknown config format: %s", e.ext)
}

type RootConfig struct {
	Log         LoggingConfig
	BugSnag     *BugSnagConfig
	Metrics     metriks.Config
	Tracing     tracing.Config
	FeatureFlag featureflag.Config
}

func DefaultConfig() RootConfig {
	return RootConfig{
		Log: LoggingConfig{
			QuoteEmptyFields: true,
		},
		Tracing: tracing.Config{
			Host: "localhost",
			Port: "8126",
		},
		Metrics: metriks.Config{
			Host: "localhost",
			Port: 8125,
		},
	}
}

/*
 Deprecated: This method relies on parsing the json/yaml to a map, then running it through mapstructure.
 This required that both tags exist (annoying!). And so there is now LoadConfigFromFile.
*/
// LoadFromFile will load the configuration from the specified file based on the file type
// There is only support for .json and .yml now
func LoadFromFile(configFile string, input interface{}) error {
	if configFile == "" {
		return nil
	}

	switch {
	case strings.HasSuffix(configFile, ".json"):
		viper.SetConfigType("json")
	case strings.HasSuffix(configFile, ".yaml"):
		fallthrough
	case strings.HasSuffix(configFile, ".yml"):
		viper.SetConfigType("yaml")
	}
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			return errors.Wrap(err, "reading configuration from files")
		}
	}

	return viper.Unmarshal(input)
}

// LoadConfigFromFile will load the configuration from the specified file based on the file type
// There is only support for .json and .yml now. It will use the underlying json/yaml packages directly.
// meaning those should be the only required tags.
func LoadConfigFromFile(configFile string, input interface{}) error {
	if configFile == "" {
		return nil
	}

	// read in all the bytes
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	configExt := filepath.Ext(configFile)
	switch configExt {
	case ".json":
		return json.Unmarshal(data, input)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, input)
	}
	return &ErrUnknownConfigFormat{configExt}
}

func LoadFromEnv(prefix, filename string, face interface{}) error {
	var err error
	if filename == "" {
		err = godotenv.Load()
		if os.IsNotExist(err) {
			err = nil
		}
	} else {
		err = godotenv.Load(filename)
	}

	if err != nil {
		return err
	}

	return envconfig.Process(prefix, face)
}
