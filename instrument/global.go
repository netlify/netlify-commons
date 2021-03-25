package instrument

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

// Init will initialize global client with a segment client
func Init(conf Config, log logrus.FieldLogger) error {
	segmentClient, err := NewClient(&conf, log)
	if err != nil {
		return err
	}
	SetGlobalClient(segmentClient)
	return nil
}
