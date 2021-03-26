package instrument

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
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

// Identify sends an identify type message to a queue to be sent to Segment.
func Identify(userID string, traits analytics.Traits) error {
	return GetGlobalClient().Identify(userID, traits)
}

// Track sends a track type message to a queue to be sent to Segment.
func Track(userID string, event string, properties analytics.Properties) error {
	return GetGlobalClient().Track(userID, event, properties)
}

// Page sends a page type message to a queue to be sent to Segment.
func Page(userID string, name string, properties analytics.Properties) error {
	return GetGlobalClient().Page(userID, name, properties)
}

// Group sends a group type message to a queue to be sent to Segment.
func Group(userID string, groupID string, traits analytics.Traits) error {
	return GetGlobalClient().Group(userID, groupID, traits)
}

// Alias sends an alias type message to a queue to be sent to Segment.
func Alias(previousID string, userID string) error {
	return GetGlobalClient().Alias(previousID, userID)
}
