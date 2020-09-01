package kafka

import (
	"context"
	"fmt"
	"log/syslog"
	"strings"
	"time"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type PartitionerAlgorithm string

// Supported auth types
const (
	AuthTypePlain    = "plain"
	AuthTypeSCRAM256 = "scram-sha256"
	AuthTypeSCRAM512 = "scram-sha512"

	PartitionerRandom           = PartitionerAlgorithm("random")            // random distribution
	PartitionerConsistent       = PartitionerAlgorithm("consistent")        //  CRC32 hash of key (Empty and NULL keys are mapped to single partition)
	PartitionerConsistentRandom = PartitionerAlgorithm("consistent_random") // CRC32 hash of key (Empty and NULL keys are randomly partitioned)
	PartitionerMurMur2          = PartitionerAlgorithm("murmur2")           // Java Producer compatible Murmur2 hash of key (NULL keys are mapped to single partition)
	PartitionerMurMur2Random    = PartitionerAlgorithm("murmur2_random")    // Java Producer compatible Murmur2 hash of key (NULL keys are randomly partitioned. Default partitioner in the Java Producer.)
	PartitionerFNV1A            = PartitionerAlgorithm("fnv1a")             // FNV-1a hash of key (NULL keys are mapped to single partition)
	PartitionerFNV1ARandom      = PartitionerAlgorithm("fnv1a_random")      // FNV-1a hash of key (NULL keys are randomly partitioned).

	DefaultTimout = time.Duration(30 * time.Second) // Default timout to be used if not set in the config
)

// DefaultLogLevel is the log level Kafka producers/consumers will use if non set.
const DefaultLogLevel = logrus.ErrorLevel

// Config holds all the configuration for this package.
type Config struct {
	Brokers        []string       `json:"brokers"`
	Topic          string         `json:"topic"`
	Producer       ProducerConfig `json:"producer"`
	Consumer       ConsumerConfig `json:"consumer"`
	AuthType       string         `json:"auth" split_words:"true"`
	User           string         `json:"user"`
	Password       string         `json:"password"`
	CAPEMFile      string         `json:"ca_pem_file"`
	LogLevel       string         `json:"log_level" split_words:"true"`
	RequestTimeout time.Duration  `json:"request_timeout"`
}

// baseKafkaConfig provides the base config that applies to both consumers and producers.
func (c Config) baseKafkaConfig() *kafkalib.ConfigMap {
	logrusToSylogLevelMapping := map[logrus.Level]syslog.Priority{
		logrus.PanicLevel: syslog.LOG_EMERG, // Skipping LOG_ALERT, LOG_CRIT. LOG_EMERG has highest priority.
		logrus.FatalLevel: syslog.LOG_EMERG, // Skipping LOG_ALERT, LOG_CRIT. LOG_EMERG has highest priority.
		logrus.ErrorLevel: syslog.LOG_ERR,
		logrus.WarnLevel:  syslog.LOG_WARNING,
		logrus.InfoLevel:  syslog.LOG_NOTICE, // Skipping LOG_INFO. LOG_NOTICE has highest priority.
		logrus.DebugLevel: syslog.LOG_DEBUG,
		logrus.TraceLevel: syslog.LOG_DEBUG,
	}

	logLevel, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		logLevel = DefaultLogLevel
	}

	// See Reference at https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	kafkaConf := &kafkalib.ConfigMap{
		"bootstrap.servers":       strings.Join(c.Brokers, ","),
		"socket.keepalive.enable": true,
		"log_level":               int(logrusToSylogLevelMapping[logLevel]),
	}

	if logLevel == logrus.DebugLevel {
		_ = kafkaConf.SetKey("debug", "consumer,broker,topic,msg")
	}

	return kafkaConf
}

// ConsumerConfig holds the specific configuration for a consumer.
type ConsumerConfig struct {
	GroupID              string               `json:"group_id" split_words:"true"`
	Partition            *int32               `json:"partition"`
	PartitionKey         string               `json:"partition_key"`
	PartitionerAlgorithm PartitionerAlgorithm `json:"partition_algorithm"`
	InitialOffset        *int64               `json:"initial_offset"`
}

// Apply applies the specific configuration for a consumer.
func (c ConsumerConfig) Apply(kafkaConf *kafkalib.ConfigMap) {
	if id := c.GroupID; id != "" {
		_ = kafkaConf.SetKey("group.id", id)
	}
}

// ProducerConfig holds the specific configuration for a producer.
type ProducerConfig struct {
	FlushPeriod     time.Duration `json:"flush_period" split_words:"true"`
	BatchSize       int           `json:"batch_size" split_words:"true"`
	DeliveryTimeout time.Duration `json:"delivery_timeout" split_words:"true"`
}

// Apply applies the specific configuration for a producer.
func (c ProducerConfig) Apply(kafkaConf *kafkalib.ConfigMap) {
	if timeout := c.DeliveryTimeout; timeout > 0 {
		_ = kafkaConf.SetKey("delivery.timeout.ms", int(timeout.Milliseconds()))
	}

	if size := c.BatchSize; size > 0 {
		_ = kafkaConf.SetKey("queue.buffering.max.messages", size)
	}

	if period := c.FlushPeriod; period > 0 {
		_ = kafkaConf.SetKey("queue.buffering.max.ms", int(period.Milliseconds()))
	}
}

func (c Config) configureAuth(configMap *kafkalib.ConfigMap) error {
	switch c.AuthType {
	case "":
		// No auth mechanism
		return nil
	case AuthTypePlain:
		_ = configMap.SetKey("security.protocol", "sasl_plain")
		_ = configMap.SetKey("sasl.mechanism", "PLAIN")
		_ = configMap.SetKey("sasl.username", c.User)
		_ = configMap.SetKey("sasl.password", c.Password)
	case AuthTypeSCRAM256:
		_ = configMap.SetKey("security.protocol", "sasl_ssl")
		_ = configMap.SetKey("sasl.mechanism", "SCRAM-SHA-256")
		_ = configMap.SetKey("sasl.username", c.User)
		_ = configMap.SetKey("sasl.password", c.Password)
	case AuthTypeSCRAM512:
		_ = configMap.SetKey("security.protocol", "sasl_ssl")
		_ = configMap.SetKey("sasl.mechanism", "SCRAM-SHA-512")
		_ = configMap.SetKey("sasl.username", c.User)
		_ = configMap.SetKey("sasl.password", c.Password)
	default:
		return fmt.Errorf("unknown auth type: %s", c.AuthType)
	}

	if c.CAPEMFile != "" {
		_ = configMap.SetKey("ssl.ca.location", c.CAPEMFile)
	}

	return nil
}

// ConfigOpt configures Kafka consumers and producers.
type ConfigOpt func(c *kafkalib.ConfigMap)

// WithLogger adds a logger to a Kafka consumer or producer.
func WithLogger(ctx context.Context, log logrus.FieldLogger) ConfigOpt {
	return func(c *kafkalib.ConfigMap) {

		syslogToLogrusLevelMapping := map[syslog.Priority]logrus.Level{
			// We don't want to let the app to panic so considering Error Level as the highest severity.
			syslog.LOG_EMERG:   logrus.ErrorLevel,
			syslog.LOG_ALERT:   logrus.ErrorLevel,
			syslog.LOG_CRIT:    logrus.ErrorLevel,
			syslog.LOG_ERR:     logrus.ErrorLevel,
			syslog.LOG_WARNING: logrus.WarnLevel,
			syslog.LOG_NOTICE:  logrus.InfoLevel,
			syslog.LOG_INFO:    logrus.InfoLevel,
			syslog.LOG_DEBUG:   logrus.DebugLevel,
		}

		// Forward logs to a channel.
		logsChan := make(chan kafkalib.LogEvent, 10000)
		_ = c.SetKey("go.logs.channel.enable", true)
		_ = c.SetKey("go.logs.channel", logsChan)

		// Read from channel and print logs using the provided logger.
		go func() {
			defer close(logsChan)
			for {
				select {
				case <-ctx.Done():
					return
				case m, ok := <-logsChan:
					if !ok {
						return
					}
					l := log.WithFields(logrus.Fields{
						"kafka_context": m.Tag,
						"kafka_client":  m.Name,
					}).WithTime(m.Timestamp)

					logrusLevel := syslogToLogrusLevelMapping[syslog.Priority(m.Level)]
					switch logrusLevel {
					case logrus.ErrorLevel:
						l.WithError(errors.New(m.Message)).Error("Error in Kafka Consumer")
					default:
						l.Log(logrusLevel, m.Message)
					}
				}
			}
		}()
	}
}

// WithConsumerGroupID sets the Consumer consumer group ID.
func WithConsumerGroupID(groupID string) ConfigOpt {
	return func(c *kafkalib.ConfigMap) {
		_ = c.SetKey("group.id", groupID)
	}
}

// WithPartitionerAlgorithm sets the partitioner algorithm
func WithPartitionerAlgorithm(algorithm PartitionerAlgorithm) ConfigOpt {
	return func(c *kafkalib.ConfigMap) {
		_ = c.SetKey("partitioner", string(algorithm))
	}
}
