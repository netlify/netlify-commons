package messaging

import (
	"time"

	"github.com/netlify/netlify-commons/nconf"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func AddNatsLogHook(nc NatsWriter, config *nconf.NatsConfig, log *logrus.Entry) error {
	if config.LogsSubject != "" {
		lvls := []logrus.Level{}
		if len(config.LogLevels) > 0 {
			for _, lstr := range config.LogLevels {
				lvl, err := logrus.ParseLevel(lstr)
				if err != nil {
					return errors.Wrapf(err, "Failed to parse '%s' into a level", lstr)
				}
				lvls = append(lvls, lvl)
			}
		}
		hook := NewNatsHook(nc, config.LogsSubject, lvls)
		log.Logger.Hooks.Add(hook)
		log.Debugf("Added NATS hook to send logs to %s", config.LogsSubject)
	}
	return nil
}

type NatsWriter interface {
	Publish(string, []byte) error
}

// NatsHook will emit logs to the subject provided
type NatsHook struct {
	conn          NatsWriter
	subject       string
	extraFields   map[string]interface{}
	dynamicFields map[string]func() interface{}
	Formatter     logrus.Formatter

	LogLevels []logrus.Level
}

// NewNatsHook will create a logrus hook that will automatically send
// new info into the channel
func NewNatsHook(conn NatsWriter, subject string, levels []logrus.Level) *NatsHook {
	hook := NatsHook{
		conn:          conn,
		subject:       subject,
		extraFields:   make(map[string]interface{}),
		dynamicFields: make(map[string]func() interface{}),
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		},
		LogLevels: levels,
	}

	if len(hook.LogLevels) == 0 {
		hook.LogLevels = logrus.AllLevels
	}

	return &hook
}

// AddField will add a simple value each emission
func (hook *NatsHook) AddField(key string, value interface{}) *NatsHook {
	hook.extraFields[key] = value
	return hook
}

// AddDynamicField will call that method on each fire
func (hook *NatsHook) AddDynamicField(key string, generator func() interface{}) *NatsHook {
	hook.dynamicFields[key] = generator
	return hook
}

// Fire will use the connection and try to send the message to the right destination
func (hook *NatsHook) Fire(entry *logrus.Entry) error {
	// add in the new fields
	for k, v := range hook.extraFields {
		entry.Data[k] = v
	}

	for k, generator := range hook.dynamicFields {
		entry.Data[k] = generator()
	}

	bytes, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}

	return hook.conn.Publish(hook.subject, bytes)
}

// Levels will describe what levels the NatsHook is associated with
func (hook *NatsHook) Levels() []logrus.Level {
	return hook.LogLevels
}
