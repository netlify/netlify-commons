package ntoml

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

const DefaultFilename = "netlify.toml"

type NetlifyToml struct {
	Settings Settings `toml:"settings"`

	Redirects []Redirect `toml:"redirects, omitempty"`

	// this is the default context
	Build   *BuildConfig             `toml:"build"`
	Context map[string]DeployContext `toml:"context, omitempty"`
}

type Settings struct {
	ID   string `toml:"id"`
	Path string `toml:"path"`
}

type BuildConfig struct {
	Command     string            `toml:"command"`
	Base        string            `toml:"base"`
	Publish     string            `toml:"publish"`
	Environment map[string]string `toml:"environment"`
}

type DeployContext struct {
	BuildConfig
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

func Load() (*NetlifyToml, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path.Join(wd, DefaultFilename))
}

func LoadFrom(paths ...string) (*NetlifyToml, error) {
	if len(paths) == 0 {
		return nil, errors.New("No paths specified")
	}

	out := new(NetlifyToml)

	for _, p := range paths {
		if data, ferr := ioutil.ReadFile(p); !os.IsNotExist(ferr) {
			if ferr != nil {
				return nil, errors.Wrapf(ferr, "Error while reading in file %s.", p)
			}

			if _, derr := toml.Decode(string(data), out); derr != nil {
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
