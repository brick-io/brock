package sdkamqp

import (
	"context"

	"github.com/rabbitmq/amqp091-go"

	"github.com/brick-io/brock/sdk"
)

// Open ...
func Open(url string, cfg Configuration) (*Connection, error) {
	return amqp091.DialConfig(url, cfg)
}

// WithPreset ...
func WithPreset() func(*Channel) {
	return func(ch *Channel) {
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
func WithOnClose(fn func(error)) func(*Channel) {
	return func(ch *Channel) {
		if fn != nil {
			fn(<-ch.NotifyClose(make(chan *Error)))
		}
	}
}

// WithOnFlow ...
func WithOnFlow(fn func(bool)) func(*Channel) {
	return func(ch *Channel) {
		if fn != nil {
			fn(<-ch.NotifyFlow(make(chan bool)))
		}
	}
}

// WithOnCancel ...
func WithOnCancel(fn func(string)) func(*Channel) {
	return func(ch *Channel) {
		if fn != nil {
			fn(<-ch.NotifyCancel(make(chan string)))
		}
	}
}

// Consume ...
func Consume(ctx context.Context, req *ConsumeRequest, h ConsumeHandler) func(*Channel) {
	return func(ch *Channel) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		if h == nil || req == nil {
			return
		}

		chDelivery, err := ch.Consume(req.Queue, req.Consumer, req.AutoAck, req.Exclusive, req.NoLocal, req.NoWait, req.Args)

		chDone, done := make(chan struct{}, (5)), err == nil
		for !done {
			chDone <- struct{}{}

			var d *Delivery
			select {
			case <-ctx.Done():
				err, done = ctx.Err(), true
			case event := <-ch.NotifyCancel(make(chan string)):
				err = sdk.Errorf("cancelled with subscription: %s", event)
			case event := <-ch.NotifyClose(make(chan *Error)):
				err = event
			case event := <-chDelivery:
				d = &event
			}

			func() {
				defer func() { _, _ = <-chDone, recover() }()
				h.Consume(ctx, d, err)
			}()
		}
	}
}

// Publish ...
func Publish(ctx context.Context, ch *Channel, req *PublishRequest) (*Confirmation, *Return, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := error(nil); req == nil {
		return nil, nil, err
	}

	err := ch.PublishWithContext(ctx, req.Exchange, req.Key, req.Mandatory, req.Immediate, req.Msg)
	if err != nil {
		return nil, nil, err
	}

	var (
		confirm *Confirmation
		ret     *Return
	)

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

type (
	Connection = amqp091.Connection
	Channel    = amqp091.Channel

	Configuration = amqp091.Config
	Error         = amqp091.Error
	Delivery      = amqp091.Delivery
	Confirmation  = amqp091.Confirmation
	Return        = amqp091.Return
	Table         = amqp091.Table
	Publishing    = amqp091.Publishing
)

// =============================================================================

type ConsumeHandlerFunc func(ctx context.Context, d *Delivery, err error)

type (
	cf  = ConsumeHandlerFunc
	cfD = Delivery
)

func (cf cf) Consume(ctx context.Context, d *cfD, err error) { cf(ctx, d, err) }

// =============================================================================

type ConsumeHandler interface {
	Consume(ctx context.Context, d *Delivery, err error)
}
type ConsumeRequest struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      Table
}

// =============================================================================

type PublishRequest struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
	Msg       Publishing
}

// =============================================================================
