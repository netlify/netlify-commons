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

	ctx := context.Background()

	err := p.Produce(ctx, &kafkalib.Message{
		Key:   []byte(`key1`),
		Value: []byte(`val1`),
	})
	err = p.Produce(ctx, &kafkalib.Message{
		Key:   []byte(`key2`),
		Value: []byte(`val2`),
	})

	msg, err := c.FetchMessage(ctx)
	require.NoError(t, err)
	require.Equal(t, "key1", string(msg.Key))
	require.Equal(t, "val1", string(msg.Value))

	msg, err = c.FetchMessage(ctx)
	require.NoError(t, err)
	require.Equal(t, "key2", string(msg.Key))
	require.Equal(t, "val2", string(msg.Value))

	require.NoError(t, c.CommitMessage(msg))
	require.True(t, p.WaitForKey([]byte(`key2`)))
}
