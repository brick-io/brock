package brock

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

var (
	AMQP _amqp
)

type _amqp struct {
	Connection _amqp_connection
	Channel    _amqp_channel
}

// Open ...
func (_amqp) Open(url string, cfg AMQPConfiguration) (*AMQPConnection, error) {
	return amqp091.DialConfig(url, cfg)
}

// _amqp_connection ...
type _amqp_connection struct{}

// Update ...
func (_amqp_connection) Update(c *AMQPConnection, opts ...func(*AMQPConnection)) *AMQPConnection {
	return Apply(c, opts...)
}

// WithOnInfo ...
func (_amqp_connection) WithOnInfo(fn func(major, minor int, properties map[string]any, locales ...string)) func(*AMQPConnection) {
	return func(c *AMQPConnection) {
		if fn != nil {
			fn(c.Major, c.Minor, c.Properties, c.Locales...)
		}
	}
}

// _amqp_channel ...
type _amqp_channel struct{}

// Update ...
func (_amqp_channel) Update(c *AMQPChannel, opts ...func(*AMQPChannel)) *AMQPChannel {
	return Apply(c, opts...)
}

// WithPreset ...
func (_amqp_channel) WithPreset() func(*AMQPChannel) {
	return func(c *AMQPChannel) {
		_ = c.Qos(1, 0, false)
		_ = func() {
			var (
				queueName,
				exchangeName,
				consumerName,
				exchangeName2,
				exchangeKind,
				key,
				_ string

				// autoAck,
				autoDelete,
				durable,
				exclusive,
				internal,
				ifUnused,
				ifEmpty,
				// immediate,
				// mandatory,
				// noLocal,
				noWait,
				_ bool
			)

			_ = c.ExchangeBind(exchangeName2, key, exchangeName, noWait, nil)
			_ = c.ExchangeUnbind(exchangeName2, key, exchangeName, noWait, nil)
			_ = c.ExchangeDelete(exchangeName, ifUnused, noWait)
			_ = c.ExchangeDeclare(exchangeName, exchangeKind, durable, autoDelete, internal, noWait, nil)
			_ = c.QueueBind(queueName, key, exchangeName, noWait, nil)
			_ = c.QueueUnbind(queueName, key, exchangeName, nil)
			_, _ = c.QueueDelete(queueName, ifUnused, ifEmpty, noWait)
			q, _ := c.QueueInspect(queueName)
			_, _, _ = q.Consumers, q.Messages, q.Name
			_, _ = c.QueuePurge(queueName, noWait)
			_, _ = c.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, nil)

			_ = c.Cancel(consumerName, noWait)
			_ = c.Flow(false) // pause
			_ = c.Tx()
			_ = c.TxCommit()
			_ = c.TxRollback()
		}
	}
}

// WithOnClose ...
func (_amqp_channel) WithOnClose(fn func(error)) func(*AMQPChannel) {
	return func(c *AMQPChannel) {
		if fn != nil {
			fn(<-c.NotifyClose(make(chan *AMQPError)))
		}
	}
}

// WithOnFlow ...
func (_amqp_channel) WithOnFlow(fn func(bool)) func(*AMQPChannel) {
	return func(c *AMQPChannel) {
		if fn != nil {
			fn(<-c.NotifyFlow(make(chan bool)))
		}
	}
}

// WithOnCancel ...
func (_amqp_channel) WithOnCancel(fn func(string)) func(*AMQPChannel) {
	return func(c *AMQPChannel) {
		if fn != nil {
			fn(<-c.NotifyCancel(make(chan string)))
		}
	}
}

// =============================================================================

type AMQPConfiguration = amqp091.Config
type AMQPError = amqp091.Error
type AMQPConnection = amqp091.Connection
type AMQPChannel = amqp091.Channel
type AMQPDelivery = amqp091.Delivery
type AMQPConfirmation = amqp091.Confirmation
type AMQPReturn = amqp091.Return
type AMQPTable = amqp091.Table
type AMQPPublishing = amqp091.Publishing
type AMQPConsumeHandlerFunc func(ctx context.Context, d *AMQPDelivery, err error) error
type AMQPPublishHandlerFunc func(c *AMQPConfirmation, r *AMQPReturn, err error) error
type AMQPConsumeHandler interface {
	ConsumeAMQP(ctx context.Context, d *AMQPDelivery, err error) error
}
type AMQPPublishHandler interface {
	PublishAMQP(c *AMQPConfirmation, r *AMQPReturn, err error) error
}

func (fn AMQPConsumeHandlerFunc) ConsumeAMQP(ctx context.Context, d *AMQPDelivery, err error) error {
	return fn(ctx, d, err)
}

func (fn AMQPPublishHandlerFunc) PublishAMQP(c *AMQPConfirmation, r *AMQPReturn, err error) error {
	return fn(c, r, err)
}

var _ = AMQPConsumeHandler(AMQPConsumeHandlerFunc(nil))
var _ = AMQPPublishHandler(AMQPPublishHandlerFunc(nil))

// =============================================================================

// Consume ...
func (_amqp) Consume(ctx context.Context, c *AMQPChannel, h AMQPConsumeHandler, opts ...func(*_amqp_consume_opts)) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if h == nil {
		err := error(nil)
		return err
	}

	o := Apply(new(_amqp_consume_opts), opts...)

	chDelivery, err := c.Consume(o.queue, o.consumer, o.autoAck, o.exclusive, o.noLocal, o.noWait, o.args)
	if err != nil {
		return err
	}

	for {
		var dlv *AMQPDelivery
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case event := <-chDelivery:
			dlv = &event
		case event := <-c.NotifyCancel(make(chan string)):
			err = Errorf("cancel %s", event)
		case event := <-c.NotifyClose(make(chan *AMQPError)):
			err = event
		}
		if err = h.ConsumeAMQP(ctx, dlv, err); err != nil {
			return err
		}
	}
}

// Publish ...
func (_amqp) Publish(ctx context.Context, c *AMQPChannel, h AMQPPublishHandler, opts ...func(*_amqp_publish_opts)) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if h == nil {
		err := error(nil)
		return err
	}

	o := Apply(new(_amqp_publish_opts), opts...)

	err := c.PublishWithContext(ctx, o.exchange, o.key, o.mandatory, o.immediate, o.msg)
	if err != nil {
		return err
	}

	var conf *AMQPConfirmation
	var ret *AMQPReturn
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case event := <-c.NotifyPublish(make(chan amqp091.Confirmation)):
		conf = &event
	case event := <-c.NotifyReturn(make(chan amqp091.Return)):
		ret = &event
	}

	return h.PublishAMQP(conf, ret, err)
}

// =============================================================================

// _amqp_consume_opts ...
type _amqp_consume_opts struct {
	queue,
	consumer string
	autoAck,
	exclusive,
	noLocal,
	noWait bool
	args AMQPTable
}

// WithQueueAndConsumer ...
func (_amqp) WithQueueAndConsumer(queue, consumer string) func(*_amqp_consume_opts) {
	return func(o *_amqp_consume_opts) {
		o.queue = queue
		o.consumer = consumer
	}
}

// WithConsumeFlag ...
func (_amqp) WithConsumeFlag(autoAck, exclusive, noLocal, noWait bool) func(*_amqp_consume_opts) {
	return func(o *_amqp_consume_opts) {
		o.autoAck = autoAck
		o.exclusive = exclusive
		o.noLocal = noLocal
		o.noWait = noWait
	}
}

// WithConsumeArgs ...
func (_amqp) WithConsumeArgs(args AMQPTable) func(*_amqp_consume_opts) {
	return func(o *_amqp_consume_opts) {
		o.args = args
	}
}

// =============================================================================

// _amqp_publish_opts ...
type _amqp_publish_opts struct {
	exchange,
	key string
	mandatory,
	immediate bool
	msg AMQPPublishing
}

// WithExchangeAndKey ...
func (_amqp) WithExchangeAndKey(exchange, key string) func(*_amqp_publish_opts) {
	return func(o *_amqp_publish_opts) {
		o.exchange = exchange
		o.key = key
	}
}

// WithPublishFlag ...
func (_amqp) WithPublishFlag(mandatory, immediate bool) func(*_amqp_publish_opts) {
	return func(o *_amqp_publish_opts) {
		o.mandatory = mandatory
		o.immediate = immediate
	}
}

// WithPublishing ...
func (_amqp) WithPublishing(msg AMQPPublishing) func(*_amqp_publish_opts) {
	return func(o *_amqp_publish_opts) {
		o.msg = msg
	}
}
