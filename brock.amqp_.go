package brock

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

var (
	AMQP _amqp
)

type _amqp struct{}

// Open ...
func (_amqp) Open(url string, cfg AMQPConfiguration) (*AMQPConnection, error) {
	return amqp091.DialConfig(url, cfg)
}

// WithPreset ...
func (_amqp) WithPreset() func(*AMQPChannel) {
	return func(ch *AMQPChannel) {
		_ = ch.Qos(1, 0, false)
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

			_ = ch.ExchangeBind(exchangeName2, key, exchangeName, noWait, nil)
			_ = ch.ExchangeUnbind(exchangeName2, key, exchangeName, noWait, nil)
			_ = ch.ExchangeDelete(exchangeName, ifUnused, noWait)
			_ = ch.ExchangeDeclare(exchangeName, exchangeKind, durable, autoDelete, internal, noWait, nil)
			_ = ch.QueueBind(queueName, key, exchangeName, noWait, nil)
			_ = ch.QueueUnbind(queueName, key, exchangeName, nil)
			_, _ = ch.QueueDelete(queueName, ifUnused, ifEmpty, noWait)
			q, _ := ch.QueueInspect(queueName)
			_, _, _ = q.Consumers, q.Messages, q.Name
			_, _ = ch.QueuePurge(queueName, noWait)
			_, _ = ch.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, nil)

			_ = ch.Cancel(consumerName, noWait)
			_ = ch.Flow(false) // pause
			_ = ch.Tx()
			_ = ch.TxCommit()
			_ = ch.TxRollback()
		}
	}
}

// WithOnClose ...
func (_amqp) WithOnClose(fn func(error)) func(*AMQPChannel) {
	return func(ch *AMQPChannel) {
		if fn != nil {
			fn(<-ch.NotifyClose(make(chan *AMQPError)))
		}
	}
}

// WithOnFlow ...
func (_amqp) WithOnFlow(fn func(bool)) func(*AMQPChannel) {
	return func(ch *AMQPChannel) {
		if fn != nil {
			fn(<-ch.NotifyFlow(make(chan bool)))
		}
	}
}

// WithOnCancel ...
func (_amqp) WithOnCancel(fn func(string)) func(*AMQPChannel) {
	return func(ch *AMQPChannel) {
		if fn != nil {
			fn(<-ch.NotifyCancel(make(chan string)))
		}
	}
}

// Consume ...
func (_amqp) Consume(ctx context.Context, req *AMQPConsumeRequest, h AMQPConsumeHandler) func(*AMQPChannel) {
	return func(ch *AMQPChannel) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		if h == nil || req == nil {
			return
		}

		chDelivery, err := ch.Consume(req.Queue, req.Consumer, req.AutoAck, req.Exclusive, req.NoLocal, req.NoWait, req.Args)
		chDone, done := make(chan struct{}, 5), err == nil
		for !done {
			chDone <- struct{}{}
			var d *AMQPDelivery
			select {
			case <-ctx.Done():
				err, done = ctx.Err(), true
			case event := <-ch.NotifyCancel(make(chan string)):
				err = Errorf("cancelled with subscription: %s", event)
			case event := <-ch.NotifyClose(make(chan *AMQPError)):
				err = event
			case event := <-chDelivery:
				d = &event
			}

			func() {
				defer func() { _, _ = <-chDone, recover() }()
				h.ConsumeAMQP(ctx, d, err)
			}()
		}
	}
}

// Publish ...
func (_amqp) Publish(ctx context.Context, ch *AMQPChannel, req *AMQPPublishRequest) (*AMQPConfirmation, *AMQPReturn, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if req == nil {
		err := error(nil)
		return nil, nil, err
	}

	err := ch.PublishWithContext(ctx, req.Exchange, req.Key, req.Mandatory, req.Immediate, req.Msg)
	if err != nil {
		return nil, nil, err
	}

	var confirm *AMQPConfirmation
	var ret *AMQPReturn
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case event := <-ch.NotifyPublish(make(chan amqp091.Confirmation)):
		confirm = &event
	case event := <-ch.NotifyReturn(make(chan amqp091.Return)):
		ret = &event
	}
	return confirm, ret, err
}

// =============================================================================

type AMQPConnection = amqp091.Connection
type AMQPChannel = amqp091.Channel

type AMQPConfiguration = amqp091.Config
type AMQPError = amqp091.Error
type AMQPDelivery = amqp091.Delivery
type AMQPConfirmation = amqp091.Confirmation
type AMQPReturn = amqp091.Return
type AMQPTable = amqp091.Table
type AMQPPublishing = amqp091.Publishing

// =============================================================================

type AMQPConsumeHandlerFunc func(ctx context.Context, d *AMQPDelivery, err error)
type (
	cf  = AMQPConsumeHandlerFunc
	cfD = AMQPDelivery
)

func (cf cf) ConsumeAMQP(ctx context.Context, d *cfD, err error) { cf(ctx, d, err) }

// =============================================================================

type AMQPConsumeHandler interface {
	ConsumeAMQP(ctx context.Context, d *AMQPDelivery, err error)
}
type AMQPConsumeRequest struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      AMQPTable
}

// =============================================================================

type AMQPPublishRequest struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
	Msg       AMQPPublishing
}

// =============================================================================
