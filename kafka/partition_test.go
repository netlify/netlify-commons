package kafka

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestPartition(t *testing.T) {
	assert := assert.New(t)

	testBroker := os.Getenv("KAFKA_TEST_BROKER")
	if testBroker == "" {
		t.Skipf("No local Kafka broker available to run tests")
	}

	conf := Config{
		Brokers: []string{testBroker},
		Topic:   "gotest",
		Consumer: ConsumerConfig{
			GroupID: "gotest",
		},
	}

	ctx := context.Background()

	p, err := NewProducer(conf, WithLogger(ctx, logger()), WithPartitionerAlgorithm(PartitionerMurMur2))

	ctx, _ = context.WithTimeout(ctx, 5*time.Second)
	meta, err := p.GetMeta(ctx, conf.Topic, false)
	assert.NoError(err)
	spew.Dump(meta)
	assert.NotNil(meta)

	key := "gotestkey"
	val := "gotestval"

	parts, err := p.GetPartions(ctx, conf.Topic)
	assert.NoError(err)
	assert.NotNil(parts)
	spew.Dump(parts)

	m := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &conf.Topic,
			Partition: GetPartition(key, parts),
		},
		Key:   []byte(key),
		Value: []byte(val),
	}

	err = p.Produce(ctx, m)
	assert.NoError(err)

	c, err := NewConsumer(logger(), conf)
	m, err = c.FetchMessage(ctx)
	assert.NoError(err)
	assert.Equal([]byte(key), m.Key)
	assert.Equal([]byte(val), m.Value)
}
