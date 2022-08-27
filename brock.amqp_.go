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

// Consume ...
func (_amqp) Consume(ctx context.Context, c *AMQPChannel, h AMQPConsumeHandler, req *AMQPConsumeRequest) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if c == nil || h == nil || req == nil {
		err := error(nil)
		return err
	}

	chDelivery, err := c.Consume(req.Queue, req.Consumer, req.AutoAck, req.Exclusive, req.NoLocal, req.NoWait, req.Args)
	if err != nil {
		return err
	}

	req.ctx = ctx
	for {
		var res *AMQPConsumeResponse
		select {
		case <-ctx.Done():
			err = ctx.Err()
			res = &AMQPConsumeResponse{err: err}
		case event := <-chDelivery:
			res = &AMQPConsumeResponse{Delivery: &event}
		case event := <-c.NotifyCancel(make(chan string)):
			res = &AMQPConsumeResponse{err: Errorf("cancelled with subscription: %s", event)}
		case event := <-c.NotifyClose(make(chan *AMQPError)):
			res = &AMQPConsumeResponse{err: event}
		}
		h.ConsumeAMQP(req, res)
		if err != nil {
			return err
		}
	}
}

// Publish ...
func (_amqp) Publish(ctx context.Context, c *AMQPChannel, h AMQPPublishHandler, req *AMQPPublishRequest) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if h == nil {
		err := error(nil)
		return err
	}

	err := c.PublishWithContext(ctx, req.Exchange, req.Key, req.Mandatory, req.Immediate, req.Msg)
	if err != nil {
		return err
	}

	var res *AMQPPublishResponse
	select {
	case <-ctx.Done():
		err = ctx.Err()
		res = &AMQPPublishResponse{err: err}
	case event := <-c.NotifyPublish(make(chan amqp091.Confirmation)):
		res = &AMQPPublishResponse{Confirmation: &event}
	case event := <-c.NotifyReturn(make(chan amqp091.Return)):
		res = &AMQPPublishResponse{Return: &event}
	}
	h.PublishAMQP(req, res)
	return err
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

// =============================================================================

type AMQPConsumeHandler interface {
	ConsumeAMQP(req *AMQPConsumeRequest, res *AMQPConsumeResponse)
}
type AMQPConsumeRequest struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      AMQPTable

	ctx context.Context
}
type AMQPConsumeResponse struct {
	Delivery *AMQPDelivery

	err error
}

func (req *AMQPConsumeRequest) Context() context.Context                    { return req.ctx }
func (res *AMQPConsumeResponse) Err() error                                 { return res.err }
func (cf cf) ConsumeAMQP(req *AMQPConsumeRequest, res *AMQPConsumeResponse) { cf(req, res) }

type cf = AMQPConsumeHandlerFunc
type AMQPConsumeHandlerFunc func(req *AMQPConsumeRequest, res *AMQPConsumeResponse)

var _ = AMQPConsumeHandler(cf(nil))

// =============================================================================

type AMQPPublishHandler interface {
	PublishAMQP(req *AMQPPublishRequest, res *AMQPPublishResponse)
}
type AMQPPublishRequest struct {
	Exchange,
	Key string
	Mandatory,
	Immediate bool
	Msg AMQPPublishing

	ctx context.Context
}
type AMQPPublishResponse struct {
	Confirmation *AMQPConfirmation
	Return       *AMQPReturn

	err error
}

func (req *AMQPPublishRequest) Context() context.Context                    { return req.ctx }
func (res *AMQPPublishResponse) Err() error                                 { return res.err }
func (pf pf) PublishAMQP(req *AMQPPublishRequest, res *AMQPPublishResponse) { pf(req, res) }

type pf = AMQPPublishHandlerFunc
type AMQPPublishHandlerFunc func(req *AMQPPublishRequest, res *AMQPPublishResponse)

var _ = AMQPPublishHandler(pf(nil))

// =============================================================================
