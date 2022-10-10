package instrument

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

type MockClient struct {
	Logger logrus.FieldLogger
}

var _ Client = MockClient{}

func (c MockClient) Identify(userID string, traits analytics.Traits) {
	c.Logger.WithFields(logrus.Fields{
		"user_id": userID,
		"traits":  traits,
	}).Info("Received Identify event")
}

func (c MockClient) Track(userID string, event string, properties analytics.Properties) {
	c.Logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"event":      event,
		"properties": properties,
	}).Info("Received Track event")
}

func (c MockClient) Page(userID string, name string, properties analytics.Properties) {
	c.Logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"name":       name,
		"properties": properties,
	}).Info("Received Page event")
}

func (c MockClient) Group(userID string, groupID string, traits analytics.Traits) {
	c.Logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"group_id": groupID,
		"traits":   traits,
	}).Info("Received Group event")
}

func (c MockClient) Alias(previousID string, userID string) {
	c.Logger.WithFields(logrus.Fields{
		"previous_id": previousID,
		"user_id":     userID,
	}).Info("Received Alias event")
}
