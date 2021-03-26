package instrument

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

var globalLock sync.Mutex
var globalClient client = MockClient{}

func setGlobalClient(client client) {
	if client == nil {
		return
	}
	globalLock.Lock()
	globalClient = client
	globalLock.Unlock()
}

func getGlobalClient() client {
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
	setGlobalClient(segmentClient)
	return nil
}

func Identify(userID string, traits analytics.Traits) error {
	return getGlobalClient().identify(userID, traits)
}

func Track(userID string, event string, properties analytics.Properties) error {
	return getGlobalClient().track(userID, event, properties)
}

func Page(userID string, name string, properties analytics.Properties) error {
	return getGlobalClient().page(userID, name, properties)
}

func Group(userID string, groupID string, traits analytics.Traits) error {
	return getGlobalClient().group(userID, groupID, traits)
}

func Alias(previousID string, userID string) error {
	return getGlobalClient().alias(previousID, userID)
}
