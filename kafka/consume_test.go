package kafka

import (
	"context"
	"testing"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
)

func TestConsumerSetInitialOffset(t *testing.T) {
	c, _ := consumer(t)
	defer checkClose(t, c)

	require.Nil(t, c.rebalanceHandler)
	require.NoError(t, c.SetInitialOffset(1))
	require.NotNil(t, c.rebalanceHandler)
}

func TestConsumerFetchMessageContextAwareness(t *testing.T) {
	c, _ := consumer(t)
	defer checkClose(t, c)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // explicitly cancelling the context

	msg, err := c.FetchMessage(ctx)
	require.Nil(t, msg)
	require.EqualError(t, err, context.Canceled.Error())
}

func TestConsumerSeek(t *testing.T) {
	c, conf := consumer(t)
	defer checkClose(t, c)
	require.NoError(t, c.c.Assign(kafkalib.TopicPartitions{{Topic: &conf.Topic, Partition: 0}})) // manually assign partition
	require.NoError(t, c.Seek(2))
}

func consumer(t *testing.T) (*ConfluentConsumer, Config) {
	conf := Config{
		Brokers: nil, // No brokers are used for unit test.
		Topic:   "gotest",
		Consumer: ConsumerConfig{
			GroupID: "gotest",
		},
	}

	c, err := NewConsumer(logger(), conf)
	require.NoError(t, err)

	return c, conf
}
