package nconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// LoadConfig loads the config from a file if specified, otherwise from the environment
func LoadConfig(serviceName string, input interface{}, possiblePaths ...string) error {
	paths := []string{}
	for _, p := range possiblePaths {
		paths = ifExists(paths, p)
	}
	paths = ifExists(paths, ".", ".env")
	paths = ifExists(paths, ".", serviceName+".env")
	paths = ifExists(paths, ".", "env")
	paths = ifExists(paths, "$HOME", ".netlify", serviceName, "env")

	if len(paths) > 0 {
		if err := godotenv.Load(paths...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to load env files: %s", strings.Join(paths, ",")))
		}
	}

	if err := envconfig.Process(serviceName, input); err != nil {
		return errors.Wrap(err, "Failed to unmarshal environment vars")
	}
	return nil
}

func ifExists(paths []string, parts ...string) []string {
	path := filepath.Join(parts...)
	if _, err := os.Stat(path); err == nil {
		return append(paths, path)
	}
	return paths
}

func WriteEnvFile(path string, kv map[string]interface{}) error {
	var lines string
	for k, v := range kv {
		lines += fmt.Sprintf("%s=%v\n", strings.ToUpper(k), v)
	}

	return ioutil.WriteFile(path, []byte(lines), 0644)
}
