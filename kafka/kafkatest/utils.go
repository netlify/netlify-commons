package kafkatest

import (
	"bytes"
	"context"
	"sync"
	"time"

	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/netlify/netlify-commons/kafka"
	"github.com/netlify/netlify-commons/util"
	"github.com/sirupsen/logrus"
)

func FakeKafkaConsumerFactory(distri <-chan *kafkalib.Message) kafka.ConsumerFactory {
	return func(log logrus.FieldLogger, _ kafka.Config, _ ...kafka.ConfigOpt) (kafka.Consumer, error) {
		return NewFakeKafkaConsumer(log, distri), nil
	}
}

func KafkaConsumerFactoryFromConsumer(c kafka.Consumer) kafka.ConsumerFactory {
	return func(_ logrus.FieldLogger, _ kafka.Config, _ ...kafka.ConfigOpt) (kafka.Consumer, error) {
		return c, nil
	}
}

func KafkaPipe(log logrus.FieldLogger) (*FakeKafkaConsumer, *FakeKafkaProducer) {
	distri := make(chan *kafkalib.Message, 200)
	rdr := NewFakeKafkaConsumer(log, distri)
	wtr := NewFakeKafkaProducer(distri)
	wtr.commits = rdr.commits
	return rdr, wtr
}

type FakeKafkaConsumer struct {
	messages   []*kafkalib.Message
	msgMu      sync.Mutex
	offset     int64
	readOffset int64
	notify     chan struct{}
	commits    chan *kafkalib.Message
	log        logrus.FieldLogger
}

type FakeKafkaProducer struct {
	distris   []chan<- *kafkalib.Message
	distrisMu sync.Mutex
	commits   <-chan *kafkalib.Message
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

func NewFakeKafkaConsumer(log logrus.FieldLogger, distri <-chan *kafkalib.Message) *FakeKafkaConsumer {
	r := &FakeKafkaConsumer{
		messages:   make([]*kafkalib.Message, 0),
		offset:     0,
		readOffset: 0,
		notify:     make(chan struct{}),
		log:        log,
		commits:    make(chan *kafkalib.Message, 1000),
	}

	go func() {
		for msg := range distri {
			r.msgMu.Lock()
			msg.TopicPartition.Offset = kafkalib.Offset(r.offset + 1)
			r.messages = append(r.messages, setMsgDefaults(msg))
			r.msgMu.Unlock()
			r.notify <- struct{}{}
		}
	}()

	return r
}

func (f *FakeKafkaConsumer) FetchMessage(ctx context.Context) (*kafkalib.Message, error) {
	for {
		f.msgMu.Lock()
		if int64(len(f.messages)) > f.readOffset {
			f.log.WithField("offset", f.readOffset).Trace("offering message")
			msg := f.messages[f.readOffset]
			f.msgMu.Unlock()

			f.readOffset = f.readOffset + 1
			return msg, nil
		}
		f.msgMu.Unlock()

		select {
		case <-ctx.Done():
			return &kafkalib.Message{}, ctx.Err()
		case <-f.notify:
		}
	}
}

func (f *FakeKafkaConsumer) CommitMessage(msg *kafkalib.Message) error {
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
	f.readOffset = offset
	f.msgMu.Unlock()
	return nil
}

func (f *FakeKafkaConsumer) Seek(offset int64) error {
	f.msgMu.Lock()
	f.readOffset = offset
	f.msgMu.Unlock()
	return nil
}

func (f *FakeKafkaConsumer) AssignPartitionByKey(key string, algorithm kafka.PartitionerAlgorithm) error {
	return nil // noop
}

func (f *FakeKafkaConsumer) AssignPartitionByID(id int32) error {
	return nil // noop
}

func (f *FakeKafkaConsumer) GetMetadata(allTopics bool) (*kafkalib.Metadata, error) {
	return &kafkalib.Metadata{}, nil // noop
}

func (f *FakeKafkaConsumer) GetPartitions() ([]int32, error) {
	return []int32{}, nil // noop
}

func (f *FakeKafkaConsumer) SeekToTime(t time.Time) error {
	return nil // noop
}

func (f *FakeKafkaConsumer) Close() error {
	close(f.commits)
	return nil
}

func NewFakeKafkaProducer(distris ...chan<- *kafkalib.Message) *FakeKafkaProducer {
	return &FakeKafkaProducer{
		distris: distris,
		closed:  util.NewAtomicBool(false),
	}
}

func (f *FakeKafkaProducer) AddDistri(d chan<- *kafkalib.Message) {
	f.distrisMu.Lock()
	f.distris = append(f.distris, d)
	f.distrisMu.Unlock()
}

func (f *FakeKafkaProducer) Produce(ctx context.Context, msgs ...*kafkalib.Message) error {
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

func setMsgDefaults(msg *kafkalib.Message) *kafkalib.Message {
	if msg.TopicPartition.Topic == nil {
		topicName := "local-test"
		msg.TopicPartition.Topic = &topicName
	}

	return msg
}
