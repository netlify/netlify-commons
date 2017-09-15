package nconf

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func LoadFromEnv(prefix, filename string, face interface{}) error {
	var err error
	if filename == "" {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(filename)
	}
	if err != nil {
		return err
	}

	return envconfig.Process(prefix, face)
}
