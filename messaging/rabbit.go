package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/netlify/netlify-commons/discovery"
	"github.com/netlify/netlify-commons/tls"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Consumer struct {
	Deliveries <-chan amqp.Delivery
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// Clone creates a new consumer on the same channel.
//
// Make sure that the delivery consumer tags do not conflict. An empty tag
// will result in an auto-generated tag.
func (c *Consumer) Clone(queueName string, delivery *DeliveryDefinition) (*Consumer, error) {
	dd := NewDeliveryDefinition(queueName)
	if delivery != nil {
		dd.merge(delivery)
	}
	del, err := Consume(c.Channel, dd)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		Deliveries: del,
		Connection: c.Connection,
		Channel:    c.Channel,
	}, nil
}

type RabbitConfig struct {
	Servers       []string    `mapstructure:"servers"`
	DiscoveryName string      `split_words:"true" mapstructure:"discovery_name"`
	TLS           *tls.Config `mapstructure:"tls_conf"`

	ExchangeDefinition ExchangeDefinition  `envconfig:"exchange" mapstructure:"exchange"`
	QueueDefinition    QueueDefinition     `envconfig:"queue" mapstructure:"queue"`
	DeliveryDefinition *DeliveryDefinition `envconfig:"delivery" mapstructure:"delivery"`
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

func ValidateRabbitConfig(config *RabbitConfig) error {
	return ValidateRabbitConfigStruct(config.Servers, config.ExchangeDefinition, config.QueueDefinition)
}

func ValidateRabbitConfigStruct(servers []string, exchange ExchangeDefinition, queue QueueDefinition) error {
	if len(servers) == 0 {
		return errors.New("missing RabbitMQ servers in the configuration")
	}

	missing := []string{}
	req := map[string]string{
		"exchange_type": exchange.Type,
		"exchange_name": exchange.Name,
		"queue_name":    queue.Name,
		"binding_key":   queue.BindingKey,
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
	if err := ValidateRabbitConfig(config); err != nil {
		return nil, err
	}

	if config.DiscoveryName != "" {
		servers, err := discoverRabbitServers(config.DiscoveryName)
		if err != nil {
			return nil, err
		}
		config.Servers = servers
	}

	conn, err := DialToRabbit(config.Servers, config.TLS, log)
	if err != nil {
		return nil, err
	}

	return CreateConsumer(conn, config.ExchangeDefinition, config.QueueDefinition, config.DeliveryDefinition, log)
}

// DialToRabbit creates a new AMQP connection.
func DialToRabbit(servers []string, tls *tls.Config, log *logrus.Entry) (*amqp.Connection, error) {
	dialConfig := amqp.Config{
		Heartbeat: 10 * time.Second,
	}

	fields := logrus.Fields{
		"servers": strings.Join(servers, ","),
	}
	if tls != nil {
		tlsConfig, err := tls.TLSConfig()
		if err != nil {
			return nil, err
		}

		if tlsConfig != nil {
			fields["cert_file"] = tls.CertFile
			fields["key_file"] = tls.KeyFile
			fields["ca_files"] = tls.CAFiles
			log.WithFields(fields).Debug("Forcing TLS connection")
			dialConfig.TLSClientConfig = tlsConfig
		}
	}

	log.WithFields(fields).Info("Dialing rabbitmq servers")

	if len(servers) == 1 {
		return amqp.DialConfig(servers[0], dialConfig)
	}

	return connectToCluster(servers, dialConfig)
}

// CreateChannel initializes a new message channel.
func CreateChannel(conn *amqp.Connection, exchange ExchangeDefinition, queue QueueDefinition, log *logrus.Entry) (*amqp.Channel, error) {
	log.Debugf("Original exchange definition: %s", exchange.JSON())
	ed := NewExchangeDefinition(exchange.Name, exchange.Type)
	ed.merge(&exchange)
	log.Debugf("Using exchange definition: %s", ed.JSON())

	log.Debugf("Original queue definition: %s", queue.JSON())
	qd := NewQueueDefinition(queue.Name, queue.BindingKey)
	qd.merge(&queue)
	log.Debugf("Using queue definition %s", qd.JSON())

	log.Info("Binding to exchange and queue")
	ch, _, err := Bind(conn, ed, qd)
	return ch, err
}

// CreateConsumer initializes a new message consumer.
func CreateConsumer(conn *amqp.Connection, exchange ExchangeDefinition, queue QueueDefinition, delivery *DeliveryDefinition, log *logrus.Entry) (*Consumer, error) {
	ch, err := CreateChannel(conn, exchange, queue, log)
	if err != nil {
		return nil, err
	}
	return CreateConsumerOnChannel(conn, ch, queue, delivery, log)
}

// CreateConsumerOnChannel initializes a message consumer on an existing channel.
func CreateConsumerOnChannel(conn *amqp.Connection, ch *amqp.Channel, queue QueueDefinition, delivery *DeliveryDefinition, log *logrus.Entry) (*Consumer, error) {
	dd := NewDeliveryDefinition(queue.Name)
	if delivery != nil {
		log.Debugf("Original delivery definition: %s", delivery.JSON())
		dd.merge(delivery)
	}
	log.Debugf("Using delivery definition: %s", dd.JSON())
	log.Info("Starting to consume from amqp channel")
	del, err := Consume(ch, dd)
	if err != nil {
		return nil, err
	}

	log.Debug("Successfully connected to rabbit and setup consumer")
	return &Consumer{
		Deliveries: del,
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

func discoverRabbitServers(serviceName string) ([]string, error) {
	rabbitUrls := []string{}

	endpoints, err := discovery.DiscoverEndpoints(serviceName)
	if err != nil {
		return rabbitUrls, err
	}

	for _, endpoint := range endpoints {
		rabbitUrls = append(rabbitUrls, fmt.Sprintf("%s:%d", endpoint.Target, endpoint.Port))
	}

	return rabbitUrls, nil
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
