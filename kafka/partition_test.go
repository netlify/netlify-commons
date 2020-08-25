package kafka

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
)

func TestPartition(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	testBrokers := os.Getenv("KAFKA_TEST_BROKERS")
	if testBrokers == "" {
		t.Skipf("No local Kafka broker available to run tests")
	}

	// create netlify kafka config
	conf := Config{
		Brokers: strings.Split(testBrokers, ","),
		Topic:   "gotest",
		Consumer: ConsumerConfig{
			GroupID: "gotest",
		},
	}

	// create the producer
	p, err := NewProducer(conf, WithLogger(ctx, logger()), WithPartitionerAlgorithm(PartitionerMurMur2))

	meta, err := p.GetMeta(true, time.Duration(10*time.Second))
	assert.NoError(err)
	assert.NotNil(meta)

	key := "gotestkey"
	val := "gotestval"

	parts, err := p.GetPartions(time.Duration(10 * time.Second))
	assert.NoError(err)
	assert.Len(parts, 3)

	//TODO produce and consume more messages in test
	m := &kafka.Message{
		TopicPartition: kafkalib.TopicPartition{
			Topic: &conf.Topic,
		},
		Key:   []byte(key),
		Value: []byte(val),
	}

	err = p.Produce(ctx, m)
	assert.NoError(err)

	c, err := NewConsumer(logger(), conf)
	assert.NoError(err)
	assert.NotNil(c)

	err = c.SetPartitionByKey(key, time.Duration(10*time.Second))
	assert.NoError(err)

	m, err = c.FetchMessage(ctx)
	assert.NoError(err)
	assert.NotNil(m)
	assert.Equal([]byte(key), m.Key)
	assert.Equal([]byte(val), m.Value)
}
