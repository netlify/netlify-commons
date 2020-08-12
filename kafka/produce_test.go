package kafka

import (
	"context"
	"testing"
	"time"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
)

func TestProducerProduce(t *testing.T) {
	p := producer(t)
	defer checkClose(t, p)

	topic := "gotest"
	msg := &kafkalib.Message{
		TopicPartition: kafkalib.TopicPartition{Topic: &topic},
		Key:            []byte("gotest"),
		Value:          []byte("gotest"),
		Timestamp:      time.Now(),
	}
	err := p.Produce(context.Background(), msg)

	// Expected error since there are no brokers.
	require.EqualError(t, err, "failed to produce a kafka message: Local: Message timed out")
}

func producer(t *testing.T) *ConfluentProducer {
	p, err := NewProducer(Config{}, func(configMap *kafkalib.ConfigMap) {
		_ = configMap.SetKey("queue.buffering.max.ms", 1)
		_ = configMap.SetKey("delivery.timeout.ms", 2)
	})
	require.NoError(t, err)

	return p
}
