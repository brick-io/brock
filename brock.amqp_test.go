package brock_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"go.onebrick.io/brock"
)

func test_amqp(t *testing.T) {
	Expect := NewWithT(t).Expect
	ctx := context.Background()

	amqp, url, cfg := brock.AMQP, "", brock.AMQPConfiguration{}
	load := func() (*brock.AMQPConnection, error) { return amqp.Open(url, cfg) }

	conn, err := load()
	Expect(err).To(Succeed())

	_, _, _, _ = conn.Major, conn.Minor, conn.Properties, conn.Locales

	ch, err := conn.Channel()
	Expect(err).To(Succeed())
	onCancel := func(s string) {
		//
	}
	onClose := func(err error) {
		var e *brock.AMQPError
		if errors.As(err, &e) {
			//
		}
		ch, err = conn.Channel()
		if err == nil && ch != nil {
			return
		}
		conn, err = load()
		if err == nil && conn != nil {
			ch, err = conn.Channel()
		}
	}
	onFlow := func(b bool) {
		//
	}

	ch = brock.Apply(ch,
		amqp.WithOnCancel(onCancel),
		amqp.WithOnClose(onClose),
		amqp.WithOnFlow(onFlow),
		amqp.Consume(new(brock.AMQPConsumeRequest).
			WithContext(ctx, "", "", false, false, false, false, nil),
			brock.AMQPConsumeHandlerFunc(func(r *brock.AMQPConsumeRequest, d *brock.AMQPDelivery, err error) {
				_ = r.Context()
			}),
		),
		amqp.Consume(new(brock.AMQPConsumeRequest).
			WithContext(ctx, "", "", false, false, false, false, nil),
			brock.AMQPConsumeHandlerFunc(func(r *brock.AMQPConsumeRequest, d *brock.AMQPDelivery, err error) {
				_ = r.Context()
			}),
		),
	)

	_, _, _ = amqp.Publish(new(brock.AMQPPublishRequest).
		WithContext(ctx, "", "", false, false, nil),
		ch,
	)
	_, _, _ = amqp.Publish(new(brock.AMQPPublishRequest).
		WithContext(ctx, "", "", false, false, nil),
		ch,
	)
	_, _, _ = amqp.Publish(new(brock.AMQPPublishRequest).
		WithContext(ctx, "", "", false, false, nil),
		ch,
	)
}
