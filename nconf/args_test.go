package nconf

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestArgsLoad(t *testing.T) {
	cfg := &struct {
		Something  string
		Other      int
		Overridden string
	}{
		Something:  "default",
		Overridden: "this should change",
	}

	tmp, err := ioutil.TempFile("", "something*.env")
	require.NoError(t, err)
	cfgStr := `
PF_OTHER=10
PF_OVERRIDDEN=not-that
PF_LOG_LEVEL=debug
PF_LOG_QUOTE_EMPTY_FIELDS=true
`
	require.NoError(t, ioutil.WriteFile(tmp.Name(), []byte(cfgStr), 0644))

	args := RootArgs{
		Prefix:     "pf",
		ConfigFile: tmp.Name(),
	}

	log, err := args.Setup(cfg, "", "")
	require.NoError(t, err)

	// check that we did call configure the logger
	assert.NotNil(t, log)
	entry := log.(*logrus.Entry)
	assert.Equal(t, logrus.DebugLevel, entry.Logger.Level)
	assert.True(t, entry.Logger.Formatter.(*logrus.TextFormatter).QuoteEmptyFields)

	assert.Equal(t, "default", cfg.Something)
	assert.Equal(t, 10, cfg.Other)
	assert.Equal(t, "not-that", cfg.Overridden)
}

func TestArgsAddToCmd(t *testing.T) {
	args := new(RootArgs)
	var called int
	cmd := cobra.Command{
		Run: func(_ *cobra.Command, _ []string) {
			assert.Equal(t, "PF", args.Prefix)
			assert.Equal(t, "file.env", args.ConfigFile)
			called++
		},
	}
	cmd.PersistentFlags().AddFlag(args.ConfigFlag())
	cmd.PersistentFlags().AddFlag(args.PrefixFlag())
	cmd.SetArgs([]string{"--config", "file.env", "--prefix", "PF"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 1, called)
}

func TestArgsLoadFromYAML(t *testing.T) {
	f, err := ioutil.TempFile("", "test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	args := RootArgs{
		ConfigFile: f.Name(),
	}

	t.Run("empty-file", func(t *testing.T) {
		cfg := RootConfig{
			Log: DefaultLoggingConfig(),
		}
		require.NoError(t, args.load(&cfg))

		assert.True(t, cfg.Log.QuoteEmptyFields)
		assert.Equal(t, DefaultLoggingConfig(), cfg.Log)
		assert.Nil(t, cfg.BugSnag)
	})

	_, err = f.WriteString(`
log:
  level: debug
  fields:
    string: value
    int: 4
`)
	require.NoError(t, err)

	t.Run("set log field", func(t *testing.T) {
		cfg := RootConfig{Log: DefaultLoggingConfig()}
		require.NoError(t, args.load(&cfg))

		// retains original value
		assert.True(t, cfg.Log.QuoteEmptyFields)
		assert.Equal(t, "debug", cfg.Log.Level)
		require.Len(t, cfg.Log.Fields, 2)
		assert.Equal(t, "value", cfg.Log.Fields["string"])
		assert.Equal(t, 4, cfg.Log.Fields["int"])
	})
}
