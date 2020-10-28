package kafkatest

import (
	"context"
	"testing"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestFakeKafkaProducer_WaitForKey(t *testing.T) {
	log := logrus.New()
	c, p := KafkaPipe(log)
	defer p.Close()
	defer c.Close()

	key := []byte(`key`)
	val := []byte(`foobar`)
	err := p.Produce(context.Background(), &kafkalib.Message{
		Value: val,
		Key:   key,
	})
	msg, err := c.FetchMessage(context.Background())
	require.NoError(t, err)
	require.Equal(t, key, msg.Key)
	require.Equal(t, val, msg.Value)

	require.NoError(t, c.CommitMessage(msg))
	require.True(t, p.WaitForKey(key))
}
