package featureflag

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"unsafe"
)

// See https://blog.dubbelboer.com/2015/08/23/rwmutex-vs-atomicvalue-vs-unsafepointer.html
var (
	defaultClient Client = MockClient{}
	globalClient         = atomic.NewUnsafePointer(unsafe.Pointer(&defaultClient))
)

func SetGlobalClient(client Client) {
	if client == nil {
		return
	}
	globalClient.Store(unsafe.Pointer(&client))
}

func GetGlobalClient() Client {
	c := (*Client)(globalClient.Load())
	return *c
}

// Init will initialize global client with a launch darkly client
func Init(conf Config, log logrus.FieldLogger) error {
	ldClient, err := NewClient(&conf, log)
	if err != nil {
		return err
	}
	SetGlobalClient(ldClient)
	return nil
}

func Enabled(key, userID string, attrs ...Attr) bool {
	return GetGlobalClient().Enabled(key, userID, attrs...)
}

func Variation(key, defaultVal, userID string, attrs ...Attr) string {
	return GetGlobalClient().Variation(key, defaultVal, userID, attrs...)
}

func Int(key string, defaultVal int, userID string, attrs ...Attr) int {
	return GetGlobalClient().Int(key, defaultVal, userID, attrs...)
}
