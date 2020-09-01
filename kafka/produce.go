package kafka

import (
	"context"
	"fmt"
	"io"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

const (
	// DefaultProducerDeliveryTimeoutMs configures `delivery.timeout.ms`. The timeout for the producer from sending a message until is considered as delivered.
	// This value should always be greater than DefaultProducerBufferMaxMs.
	// The default value in librdkafka is `300000`, but we reduced it to `5000`.
	DefaultProducerDeliveryTimeoutMs = 5000

	// DefaultProducerBufferMaxMs configures `queue.buffering.max.ms`. The max amount of ms the buffer will wait before sending it to kafka.
	// This value should always be lower than DefaultProducerDeliveryTimeoutMs.
	// The default value in librdkafka is `5`.
	DefaultProducerBufferMaxMs = 5

	// DefaultProducerBufferMaxMessages configures `queue.buffering.max.messages`. The max number of messages in buffer before sending to Kafka.
	// The default value in librdkafka is `100000`.
	DefaultProducerBufferMaxMessages = 100000
)

// Producer produces messages into Kafka.
type Producer interface {
	io.Closer
	Produce(ctx context.Context, msgs ...*kafkalib.Message) error
}

// ConfluentProducer implements Producer interface.
type ConfluentProducer struct {
	p    *kafkalib.Producer
	conf Config
}

// NewProducer creates a ConfluentProducer based on config.
func NewProducer(conf Config, opts ...ConfigOpt) (w *ConfluentProducer, err error) {
	// See Reference at https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	kafkaConf := conf.baseKafkaConfig()
	_ = kafkaConf.SetKey("delivery.timeout.ms", DefaultProducerDeliveryTimeoutMs)
	_ = kafkaConf.SetKey("queue.buffering.max.messages", DefaultProducerBufferMaxMessages)
	_ = kafkaConf.SetKey("queue.buffering.max.ms", DefaultProducerBufferMaxMs)

	conf.Producer.Apply(kafkaConf)
	for _, opt := range opts {
		opt(kafkaConf)
	}

	if err := conf.configureAuth(kafkaConf); err != nil {
		return nil, errors.Wrap(err, "error configuring auth for the Kafka producer")
	}

	// catch when NewProducer panics
	defer func() {
		if r := recover(); r != nil {
			w = nil
			err = fmt.Errorf("failed to create producer: %s", r)
		}
	}()

	p, err := kafkalib.NewProducer(kafkaConf)
	if err != nil {
		return nil, err
	}

	if conf.RequestTimeout == 0 {
		conf.RequestTimeout = DefaultTimout
	}

	return &ConfluentProducer{p: p, conf: conf}, nil
}

// Close should be called when no more writes will be performed.
func (w ConfluentProducer) Close() error {
	w.p.Close()
	return nil
}

// Produce produces messages into Kafka.
func (w ConfluentProducer) Produce(ctx context.Context, msgs ...*kafkalib.Message) error {
	deliveryChan := make(chan kafkalib.Event, 1)
	defer close(deliveryChan)
	for _, m := range msgs {
		if m.TopicPartition.Topic == nil {
			m.TopicPartition.Topic = &w.conf.Topic
		}

		if m.TopicPartition.Partition <= 0 {
			m.TopicPartition.Partition = kafkalib.PartitionAny
		}

		if err := w.p.Produce(m, deliveryChan); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case e, ok := <-deliveryChan:
			if !ok {
				return nil
			}

			m, ok := e.(*kafkalib.Message)
			if !ok {
				// This should not happen.
				return errors.New("Producer delivery channel received a Kafka.Event but was not a kafka.Message")
			}

			if m.TopicPartition.Error != nil {
				return errors.Wrap(m.TopicPartition.Error, "failed to produce a kafka message")
			}

			break
		}
	}

	return nil
}

// GetMetadata return the confluence producers metatdata
func (w *ConfluentProducer) GetMetadata(allTopics bool) (*kafkalib.Metadata, error) {
	if allTopics {
		return w.p.GetMetadata(nil, true, int(w.conf.RequestTimeout.Milliseconds()))
	}

	return w.p.GetMetadata(&w.conf.Topic, false, int(w.conf.RequestTimeout.Milliseconds()))
}

// GetPartions returns the partition ids of a given topic
func (w *ConfluentProducer) GetPartions() ([]int32, error) {
	meta, err := w.GetMetadata(false)
	if err != nil {
		return nil, err
	}

	return getPartitionIds(w.conf.Topic, meta)
}
