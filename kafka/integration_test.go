package kafka

import (
	"context"
	"fmt"
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
	assert.NoError(err)
	assert.NotNil(p)

	meta, err := p.GetMetadata(true)
	assert.NoError(err)
	assert.NotNil(meta)

	key := "gotestkey"
	val := "gotestval"

	parts, err := p.GetPartions()
	assert.NoError(err)
	assert.Len(parts, 3)

	c, err := NewConsumer(logger(), conf)
	assert.NoError(err)
	assert.NotNil(c)

	// test consuming on multiple partitions
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("%s-%d", key, i)
		v := fmt.Sprintf("%s-%d", val, i)
		m := &kafka.Message{
			TopicPartition: kafkalib.TopicPartition{
				Topic: &conf.Topic,
			},
			Key:   []byte(k),
			Value: []byte(v),
		}

		t := time.Now()
		err = p.Produce(ctx, m)
		assert.NoError(err)

		p := GetPartition(k, parts)

		err = c.SetPartitionByKey(k, PartitionerMurMur2)
		assert.NoError(err)

		err = c.SeekToTime(t)
		assert.NoError(err)

		m, err = c.FetchMessage(ctx)
		assert.NoError(err)
		assert.NotNil(m)
		assert.Equal([]byte(k), m.Key, "Partition to read from: %d, Msg: %+v", p, m)
		assert.Equal([]byte(v), m.Value, "Partition to read from: %d, Msg: %+v", p, m)

		err = c.CommitMessage(m)
		assert.NoError(err)
	}
}
