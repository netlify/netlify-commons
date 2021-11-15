package featureflag

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var globalLock sync.Mutex
var globalClient Client = MockClient{}

func SetGlobalClient(client Client) {
	if client == nil {
		return
	}
	globalLock.Lock()
	globalClient = client
	globalLock.Unlock()
}

func GetGlobalClient() Client {
	globalLock.Lock()
	defer globalLock.Unlock()
	return globalClient
}

// Init will initialize global client with a launch darkly client
func Init(conf Config, log logrus.FieldLogger, defaultAttrs ...Attr) error {
	ldClient, err := NewClient(&conf, log, defaultAttrs...)
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
