package instrument

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

type Client interface {
	Identify(userID string, traits analytics.Traits)
	Track(userID string, event string, properties analytics.Properties)
	Page(userID string, name string, properties analytics.Properties)
	Group(userID string, groupID string, traits analytics.Traits)
	Alias(previousID string, userID string)
}

type segmentClient struct {
	analytics.Client
	log logrus.FieldLogger
}

var _ Client = &segmentClient{}

func NewClient(cfg *Config, logger logrus.FieldLogger) (Client, error) {
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

func (c segmentClient) Identify(userID string, traits analytics.Traits) {
	err := c.Client.Enqueue(analytics.Identify{
		UserId: userID,
		Traits: traits,
	})
	if err != nil {
		c.log.WithError(err).Info("Failed to send instrument Identify metrics")
	}
}

func (c segmentClient) Track(userID string, event string, properties analytics.Properties) {
	err := c.Client.Enqueue(analytics.Track{
		UserId:     userID,
		Event:      event,
		Properties: properties,
	})
	if err != nil {
		c.log.WithError(err).Info("Failed to send instrument Track metrics")
	}
}

func (c segmentClient) Page(userID string, name string, properties analytics.Properties) {
	err := c.Client.Enqueue(analytics.Page{
		UserId:     userID,
		Name:       name,
		Properties: properties,
	})
	if err != nil {
		c.log.WithError(err).Info("Failed to send instrument Page metrics")
	}
}

func (c segmentClient) Group(userID string, groupID string, traits analytics.Traits) {
	err := c.Client.Enqueue(analytics.Group{
		UserId:  userID,
		GroupId: groupID,
		Traits:  traits,
	})
	if err != nil {
		c.log.WithError(err).Info("Failed to send instrument Group metrics")
	}
}

func (c segmentClient) Alias(previousID string, userID string) {
	err := c.Client.Enqueue(analytics.Alias{
		PreviousId: previousID,
		UserId:     userID,
	})
	if err != nil {
		c.log.WithError(err).Info("Failed to send instrument Alias metrics")
	}
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
