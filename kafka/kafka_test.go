package kafka

import (
	"context"
	"io"
	"io/ioutil"
	"testing"
	"time"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
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
	require.NoError(t, c.Seek(2, time.Millisecond))
}

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

func checkClose(t *testing.T, c io.Closer) {
	require.NoError(t, c.Close())
}

func consumer(t *testing.T) (*ConfluentConsumer, Config) {
	conf := Config{
		Brokers: nil, // No brokers are used for unit test.
		Topic:   "gotest",
		ConsumerConf: ConsumerConfig{
			GroupID: "gotest",
		},
	}

	c, err := NewConsumer(logger(), conf)
	require.NoError(t, err)

	return c, conf
}

func producer(t *testing.T) *ConfluentProducer {
	p, err := NewProducer(Config{}, func(configMap *kafkalib.ConfigMap) {
		_ = configMap.SetKey("queue.buffering.max.ms", 1)
		_ = configMap.SetKey("delivery.timeout.ms", 2)
	})
	require.NoError(t, err)

	return p
}

func logger() logrus.FieldLogger {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)

	return log
}
