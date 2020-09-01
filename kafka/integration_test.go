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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {

	testBrokers := os.Getenv("KAFKA_TEST_BROKERS")
	if testBrokers == "" {
		t.Skipf("No local Kafka broker available to run tests")
	}

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
		TimestampFormat:  time.RFC3339Nano,
		DisableColors:    true,
		QuoteEmptyFields: true,
	})
	log.SetLevel(logrus.DebugLevel)

	t.Run("PartitionConsumer", func(t *testing.T) {

		assert := assert.New(t)

		ctx := context.Background()
		offset := int64(0)

		// create netlify kafka config
		conf := Config{
			Brokers: strings.Split(testBrokers, ","),
			Topic:   "gotest",
			Consumer: ConsumerConfig{
				GroupID:       "gotest",
				PartitionKey:  "test",
				InitialOffset: &offset,
			},
		}

		// create the producer
		p, err := NewProducer(conf, WithLogger(ctx, log), WithPartitionerAlgorithm(PartitionerMurMur2))
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

		c, err := NewConsumer(log, conf)
		assert.NoError(err)
		assert.NotNil(c)

		// test consuming on multiple partitions
		for i := 0; i < 100; i++ {
			k := fmt.Sprintf("%s-%d", key, i)
			v := fmt.Sprintf("%s-%d", val, i)
			m := &kafkalib.Message{
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

			err = c.AssignPartitionByKey(k, PartitionerMurMur2)
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

		// chaos 🙈🙊🙉
		// force a rebalance event
		chaosTest(testBrokers, assert)

		err = c.Close()
		assert.NoError(err)

	})

	t.Run("ConsumerWithGroup", func(t *testing.T) {
		assert := assert.New(t)

		ctx := context.Background()

		// create netlify kafka config
		conf := Config{
			Brokers: strings.Split(testBrokers, ","),
			Topic:   "gotest",
			Consumer: ConsumerConfig{
				GroupID: "gotest",
			},
		}

		key := "gotestkey"
		val := "gotestval"

		// create the producer
		p, err := NewProducer(conf, WithLogger(ctx, log), WithPartitionerAlgorithm(PartitionerMurMur2))
		assert.NoError(err)
		assert.NotNil(p)

		m := &kafkalib.Message{
			TopicPartition: kafkalib.TopicPartition{
				Topic: &conf.Topic,
			},
			Key:   []byte(key),
			Value: []byte(val),
		}

		err = p.Produce(ctx, m)
		assert.NoError(err)

		c, err := NewConsumer(log, conf, WithConsumerGroupID("gotest"))
		assert.NoError(err)
		assert.NotNil(c)

		m, err = c.FetchMessage(ctx)
		assert.NoError(err)
		assert.Contains(string(m.Value), val)
		assert.Equal(kafkalib.Offset(30), m.TopicPartition.Offset)

		err = c.CommitMessage(m)
		assert.NoError(err)

		// chaos 🙈🙊🙉
		// force a rebalance event
		chaosTest(testBrokers, assert)

		err = c.Close()
		assert.NoError(err)

	})

	t.Run("ConsumerWithGroupAndOffset", func(t *testing.T) {
		assert := assert.New(t)

		ctx := context.Background()
		offset := int64(1)

		// create netlify kafka config
		conf := Config{
			Brokers: strings.Split(testBrokers, ","),
			Topic:   "gotest",
			Consumer: ConsumerConfig{
				GroupID:       "gotest",
				InitialOffset: &offset,
			},
		}

		key := "gotestkey"
		val := "gotestval"

		_ = key
		_ = val

		c, err := NewConsumer(log, conf, WithConsumerGroupID("gotest"))
		assert.NoError(err)
		assert.NotNil(c)

		m, err := c.FetchMessage(ctx)
		assert.NoError(err)
		assert.Equal(int32(0), m.TopicPartition.Partition)
		assert.Equal(kafkalib.Offset(1), m.TopicPartition.Offset)

		err = c.CommitMessage(m)
		assert.NoError(err)

		m, err = c.FetchMessage(ctx)
		assert.NoError(err)
		assert.Equal(int32(0), m.TopicPartition.Partition)
		assert.Equal(kafkalib.Offset(2), m.TopicPartition.Offset)

		err = c.CommitMessage(m)
		assert.NoError(err)

		// chaos 🙈🙊🙉
		// force a rebalance event
		chaosTest(testBrokers, assert)

		m, err = c.FetchMessage(ctx)
		assert.NoError(err)
		assert.Equal(int32(0), m.TopicPartition.Partition)
		assert.Equal(kafkalib.Offset(3), m.TopicPartition.Offset)

		err = c.CommitMessage(m)
		assert.NoError(err)

		err = c.Close()
		assert.NoError(err)

	})
}

func chaosTest(testBrokers string, assert *assert.Assertions) {
	chaos := os.Getenv("KAFKA_CHAOS")
	ctx := context.Background()
	if chaos == "" {
		a, err := kafkalib.NewAdminClient(&kafkalib.ConfigMap{"bootstrap.servers": testBrokers})
		assert.NoError(err)
		assert.NotNil(a)

		results, err := a.CreateTopics(
			ctx,
			[]kafkalib.TopicSpecification{{
				Topic:             "gotest",
				NumPartitions:     5,
				ReplicationFactor: 1}},
			kafka.SetAdminOperationTimeout(time.Duration(1*time.Minute)))
		assert.NoError(err)
		assert.NotNil(results)
		a.Close()
	}

}
