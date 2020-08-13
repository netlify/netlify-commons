package kafka

import (
	"context"
	"io"
	"sync"
	"time"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrSeekTimedOut is the error returned when a consumer timed out during Seek.
var ErrSeekTimedOut = errors.New("Kafka Seek timed out. Please try again.")

// Consumer reads messages from Kafka.
type Consumer interface {
	io.Closer

	// FetchMessage fetches one message, if there is any available at the current offset.
	FetchMessage(ctx context.Context) (*kafkalib.Message, error)

	// CommitMessage commits the offset of a given message.
	CommitMessage(msg *kafkalib.Message) error
}

// OffsetAwareConsumer is a Consumer that can reset its offset.
type OffsetAwareConsumer interface {
	Consumer

	// SetInitialOffset resets the current offset to the given one.
	// Used for setting the initial offset a consumer should start consuming from.
	// Should be called before start consuming messages.
	SetInitialOffset(offset int64) error

	// Seek seeks the assigned topic partitions to the given offset.
	// Seek() may only be used for partitions already being consumed.
	Seek(offset int64, timeout time.Duration) error
}

// ConfluentConsumer implements Consumer interface.
type ConfluentConsumer struct {
	c    *kafkalib.Consumer
	conf Config
	log  logrus.FieldLogger

	rebalanceHandler      func(c *kafkalib.Consumer, ev kafkalib.Event) error // Only set when an initial offset should be set
	rebalanceHandlerMutex sync.Mutex
	subscribeOnce         sync.Once
}

// NewConsumer creates a ConfluentConsumer based on config.
func NewConsumer(log logrus.FieldLogger, conf Config, opts ...ConfigOpt) (*ConfluentConsumer, error) {
	// See Reference at https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	kafkaConf := conf.baseKafkaConfig()
	_ = kafkaConf.SetKey("enable.auto.offset.store", false) // manually StoreOffset after processing a message. Otherwise races may happen.)

	// In case we try to assign an offset out of range (greater than log-end-offset), consumer will use start consuming from offset zero.
	_ = kafkaConf.SetKey("auto.offset.reset", "earliest")

	conf.Consumer.Apply(kafkaConf)
	for _, opt := range opts {
		opt(kafkaConf)
	}

	if err := conf.configureAuth(kafkaConf); err != nil {
		return nil, errors.Wrap(err, "error configuring auth for the Kafka consumer")
	}

	consumer, err := kafkalib.NewConsumer(kafkaConf)
	if err != nil {
		return nil, err
	}

	confluenceConsumer := &ConfluentConsumer{
		c:    consumer,
		conf: conf,
		log:  log,
	}

	return confluenceConsumer, nil
}

// Seek implements OffsetAwareConsumer interface.
func (r *ConfluentConsumer) Seek(offset int64, timeout time.Duration) error {
	timeoutMs := timeout.Milliseconds()
	if timeoutMs == 0 {
		// Otherwise the call will be asynchronous, losing error handling.
		return errors.New("Timeout should be set when calling Seek")
	}

	err := r.c.Seek(kafkalib.TopicPartition{Topic: &r.conf.Topic, Offset: kafkalib.Offset(offset)}, int(timeoutMs))
	if err, ok := err.(kafkalib.Error); ok && err.Code() == kafkalib.ErrTimedOut {
		return ErrSeekTimedOut
	}

	return err
}

// SetInitialOffset implements OffsetAwareConsumer interface.
func (r *ConfluentConsumer) SetInitialOffset(offset int64) error {
	r.rebalanceHandlerMutex.Lock()
	defer r.rebalanceHandlerMutex.Unlock()

	// Setting the following rebalance handler ensures the offset is set right after a rebalance, avoiding
	// connectivity problems caused by race conditions (consumer did not join the group yet).
	// Once set, the responsibility of assigning/unassigning partitions after a rebalance happens is moved to our app.
	// This mechanism is the recommended one by confluent-kafka-go creators. Since our consumers are tied to consumer groups,
	// the Subscribe() method should be called eventually, which will trigger a rebalance. Otherwise, if the consumer would
	// not be a member of a group, we could just use Assign() with the hardcoded partitions instead, but this is not the case.
	// See https://docs.confluent.io/current/clients/confluent-kafka-go/index.html#hdr-High_level_Consumer
	var once sync.Once
	r.rebalanceHandler = func(c *kafkalib.Consumer, ev kafkalib.Event) error {
		log := r.log.WithField("kafka_event", ev.String())
		switch e := ev.(type) {
		case kafkalib.AssignedPartitions:
			partitions := e.Partitions
			once.Do(func() {
				log.WithField("kafka_offset", offset).Debug("Skipping Kafka assignment given by coordinator after rebalance in favor of resetting the offset")
				partitions = kafkalib.TopicPartitions{{Topic: &r.conf.Topic, Offset: kafkalib.Offset(offset)}}
			})

			log.WithField("kafka_partitions", partitions).Debug("Assigning Kafka partitions after rebalance")
			if err := c.Assign(partitions); err != nil {
				log.WithField("kafka_partitions", partitions).WithError(err).Error("failed assigning Kafka partitions after rebalance")
				return err
			}
		case kafkalib.RevokedPartitions:
			r.log.WithField("kafka_event", e.String()).Debug("Unassigning Kafka partitions after rebalance")
			if err := c.Unassign(); err != nil {
				log.WithError(err).Error("failed unassigning current Kafka partitions after rebalance")
				return err
			}
		}
		return nil
	}

	return nil
}

// FetchMessage implements Consumer interface.
func (r *ConfluentConsumer) FetchMessage(ctx context.Context) (*kafkalib.Message, error) {
	var err error
	r.subscribeOnce.Do(func() {
		r.log.WithField("kafka_topic", r.conf.Topic).Debug("Subscribing to Kafka topic")
		r.rebalanceHandlerMutex.Lock()
		defer r.rebalanceHandlerMutex.Unlock()
		if err := r.c.Subscribe(r.conf.Topic, r.rebalanceHandler); err != nil {
			err = errors.Wrap(err, "error subscribing to topic")
		}
	})
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// The timeout applies for the poll time, meaning if no messages during 5 min, read will timeout.
			// Used for checking <-ctx.Done() from time to time.
			msg, err := r.c.ReadMessage(time.Minute * 5)
			if err != nil {
				if err.(kafkalib.Error).Code() == kafkalib.ErrTimedOut {
					// Avoid logging errors when timing out.
					continue
				}

				if err := handleConfluentReadMessageError(r.log, err, "failed fetching Kafka message"); err != nil {
					return nil, err
				}

				// a backoff is take in place inside librdkafka, so next call to consume will wait until that backoff.
				// `fetch.error.backoff.ms` defaults to 500ms
				continue
			}
			return msg, nil
		}
	}
}

// CommitMessage implements Consumer interface.
func (r *ConfluentConsumer) CommitMessage(msg *kafkalib.Message) error {
	_, err := r.c.CommitMessage(msg)
	return errors.Wrap(err, "failed committing Kafka message")
}

// Close should be called when no more readings will be performed.
func (r *ConfluentConsumer) Close() error {
	return r.c.Close()
}

// handleConfluentReadMessageError returns an error if the error is fatal.
// confluent-kafka-go manages most of the errors internally except for fatal errors which are non-recoverable.
// Non fatal errors will be just ignored (just logged)
// See https://github.com/edenhill/librdkafka/blob/master/src/rdkafka_request.h#L35-L45
func handleConfluentReadMessageError(log logrus.FieldLogger, originalErr error, msg string) error {
	if originalErr == nil {
		return nil
	}

	err, ok := originalErr.(kafkalib.Error)
	if !ok {
		return nil
	}

	log = log.WithError(err).WithField("kafka_err_fatal", err.IsFatal())
	if err.IsFatal() {
		log.Errorf("%s. No retry will take place.", msg)
		return err
	}

	log.WithError(err).Errorf("%s. A retry will take place.", msg)
	return nil
}