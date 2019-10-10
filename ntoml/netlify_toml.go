package ntoml

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"
)

const DefaultFilename = "netlify.toml"

// cf. https://github.com/netlify/build/blob/3c9cf4dda7a39994a3f0f1a544242d386b2bc2dd/packages/%40netlify-config/path.js#L16
var netlifyConfigFileNames = []string{
	"netlify.toml", "netlify.yml", "netlify.yaml", "netlify.json",
}

type NetlifyToml struct {
	Settings Settings `toml:"settings" json:"settings" yaml:"settings"`

	Redirects []Redirect `toml:"redirects,omitempty" json:"redirects,omitempty" yaml:"redirects,omitempty"`

	// this is the default context
	Build   *BuildConfig             `toml:"build" json:"build" yaml:"build"`
	Context map[string]DeployContext `toml:"context,omitempty" json:"context,omitempty" yaml:"context,omitempty"`
}

type Settings struct {
	ID   string `toml:"id" json:"id" yaml:"id"`
	Path string `toml:"path" json:"path" yaml:"path"`
}

type BuildConfig struct {
	Command     string            `toml:"command" json:"command" yaml:"command"`
	Base        string            `toml:"base" json:"base" yaml:"base"`
	Publish     string            `toml:"publish" json:"publish" yaml:"publish"`
	Environment map[string]string `toml:"environment" json:"environment" yaml:"environment"`
}

type DeployContext struct {
	BuildConfig `yaml:",inline"`
}

type Redirect struct {
	Origin      string             `toml:"origin" json:"origin" yaml:"origin"`
	Destination string             `toml:"destination" json:"destination" yaml:"destination"`
	Parmeters   map[string]string  `toml:"parameters" json:"parameters" yaml:"parameters"`
	Status      int                `toml:"status" json:"status" yaml:"status"`
	Force       bool               `toml:"force" json:"force" yaml:"force"`
	Conditions  *RedirectCondition `toml:"conditions" json:"conditions" yaml:"conditions"`
	Headers     map[string]string  `toml:"headers" json:"headers" yaml:"headers"`
}

type RedirectCondition struct {
	Language []string `toml:"language" json:"language" yaml:"language"`
	Country  []string `toml:"country" json:"country" yaml:"country"`
	Role     []string `toml:"role" json:"role" yaml:"role"`
}

type FoundNoConfigPathError struct {
	checked []string
}

func (f *FoundNoConfigPathError) Error() string {
	return fmt.Sprintf("none of the checked paths exist. checked = %v", f.checked)
}

type FoundMoreThanOneConfigPathError struct {
	checked []string
	found   []string
}

func (f *FoundMoreThanOneConfigPathError) Error() string {
	return fmt.Sprintf("more than one of the checked paths exist. checked = %v, found = %v", f.checked, f.found)
}

func findOnlyOneExistingPath(base string, paths ...string) (path string, err error) {
	joinedPaths := make([]string, 0, len(paths))
	foundPaths := make([]string, 0, len(paths))
	for _, possiblePath := range paths {
		p := filepath.Join(base, possiblePath)
		joinedPaths = append(joinedPaths, p)
		if fi, err := os.Stat(p); !os.IsNotExist(err) && !fi.IsDir() {
			foundPaths = append(foundPaths, p)
		}
	}
	if len(foundPaths) == 0 {
		return "", &FoundNoConfigPathError{checked: paths}
	}
	if len(foundPaths) > 1 {
		return "", &FoundMoreThanOneConfigPathError{checked: paths, found: foundPaths}
	}
	return foundPaths[0], nil
}

func GetNetlifyConfigPath(base string) (path string, err error) {
	return findOnlyOneExistingPath(base, netlifyConfigFileNames...)
}

func Load() (*NetlifyToml, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	configPath, err := GetNetlifyConfigPath(wd)
	if err != nil {
		return nil, err
	}
	return LoadFrom(configPath)
}

func LoadFrom(paths ...string) (*NetlifyToml, error) {
	if len(paths) == 0 {
		return nil, errors.New("No paths specified")
	}

	out := new(NetlifyToml)

	for _, p := range paths {
		extension := filepath.Ext(p)

		if data, ferr := ioutil.ReadFile(p); !os.IsNotExist(ferr) {
			if ferr != nil {
				return nil, errors.Wrapf(ferr, "Error while reading in file %s", p)
			}

			var derr error

			switch extension {
			case ".toml":
				derr = toml.Unmarshal(data, out)
			case ".json":
				derr = json.Unmarshal(data, out)
			case ".yaml":
				fallthrough
			case ".yml":
				derr = yaml.Unmarshal(data, out)
			default:
				return nil, errors.New(fmt.Sprintf("Invalid config extension %s of path %s", extension, p))
			}

			if derr != nil {
				return nil, errors.Wrapf(derr, "Error while decoding file %s", p)
			}

			return out, nil
		}
	}
	return nil, os.ErrNotExist
}

func Save(conf *NetlifyToml) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return SaveTo(conf, path.Join(wd, DefaultFilename))
}

func SaveTo(conf *NetlifyToml, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to open file %s", path)
	}

	defer f.Close()

	if err := toml.NewEncoder(f).Encode(conf); err != nil {
		return errors.Wrap(err, "Failed to encode the toml file")
	}

	return nil
}
