package nconf

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LoadConfig loads the config from a file if specified, otherwise from the environment
func LoadConfig(cmd *cobra.Command, serviceName string, input interface{}) error {
	viper.SetConfigType("json")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	viper.SetEnvPrefix(serviceName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if configFile, _ := cmd.Flags().GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
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
