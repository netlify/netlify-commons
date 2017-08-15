package nconf

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnvVar(t *testing.T) {
	c := struct {
		TopVal string `mapstructure:"top_val"`
		Nested struct {
			StringVal string `mapstructure:"string_val"`
			NumVal    int    `mapstructure:"num_val"`
			EmptyVal  string `mapstructure:"empty_val"`
		} `mapstructure:"nested"`
		Pointer *struct {
			StringVal string `mapstructure:"string_val"`
		}
		Missing *struct {
			StringVal string `mapstructure:"string_val"`
		}
	}{}

	os.Setenv("TEST_SVC_TOP_VAL", "envval1")
	os.Setenv("TEST_SVC_NESTED_STRING_VAL", "s1")
	os.Setenv("TEST_SVC_NESTED_NUM_VAL", "1")
	os.Setenv("TEST_SVC_POINTER_STRING_VAL", "s2")

	err := LoadConfig(new(cobra.Command), "test_svc", &c)
	require.NoError(t, err)

	assert.Equal(t, "envval1", c.TopVal)
	assert.Equal(t, "s1", c.Nested.StringVal)
	assert.Equal(t, "", c.Nested.EmptyVal)
	assert.Equal(t, 1, c.Nested.NumVal)

	assert.NotNil(t, c.Pointer)
	assert.Equal(t, "s2", c.Pointer.StringVal)

	assert.Nil(t, c.Missing)
}
