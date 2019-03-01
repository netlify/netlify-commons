package nconf

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LoadConfig loads the config from a file if specified, otherwise from the environment
func LoadConfig(cmd *cobra.Command, serviceName string, input interface{}) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	configFile, _ := cmd.Flags().GetString("config")
	return LoadConfigWithFile(serviceName, configFile, input)
}

// LoadConfigWithFile loads the service configuration from an optional file, otherwise from the environment
func LoadConfigWithFile(serviceName, configFile string, input interface{}) error {
	viper.SetEnvPrefix(serviceName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if configFile != "" {
		viper.SetConfigFile(configFile)

		if ext := filepath.Ext(configFile); len(ext) > 1 {
			switch strings.ToLower(ext[1:]) {
			case "yaml", "yml":
				viper.SetConfigType("yaml")
			case "json":
				fallthrough
			default:
				viper.SetConfigType("json")
			}
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./")
		viper.AddConfigPath("$HOME/.netlify/" + serviceName + "/")
	}

	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			return errors.Wrap(err, "reading configuration from files")
		}
	}

	if err := viper.Unmarshal(input); err != nil {
		return err
	}
	return nil
}
