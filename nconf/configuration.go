package nconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// ErrUnknownConfigFormat indicates the extension of the config file is not supported as a config source
type ErrUnknownConfigFormat struct {
	ext string
}

func (e *ErrUnknownConfigFormat) Error() string {
	return fmt.Sprintf("Unknown config format: %s", e.ext)
}

// LoadFromFile will load the configuration from the specified file based on the file type
// There is only support for .json and .yml now
func LoadFromFile(configFile string, input interface{}) error {
	if configFile == "" {
		return nil
	}

	// read in all the bytes
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	config := make(map[string]interface{})

	configExt := filepath.Ext(configFile)

	switch configExt {
	case ".json":
		err = json.Unmarshal(data, &config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	default:
		err = &ErrUnknownConfigFormat{configExt}
	}
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if err := mapstructure.Decode(&config, input); err != nil {
		return fmt.Errorf("failed to map data: %w", err)
	}
	return nil
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
