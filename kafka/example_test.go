package kafka

import (
	"context"
	"time"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

func ExampleConfig_auth() {
	_ = Config{
		// Append the following to your configuration (Consumer or Producer)
		AuthType:  AuthTypeSCRAM256,
		User:      "my-user",
		Password:  "my-secret-password",
		CAPEMFile: "/etc/certificate.pem",
	}
}

func ExampleConsumer() {
	conf := Config{
		Topic:   "example-topic",
		Brokers: []string{"localhost:9092"},
		Consumer: ConsumerConfig{
			GroupID: "example-group",
		},
	}

	log := logrus.New()
	c, err := NewConsumer(log, conf)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// Consider implementing a retry mechanism.
	for {
		// 1. Fetch the message.
		msg, err := c.FetchMessage(ctx)
		if err != nil {
			log.WithError(err).Fatal("error fetching message")
		}

		log.WithField("msg", msg.String()).Debug("Msg got fetched")

		// 2. Do whatever you need to do with the msg.

		// 3. Then commit the message.
		if err := c.CommitMessage(msg); err != nil {
			log.WithError(err).Fatal("error commiting message")
		}
	}
}

func ExampleProducer() {
	conf := Config{
		Brokers: []string{"localhost:9092"},
	}

	log := logrus.New()
	p, err := NewProducer(conf)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	topic := "example-topic"
	msg := &kafkalib.Message{
		TopicPartition: kafkalib.TopicPartition{Topic: &topic},
		Key:            []byte("example"),
		Value:          []byte("Hello World!"),
		Timestamp:      time.Now(),
	}
	if err := p.Produce(ctx, msg); err != nil {
		log.WithError(err).Fatal("error producing message")
	}
}
