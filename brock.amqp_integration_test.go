//go:build integration

package brock_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"go.onebrick.io/brock"
)

func testAMQP(t *testing.T) {
	t.SkipNow()
	if t.Skipped() {
		return
	}

	Expect := NewWithT(t).Expect
	ctx := context.Background()

	amqp, url, cfg := brock.AMQP, "amqps://", brock.AMQPConfiguration{}
	load := func() (*brock.AMQPConnection, error) { return amqp.Open(url, cfg) }

	conn, err := load()
	Expect(err).To(Succeed())

	brock.Nop(conn.Major, conn.Minor, conn.Properties, conn.Locales)

	ch, err := conn.Channel()
	Expect(err).To(Succeed())
	onCancel := func(s string) { brock.Nop(s) }
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
	onFlow := func(b bool) { brock.Nop(b) }

	ch = brock.Apply(ch,
		amqp.WithOnCancel(onCancel),
		amqp.WithOnClose(onClose),
		amqp.WithOnFlow(onFlow),
		amqp.Consume(ctx,
			&brock.AMQPConsumeRequest{
				//
			},
			brock.AMQPConsumeHandlerFunc(func(ctx context.Context, d *brock.AMQPDelivery, err error) {
				brock.Nop(ctx, d, err)
			}),
		),
		amqp.Consume(ctx,
			&brock.AMQPConsumeRequest{
				//
			},
			brock.AMQPConsumeHandlerFunc(func(ctx context.Context, d *brock.AMQPDelivery, err error) {
				brock.Nop(ctx, d, err)
			}),
		),
	)

	c, r, err := amqp.Publish(ctx, ch, &brock.AMQPPublishRequest{
		//
	})
	brock.Nop(c, r, err)
	c, r, err = amqp.Publish(ctx, ch, &brock.AMQPPublishRequest{
		//
	})
	brock.Nop(c, r, err)
}
