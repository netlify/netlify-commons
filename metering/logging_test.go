package metering

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	f, err := ioutil.TempFile("", "metering-test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	require.NoError(t, os.Setenv("METERING_FILENAME", f.Name()))

	initFromEnv()
	assert.NotNil(t, logger)
	assert.NotNil(t, meteringBuffer)
	assert.NotNil(t, encoder)
}

func TestWriteDataAppendFile(t *testing.T) {
	f, err := ioutil.TempFile("", "metering-test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	ogcontents := []byte("this is a test\n")
	f.Write(ogcontents)

	require.NoError(t, SetMeteringFile(f.Name()))

	startContents, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, ogcontents, startContents)

	RecordEvent("someevent", map[string]interface{}{"somekey": 1})
	Flush()
	endContents, err := ioutil.ReadFile(f.Name())
	require.NoError(t, err)
	assert.Equal(t, ogcontents, endContents[0:len(ogcontents)])

	jsonStr := endContents[len(ogcontents):]
	me := new(MeteringEvent)
	require.NoError(t, json.Unmarshal(jsonStr, me))
	assert.Equal(t, "someevent", me.Event)
	assert.Len(t, me.Data, 1)
	assert.EqualValues(t, 1, me.Data["somekey"])
	assert.WithinDuration(t, time.Now(), time.Unix(0, me.Timestamp), time.Second*5)
}

func TestWriteDataCreateFile(t *testing.T) {
	d, err := ioutil.TempDir("", "metering-test-")
	require.NoError(t, err)
	defer os.RemoveAll(d)

	path := filepath.Join(d, "test-file")
	require.NoError(t, SetMeteringFile(path))

	_, err = os.Open(path)
	assert.NoError(t, err)
}

func TestConcurrentWrites(t *testing.T) {
	f, err := ioutil.TempFile("", "metering-test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	require.NoError(t, SetMeteringFile(f.Name()))
	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			RecordEvent("someevent", map[string]interface{}{"somekey": i})
			wg.Done()
		}()
	}
	wg.Wait()
	Flush()

	lines := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines++
		m := make(map[string]interface{})
		if assert.NoError(t, json.Unmarshal(sc.Bytes(), &m)) {
			assert.Equal(t, "someevent", m["event"].(string))
		}
	}

	assert.Equal(t, 100, lines)
}
