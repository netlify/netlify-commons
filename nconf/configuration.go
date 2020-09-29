package nconf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

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

	switch {
	case strings.HasSuffix(configFile, ".json"):
		err = json.Unmarshal(data, input)
	case strings.HasSuffix(configFile, ".yaml"):
		fallthrough
	case strings.HasSuffix(configFile, ".yml"):
		err = yaml.Unmarshal(data, input)
	}
	return err
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
