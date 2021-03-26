package instrument

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

type Client interface {
	identify(userID string, traits analytics.Traits) error
	track(userID string, event string, properties analytics.Properties) error
	page(userID string, name string, properties analytics.Properties) error
	group(userID string, groupID string, traits analytics.Traits) error
	alias(previousID string, userID string) error
}

type segmentClient struct {
	analytics.Client
	log logrus.FieldLogger
}

var _ Client = &segmentClient{}

func newClient(cfg *Config, logger logrus.FieldLogger) (Client, error) {
	config := analytics.Config{}

	if !cfg.Enabled {
		// use mockClient instead
		return &MockClient{logger}, nil
	}

	configureLogger(&config, logger)

	inner, err := analytics.NewWithConfig(cfg.Key, config)
	if err != nil {
		logger.WithError(err).Error("Unable to construct Segment client")
	}
	return &segmentClient{inner, logger}, err
}

func (c segmentClient) identify(userID string, traits analytics.Traits) error {
	return c.Client.Enqueue(analytics.Identify{
		UserId: userID,
		Traits: traits,
	})
}

func (c segmentClient) track(userID string, event string, properties analytics.Properties) error {
	return c.Client.Enqueue(analytics.Track{
		UserId:     userID,
		Event:      event,
		Properties: properties,
	})
}

func (c segmentClient) page(userID string, name string, properties analytics.Properties) error {
	return c.Client.Enqueue(analytics.Page{
		UserId:     userID,
		Name:       name,
		Properties: properties,
	})
}

func (c segmentClient) group(userID string, groupID string, traits analytics.Traits) error {
	return c.Client.Enqueue(analytics.Group{
		UserId:  userID,
		GroupId: groupID,
		Traits:  traits,
	})
}

func (c segmentClient) alias(previousID string, userID string) error {
	return c.Client.Enqueue(analytics.Alias{
		PreviousId: previousID,
		UserId:     userID,
	})
}

func configureLogger(conf *analytics.Config, log logrus.FieldLogger) {
	if log == nil {
		l := logrus.New()
		l.SetOutput(ioutil.Discard)
		log = l
	}
	log = log.WithField("component", "segment")
	conf.Logger = &wrapLog{log.Printf, log.Errorf}
}

type wrapLog struct {
	printf func(format string, args ...interface{})
	errorf func(format string, args ...interface{})
}

// Logf implements analytics.Logger interface
func (l *wrapLog) Logf(format string, args ...interface{}) {
	l.printf(format, args...)
}

// Errorf implements analytics.Logger interface
func (l *wrapLog) Errorf(format string, args ...interface{}) {
	l.errorf(format, args...)
}
