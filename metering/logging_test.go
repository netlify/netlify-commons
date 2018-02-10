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

	assert.NotNil(t, global.buffer)
	assert.NotNil(t, global.encoder)
	assert.NotNil(t, global.writelock)
}

func TestInitNoFile(t *testing.T) {
	require.NoError(t, os.Setenv("METERING_FILENAME", ""))
	initFromEnv()
	assert.NotNil(t, global)
}

func TestWriteDataAppendFile(t *testing.T) {
	f, err := ioutil.TempFile("", "metering-test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	ogcontents := []byte("this is a test\n")
	f.Write(ogcontents)

	ut, err := NewMeteringLog(f.Name())
	require.NoError(t, err)

	assert.NotNil(t, ut.buffer)
	assert.NotNil(t, ut.encoder)
	assert.NotNil(t, ut.writelock)

	startContents, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, ogcontents, startContents)

	require.NoError(t, ut.RecordEvent("someevent", map[string]interface{}{"somekey": 1}))
	require.NoError(t, ut.Flush())

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

func TestSeparationOfGlobal(t *testing.T) {
	d, err := ioutil.TempDir("", "metering-test-")
	require.NoError(t, err)
	defer os.RemoveAll(d)
	f1 := filepath.Join(d, "file1")
	f2 := filepath.Join(d, "file2")

	require.NoError(t, SetMeteringFile(f1))
	ut, err := NewMeteringLog(f2)
	require.NoError(t, err)

	RecordEvent("first", nil)
	assert.NoError(t, Global().RecordEvent("second", nil))
	assert.NoError(t, ut.RecordEvent("third", nil))

	assert.NoError(t, global.Flush())
	assert.NoError(t, ut.Flush())

	file1, err := os.Open(f1)
	require.NoError(t, err)
	require.NoError(t, err)
	s1 := bufio.NewScanner(file1)
	count := 0
	for s1.Scan() {
		count++
		read := new(MeteringEvent)
		assert.NoError(t, json.Unmarshal(s1.Bytes(), read))
		switch read.Event {
		case "first", "second":
		default:
			assert.Fail(t, "Unexpected event in the global file", read.Event)
		}
	}
	assert.Equal(t, 2, count)

	file2, err := os.Open(f2)
	s1 = bufio.NewScanner(file2)
	for s1.Scan() {
		count++
		read := new(MeteringEvent)
		assert.NoError(t, json.Unmarshal(s1.Bytes(), read))
		assert.Equal(t, "third", read.Event)
	}
	assert.Equal(t, 3, count)
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
