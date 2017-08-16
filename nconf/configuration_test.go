package nconf

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnvVar(t *testing.T) {
	c := struct {
		TopVal string
		Nested struct {
			StringVal string `envconfig:"string"`
			NumVal    int    `split_words:"true"`
			EmptyVal  string
		}
		Pointer *struct {
			StringVal string `required:"true"`
			NumVal    int
		} `envconfig:"ptr"`
		Missing *struct {
			StringVal string
		}
	}{}

	os.Setenv("TEST_SVC_TOPVAL", "envval1")
	os.Setenv("TEST_SVC_NESTED_STRING", "s1")
	os.Setenv("TEST_SVC_NESTED_NUM_VAL", "1")
	os.Setenv("TEST_SVC_PTR_STRINGVAL", "s2")

	// write a env file too
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	err = WriteEnvFile(f.Name(), map[string]interface{}{
		"test_svc_top_val":    "envval2",
		"test_svc_PTR_NUMVAL": 10,
	})
	require.NoError(t, err)

	// cc := WithConfigFlag(new(cobra.Command))
	require.NoError(t, LoadConfig("test_svc", &c, f.Name()))

	assert.Equal(t, "envval1", c.TopVal) // env takes precedence
	assert.Equal(t, "s1", c.Nested.StringVal)
	assert.Equal(t, "", c.Nested.EmptyVal)
	assert.Equal(t, 1, c.Nested.NumVal)

	assert.NotNil(t, c.Pointer)
	assert.Equal(t, "s2", c.Pointer.StringVal)
	assert.Equal(t, 10, c.Pointer.NumVal)

	// it will *always* populate pointer structs
	assert.NotNil(t, c.Missing)
	assert.Equal(t, "", c.Missing.StringVal)
}
