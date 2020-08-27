package kafka

import (
	"context"
	"fmt"
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
	// FetchMessage fetches one message, if there is any available at the current offset.
	FetchMessage(ctx context.Context) (*kafkalib.Message, error)

	// Close closes the consumer.
	Close() error

	// CommitMessage commits the offset of a given message.
	CommitMessage(msg *kafkalib.Message) error

	// GetMetadata gets the metadata for a consumer.
	GetMetadata(allTopics bool) (*kafkalib.Metadata, error)

	// GetPartitions returns the partitions on the consumer.
	GetPartions() ([]int32, error)

	// Seek seeks the assigned topic partitions to the given offset.
	Seek(offset int64) error

	// SeekToTime seeks to the specified time.
	SeekToTime(t time.Time) error

	// SetPartittionByKey sets the current consumer to read from a partion by a hashed key.
	SetPartitionByKey(key string, algorithm PartitionerAlgorithm) error

	// SetPartitionByID sets the current consumer to read from the specified partition.
	SetPartitionByID(id int32) error
}

// ConfluentConsumer implements Consumer interface.
type ConfluentConsumer struct {
	c    *kafkalib.Consumer
	conf Config
	log  logrus.FieldLogger

	rebalanceHandler      func(c *kafkalib.Consumer, ev kafkalib.Event) error // Only set when an initial offset should be set
	rebalanceHandlerMutex sync.Mutex

	eventChan chan kafkalib.Event
}

// NewConsumer creates a ConfluentConsumer based on config.
// - NOTE if the partition is set and the partition key is not set in config we have no way
//   of knowing where to assign the consumer to in the case of a rebalance
func NewConsumer(log logrus.FieldLogger, conf Config, opts ...ConfigOpt) (Consumer, error) {
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

	if conf.Timeout == 0 {
		conf.Timeout = DefaultTimout
	}

	cc := &ConfluentConsumer{
		c:    consumer,
		conf: conf,
		log:  log,
	}

	if cc.conf.Consumer.Partition == nil {
		// if we are not reading from a specific partition, we subscribe
		cc.setupRebalanceHandler(cc.conf.Consumer.InitialOffset)
		cc.log.WithField("kafka_topic", cc.conf.Topic).Debug("Subscribing to Kafka topic")
		cc.rebalanceHandlerMutex.Lock()
		defer cc.rebalanceHandlerMutex.Unlock()
		if serr := cc.c.Subscribe(cc.conf.Topic, cc.rebalanceHandler); serr != nil {
			err = errors.Wrap(serr, "error subscribing to topic")
		}
	} else {
		// if we are reading from a specific partition, we assign
		tps := []kafkalib.TopicPartition{
			kafkalib.TopicPartition{
				Topic:     &cc.conf.Topic,
				Partition: *cc.conf.Consumer.Partition,
			},
		}
		// Set the partition if a key is set to determine the partition
		if cc.conf.Consumer.PartitionKey != "" && cc.conf.Consumer.PartitionerAlgorithm != "" {
			cc.SetPartitionByKey(cc.conf.Consumer.PartitionKey, cc.conf.Consumer.PartitionerAlgorithm)
		}
		err = cc.c.Assign(tps)

		// handle rebalance events for consumer assigned to a specific partition
		// - NOTE if the partition is set and the partition key is not set we have no way
		//   of knowing where to assign the consumer to in the case of a rebalance
		cc.eventChan = cc.c.Events()
		go func(cc *ConfluentConsumer) {
			for ev := range cc.eventChan {
				log := cc.log.WithField("kafka_event", ev.String())
				switch e := ev.(type) {
				case kafkalib.RevokedPartitions:
					// check if we are assigned to this partition
					revokedParts := e.Partitions
					revoked := false
					for _, part := range revokedParts {
						if part.Partition == *cc.conf.Consumer.Partition && *part.Topic == cc.conf.Topic {
							revoked = true
							break
						}
					}
					if revoked {
						cc.log.WithField("kafka_event", e.String()).Debug("Unassigning Kafka partitions after rebalance")
						if err := cc.c.Unassign(); err != nil {
							log.WithError(err).Error("failed unassigning current Kafka partitions after rebalance")
						}
						// if we know the partition key we can reassign
						if cc.conf.Consumer.PartitionKey != "" && cc.conf.Consumer.PartitionerAlgorithm != "" {
							cc.SetPartitionByKey(cc.conf.Consumer.PartitionKey, cc.conf.Consumer.PartitionerAlgorithm)
						}
					}
				}
			}
		}(cc)
	}
	if err != nil {
		return nil, err
	}

	return cc, nil
}

// Seek seeks the assigned topic partitions to the given offset.
func (r *ConfluentConsumer) Seek(offset int64) error {
	tp := kafkalib.TopicPartition{Topic: &r.conf.Topic, Offset: kafkalib.Offset(offset)}
	if r.conf.Consumer.Partition != nil {
		tp.Partition = *r.conf.Consumer.Partition
	}

	err := r.c.Seek(tp, int(r.conf.Timeout.Milliseconds()))
	if err, ok := err.(kafkalib.Error); ok && err.Code() == kafkalib.ErrTimedOut {
		return ErrSeekTimedOut
	}

	return nil
}

// SeekToTime seeks to the specified time.
func (r *ConfluentConsumer) SeekToTime(t time.Time) error {
	var offsets []kafkalib.TopicPartition
	millisSinceEpoch := t.UnixNano() / 1000000
	tps := []kafkalib.TopicPartition{{Topic: &r.conf.Topic, Offset: kafkalib.Offset(millisSinceEpoch)}}
	if r.conf.Consumer.Partition != nil {
		tps[0].Partition = *r.conf.Consumer.Partition
	}
	offsets, err := r.c.OffsetsForTimes(tps, int(r.conf.Timeout.Milliseconds()))
	if err != nil {
		return err
	}
	if len(offsets) == 1 {
		return r.Seek(int64(offsets[0].Offset))
	}

	return fmt.Errorf("error finding offset to seek to")
}

// setupReabalnceHandler does the setup of the rebalance handler
func (r *ConfluentConsumer) setupRebalanceHandler(offset int64) error {
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

// FetchMessage fetches one message, if there is any available at the current offset.
func (r *ConfluentConsumer) FetchMessage(ctx context.Context) (*kafkalib.Message, error) {
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

// CommitMessage commits the offset of a given message.
func (r *ConfluentConsumer) CommitMessage(msg *kafkalib.Message) error {
	_, err := r.c.CommitMessage(msg)
	return errors.Wrap(err, "failed committing Kafka message")
}

// Close closes the consumer.
func (r *ConfluentConsumer) Close() error {
	return r.c.Close()
}

// GetMetadata return the confluence consumer metatdata
func (r *ConfluentConsumer) GetMetadata(allTopics bool) (*kafkalib.Metadata, error) {
	if allTopics {
		return r.c.GetMetadata(nil, true, int(r.conf.Timeout.Milliseconds()))
	}

	return r.c.GetMetadata(&r.conf.Topic, false, int(r.conf.Timeout.Milliseconds()))
}

// GetPartions returns the partition ids of the configured topic
func (r *ConfluentConsumer) GetPartions() ([]int32, error) {
	meta, err := r.GetMetadata(false)
	if err != nil {
		return nil, err
	}

	return getPartitionIds(r.conf.Topic, meta)
}

// SetPartitionByKey sets the partition to consume messages from by the passed key and algorithm
// - NOTE we currently only support the murmur2 hashing algorithm in the consumer
func (r *ConfluentConsumer) SetPartitionByKey(key string, algorithm PartitionerAlgorithm) error {
	if algorithm != PartitionerMurMur2 {
		return fmt.Errorf("we currently only support the murmur2 hashing algorithm in the consumer")
	}
	parts, err := r.GetPartions()
	if err != nil {
		return err
	}

	return r.SetPartitionByID(GetPartition(key, parts))
}

// SetPartitionByID sets the partition to consume messages from by the passed partition ID
func (r *ConfluentConsumer) SetPartitionByID(id int32) error {
	parts, err := r.GetPartions()
	if err != nil {
		return err
	}
	found := false
	for _, part := range parts {
		if part == id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("%d is not a valid partition id", id)
	}
	r.conf.Consumer.Partition = &id
	pt := []kafkalib.TopicPartition{
		kafkalib.TopicPartition{
			Topic:     &r.conf.Topic,
			Partition: *r.conf.Consumer.Partition,
		},
	}
	err = r.c.Assign(pt)

	r.log.WithField("kafka_partition_id", id).Debug("Assigning Kafka partition")

	return nil
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
