package kafkatest

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/netlify/netlify-commons/util"
	"github.com/sirupsen/logrus"
)

func KafkaPipe(log logrus.FieldLogger) (*FakeKafkaConsumer, *FakeKafkaProducer) {
	distri := make(chan *kafka.Message, 200)
	rdr := NewFakeKafkaConsumer(log, distri)
	wtr := NewFakeKafkaProducer(distri)
	wtr.commits = rdr.commits
	return rdr, wtr
}

type FakeKafkaConsumer struct {
	messages []*kafka.Message
	msgMu    sync.Mutex
	offset   int64
	notify   chan struct{}
	commits  chan *kafka.Message
	log      logrus.FieldLogger
}

func (f *FakeKafkaConsumer) Close() error {
	close(f.commits)
	return nil
}

type FakeKafkaProducer struct {
	distris   []chan<- *kafka.Message
	distrisMu sync.Mutex
	commits   <-chan *kafka.Message
	closed    util.AtomicBool
}

func (f *FakeKafkaProducer) Close() error {
	if closed := f.closed.Set(true); closed {
		return nil
	}

	f.distrisMu.Lock()
	for _, d := range f.distris {
		close(d)
	}
	f.distrisMu.Unlock()
	return nil
}

func NewFakeKafkaConsumer(log logrus.FieldLogger, distri <-chan *kafka.Message) *FakeKafkaConsumer {
	r := &FakeKafkaConsumer{
		messages: make([]*kafka.Message, 0),
		offset:   0,
		notify:   make(chan struct{}),
		log:      log,
		commits:  make(chan *kafka.Message, 1000),
	}

	go func() {
		for msg := range distri {
			r.msgMu.Lock()
			msg.TopicPartition.Offset = kafka.Offset(r.offset + 1)
			r.messages = append(r.messages, setMsgDefaults(msg))
			r.msgMu.Unlock()
			r.notify <- struct{}{}
		}
	}()

	return r
}

func (f *FakeKafkaConsumer) FetchMessage(ctx context.Context) (*kafka.Message, error) {
	for {
		f.msgMu.Lock()
		if int64(len(f.messages)) > f.offset {
			f.log.WithField("offset", f.offset).Trace("offering message")
			msg := f.messages[f.offset]
			f.msgMu.Unlock()

			return msg, nil
		}
		f.msgMu.Unlock()

		select {
		case <-ctx.Done():
			return &kafka.Message{}, ctx.Err()
		case <-f.notify:
		}
	}
}

func (f *FakeKafkaConsumer) CommitMessage(msg *kafka.Message) error {
	f.msgMu.Lock()
	f.log.WithField("offset", msg.TopicPartition.Offset).Trace("commiting message...")
	if int64(msg.TopicPartition.Offset) > f.offset {
		f.offset = int64(msg.TopicPartition.Offset)
		f.log.WithField("offset", f.offset).Trace("set new offset")
	}
	select {
	case f.commits <- msg:
	default: // drop if channel is full
	}
	f.msgMu.Unlock()
	return nil
}

func (f *FakeKafkaConsumer) SetInitialOffset(offset int64) error {
	f.msgMu.Lock()
	f.offset = offset
	f.msgMu.Unlock()
	return nil
}

func (f *FakeKafkaConsumer) Seek(offset int64, _ time.Duration) error {
	f.msgMu.Lock()
	f.offset = offset
	f.msgMu.Unlock()
	return nil
}

func NewFakeKafkaProducer(distris ...chan<- *kafka.Message) *FakeKafkaProducer {
	return &FakeKafkaProducer{
		distris: distris,
		closed:  util.NewAtomicBool(false),
	}
}

func (f *FakeKafkaProducer) AddDistri(d chan<- *kafka.Message) {
	f.distrisMu.Lock()
	f.distris = append(f.distris, d)
	f.distrisMu.Unlock()
}

func (f *FakeKafkaProducer) Produce(ctx context.Context, msgs ...*kafka.Message) error {
	f.distrisMu.Lock()
	for _, msg := range msgs {
		for _, d := range f.distris {
			d <- setMsgDefaults(msg)
		}
	}
	f.distrisMu.Unlock()
	return nil
}

func (f *FakeKafkaProducer) WaitForKey(key []byte) (gotKey bool) {
	if f.commits == nil {
		return false
	}

	for msg := range f.commits {
		if bytes.Compare(msg.Key, key) == 0 {
			return true
		}
	}
	// channel closed
	return false
}

func setMsgDefaults(msg *kafka.Message) *kafka.Message {
	if msg.TopicPartition.Topic == nil {
		topicName := "local-test"
		msg.TopicPartition.Topic = &topicName
	}

	return msg
}
