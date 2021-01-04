package nconf

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

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
