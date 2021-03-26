package instrument

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

type MockClient struct {
	Logger logrus.FieldLogger
}

var _ Client = MockClient{}

func (c MockClient) Identify(userID string, traits analytics.Traits) error {
	c.Logger.WithFields(logrus.Fields{
		"user_id": userID,
		"traits":  traits,
	}).Infof("Received Identity event")
	return nil
}

func (c MockClient) Track(userID string, event string, properties analytics.Properties) error {
	c.Logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"event":      event,
		"properties": properties,
	}).Infof("Received Track event")
	return nil
}

func (c MockClient) Page(userID string, name string, properties analytics.Properties) error {
	c.Logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"name":       name,
		"properties": properties,
	}).Infof("Received Page event")
	return nil
}

func (c MockClient) Group(userID string, groupID string, traits analytics.Traits) error {
	c.Logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"group_id": groupID,
		"traits":   traits,
	}).Infof("Received Group event")
	return nil
}

func (c MockClient) Alias(previousID string, userID string) error {
	c.Logger.WithFields(logrus.Fields{
		"previous_id": previousID,
		"user_id":     userID,
	}).Infof("Received Alias event")
	return nil
}
