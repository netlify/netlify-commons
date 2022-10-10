package ntoml

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

const DefaultFilename = "netlify.toml"

type NetlifyToml struct {
	Settings Settings `toml:"settings"`

	Redirects []Redirect `toml:"redirects,omitempty"`

	// this is the default context
	Build   *BuildConfig             `toml:"build"`
	Plugins []Plugin                 `toml:"plugins"`
	Context map[string]DeployContext `toml:"context,omitempty"`
}

type Settings struct {
	ID   string `toml:"id"`
	Path string `toml:"path"`
}

type BuildConfig struct {
	Command      string            `toml:"command"`
	Base         string            `toml:"base"`
	Publish      string            `toml:"publish"`
	Ignore       string            `toml:"ignore"`
	Environment  map[string]string `toml:"environment"`
	Functions    string            `toml:"functions"`
	EdgeHandlers string            `toml:"edge_handlers"`
}

type Plugin struct {
	Package       string `toml:"package" json:"package"`
	PinnedVersion string `toml:"pinned_version,omitempty" json:"pinned_version,omitempty"`
}

type DeployContext struct {
	BuildConfig `yaml:",inline"`
}

type Redirect struct {
	Origin      string             `toml:"origin"`
	Destination string             `toml:"destination"`
	Parmeters   map[string]string  `toml:"parameters"`
	Status      int                `toml:"status"`
	Force       bool               `toml:"force"`
	Conditions  *RedirectCondition `toml:"conditions"`
	Headers     map[string]string  `toml:"headers"`
}

type RedirectCondition struct {
	Language []string `toml:"language"`
	Country  []string `toml:"country"`
	Role     []string `toml:"role"`
}

type FoundNoConfigPathError struct {
	base    string
	checked string
}

func (f *FoundNoConfigPathError) Error() string {
	return fmt.Sprintf("No Netlify configuration file found.")
}

func GetNetlifyConfigPath(base string) (path string, err error) {
	filePath := filepath.Join(base, DefaultFilename)

	if fi, err := os.Stat(filePath); err == nil && !fi.IsDir() {
		return filePath, nil
	} else {
		return "", &FoundNoConfigPathError{base: base, checked: DefaultFilename}
	}
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
		if data, ferr := ioutil.ReadFile(p); !os.IsNotExist(ferr) {
			if ferr != nil {
				return nil, errors.Wrapf(ferr, "Error while reading in file %s", p)
			}

			if derr := toml.Unmarshal(data, out); derr != nil {
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
