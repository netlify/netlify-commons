package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/netlify/netlify-commons/tls"
	"github.com/streadway/amqp"
)

type Consumer struct {
	Config     *RabbitConfig
	Deliveries <-chan amqp.Delivery
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

type RabbitConfig struct {
	Servers []string    `mapstructure:"servers"`
	TLS     *tls.Config `mapstructure:"tls_conf"`

	ExchangeDefinition ExchangeDefinition  `mapstructure:"exchange"`
	QueueDefinition    QueueDefinition     `mapstructure:"queue"`
	DeliveryDefinition *DeliveryDefinition `mapstructure:"delivery"`
}

// ExchangeDefinition defines all the parameters for an exchange
type ExchangeDefinition struct {
	Name string `mapstructure:"name"`
	Type string `mapstructure:"type"`

	// defaulted usually
	Durable    *bool       `mapstructure:"durable"`
	AutoDelete *bool       `mapstructure:"auto_delete"`
	Internal   *bool       `mapstructure:"internal"`
	NoWait     *bool       `mapstructure:"no_wait"`
	Table      *amqp.Table `mapstructure:"table"`
}

func (e *ExchangeDefinition) JSON() string {
	bs, _ := json.Marshal(e)
	return string(bs)
}

func (left *ExchangeDefinition) merge(right *ExchangeDefinition) {
	left.Durable = mergeBool(left.Durable, right.Durable)
	left.AutoDelete = mergeBool(left.AutoDelete, right.AutoDelete)
	left.Internal = mergeBool(left.Internal, right.Internal)
	left.NoWait = mergeBool(left.NoWait, right.NoWait)
	if right.Table != nil {
		left.Table = right.Table
	}
}

// NewExchangeDefinition builds an ExchangeDefinition with defaults
func NewExchangeDefinition(name, exType string) *ExchangeDefinition {
	return &ExchangeDefinition{
		Name:       name,
		Type:       exType,
		Durable:    newBool(true),
		AutoDelete: newBool(true),
		Internal:   newBool(false),
		NoWait:     newBool(false),
	}
}

// QueueDefinition defines all the parameters for a queue
type QueueDefinition struct {
	Name       string `mapstructure:"name"`
	BindingKey string `mapstructure:"binding_key"`

	// defaulted usually
	Durable    *bool       `mapstructure:"durable"`
	AutoDelete *bool       `mapstructure:"auto_delete"`
	Exclusive  *bool       `mapstructure:"exclusive"`
	NoWait     *bool       `mapstructure:"no_wait"`
	Table      *amqp.Table `mapstructure:"table"`
}

func (q *QueueDefinition) JSON() string {
	bs, _ := json.Marshal(q)
	return string(bs)
}

func (left *QueueDefinition) merge(right *QueueDefinition) {
	left.Durable = mergeBool(left.Durable, right.Durable)
	left.AutoDelete = mergeBool(left.AutoDelete, right.AutoDelete)
	left.Exclusive = mergeBool(left.Exclusive, right.Exclusive)
	left.NoWait = mergeBool(left.NoWait, right.NoWait)
	if right.Table != nil {
		left.Table = right.Table
	}
}

// NewQueueDefinition builds a QueueDefinition with defaults
func NewQueueDefinition(name, key string) *QueueDefinition {
	qd := &QueueDefinition{
		Name:       name,
		BindingKey: key,
		Durable:    newBool(true),
		AutoDelete: newBool(true),
		Exclusive:  newBool(false),
		NoWait:     newBool(false),
	}
	return qd
}

// DeliveryDefinition defines all the parameters for a delivery
type DeliveryDefinition struct {
	QueueName   *string    `mapstructure:"queue_name"`
	ConsumerTag *string    `mapstructure:"consumer_tag"`
	Exclusive   *bool      `mapstructure:"exclusive"`
	NoACK       *bool      `mapstructure:"ack"`
	NoLocal     *bool      `mapstructure:"no_local"`
	NoWait      *bool      `mapstructure:"no_wait"`
	Table       amqp.Table `mapstructure:"table"`
}

func (d *DeliveryDefinition) JSON() string {
	bs, _ := json.Marshal(d)
	return string(bs)
}

// NewDeliveryDefinition builds a DeliveryDefinition with defaults
func NewDeliveryDefinition(queueName string) *DeliveryDefinition {
	dd := &DeliveryDefinition{
		QueueName:   new(string),
		ConsumerTag: new(string),
		NoACK:       newBool(false),
		NoLocal:     newBool(false),
		Exclusive:   newBool(false),
		NoWait:      newBool(false),
		Table:       nil,
	}

	*dd.QueueName = queueName
	return dd
}

func (left *DeliveryDefinition) merge(right *DeliveryDefinition) {
	left.QueueName = mergeString(left.QueueName, right.QueueName)
	left.ConsumerTag = mergeString(left.ConsumerTag, right.ConsumerTag)
	left.Exclusive = mergeBool(left.Exclusive, right.Exclusive)
	left.NoACK = mergeBool(left.NoACK, right.NoACK)
	left.NoLocal = mergeBool(left.NoLocal, right.NoLocal)
	left.NoWait = mergeBool(left.NoWait, right.NoWait)
	if right.Table != nil {
		left.Table = right.Table
	}
}

func sanityCheck(config *RabbitConfig) error {
	if len(config.Servers) == 0 {
		return errors.New("missing RabbitMQ servers in the configuration")
	}

	missing := []string{}
	req := map[string]string{
		"exchange_type": config.ExchangeDefinition.Type,
		"exchange_name": config.ExchangeDefinition.Name,
		"queue_name":    config.QueueDefinition.Name,
		"binding_key":   config.QueueDefinition.BindingKey,
	}
	for k, v := range req {
		if v == "" {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return errors.New("Missing required config values: " + strings.Join(missing, ","))
	}

	return nil
}

// ConnectToRabbit will open a TLS connection to rabbit mq
func ConnectToRabbit(config *RabbitConfig, log *logrus.Entry) (*Consumer, error) {
	dialConfig := amqp.Config{
		Heartbeat: 10 * time.Second,
	}

	fields := logrus.Fields{}
	if config.TLS != nil {
		fields["cert_file"] = config.TLS.CertFile
		fields["key_file"] = config.TLS.KeyFile
		fields["ca_files"] = config.TLS.CAFiles

		tlsConfig, err := config.TLS.TLSConfig()
		if err != nil {
			return nil, err
		}

		dialConfig.TLSClientConfig = tlsConfig
	}

	if err := sanityCheck(config); err != nil {
		return nil, err
	}

	log.WithFields(fields).Info("Dialing rabbitmq servers")

	var conn *amqp.Connection
	var err error
	if len(config.Servers) == 1 {
		conn, err = amqp.DialConfig(config.Servers[0], dialConfig)
	} else {
		conn, err = connectToCluster(config.Servers, dialConfig)
	}
	if err != nil {
		return nil, err
	}

	ed := NewExchangeDefinition(config.ExchangeDefinition.Name, config.ExchangeDefinition.Type)
	ed.merge(&config.ExchangeDefinition)
	log.Debugf("Using exchange definition: %s", ed.JSON())
	qd := NewQueueDefinition(config.QueueDefinition.Name, config.QueueDefinition.BindingKey)
	qd.merge(&config.QueueDefinition)
	log.Debugf("Using queue definition %s", qd.JSON())
	dd := NewDeliveryDefinition(config.QueueDefinition.Name)
	if config.DeliveryDefinition != nil {
		dd.merge(config.DeliveryDefinition)
	}
	log.Debugf("Using delivery definition: %s", dd.JSON())

	log.Info("Binding to exchange and queue")

	ch, _, err := Bind(conn, ed, qd)
	if err != nil {
		return nil, err
	}

	log.Info("Starting to consume from amqp channel")
	del, err := Consume(ch, dd)
	if err != nil {
		return nil, err
	}

	log.Debug("Successfully connected to rabbit and setup consumer")
	return &Consumer{
		Deliveries: del,
		Config:     config,
		Connection: conn,
		Channel:    ch,
	}, nil
}

func connectToCluster(addresses []string, dialConfig amqp.Config) (*amqp.Connection, error) {
	// shuffle addresses
	length := len(addresses) - 1
	for i := length; i > 0; i-- {
		j := rand.Intn(i + 1)
		addresses[i], addresses[j] = addresses[j], addresses[i]
	}

	// try to connect one address at a time
	// and fallback to the next connection
	// if there is any error dialing in.
	for i, addr := range addresses {
		c, err := amqp.DialConfig(addr, dialConfig)
		if err != nil {
			if i == length {
				return nil, err
			}
			continue
		}

		if c != nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("unable to connect to the RabbitMQ cluster: %s", strings.Join(addresses, ", "))
}

// Bind will connect to the exchange and queue defined
func Bind(conn *amqp.Connection, ex *ExchangeDefinition, queueDef *QueueDefinition) (*amqp.Channel, *amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	var exTable amqp.Table
	if ex.Table != nil {
		exTable = *ex.Table
	}

	err = channel.ExchangeDeclare(
		ex.Name,
		ex.Type,
		*ex.Durable,
		*ex.AutoDelete,
		*ex.Internal,
		*ex.NoWait,
		exTable,
	)
	if err != nil {
		return nil, nil, err
	}

	var qTable amqp.Table
	if queueDef.Table != nil {
		qTable = *queueDef.Table
	}

	queue, err := channel.QueueDeclare(
		queueDef.Name,
		*queueDef.Durable,
		*queueDef.AutoDelete,
		*queueDef.Exclusive,
		*queueDef.NoWait,
		qTable,
	)
	if err != nil {
		return nil, nil, err
	}

	channel.QueueBind(
		queueDef.Name,
		queueDef.BindingKey,
		ex.Name,
		*queueDef.NoWait,
		qTable,
	)
	if err != nil {
		return nil, nil, err
	}

	return channel, &queue, nil
}

// Consume start to consume off the queue specified
func Consume(channel *amqp.Channel, deliveryDef *DeliveryDefinition) (<-chan amqp.Delivery, error) {
	return channel.Consume(
		*deliveryDef.QueueName,
		*deliveryDef.ConsumerTag,
		*deliveryDef.NoACK,
		*deliveryDef.Exclusive,
		*deliveryDef.NoLocal,
		*deliveryDef.NoWait,
		deliveryDef.Table,
	)
}

// ----------------------------------------------------------------------------
// utils
// ----------------------------------------------------------------------------

func newBool(val bool) *bool {
	res := new(bool)
	*res = val
	return res
}

func mergeString(left *string, right *string) *string {
	if right != nil {
		if left == nil {
			left = new(string)
		}
		*left = *right
	}

	return left
}

func mergeBool(left *bool, right *bool) *bool {
	if right != nil {
		if left == nil {
			left = new(bool)
		}

		*left = *right
	}

	return left
}
