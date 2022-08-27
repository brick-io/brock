package brock_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"go.onebrick.io/brock"
)

func test_amqp(t *testing.T) {
	Expect := NewWithT(t).Expect
	ctx := context.Background()

	mq := brock.AMQP
	conn, err := mq.Open("", brock.AMQPConfiguration{})
	Expect(err).To(Succeed())

	conn = mq.Connection.Update(conn,
		mq.Connection.WithOnInfo(func(major, minor int, properties map[string]any, locales ...string) {}),
	)

	ch, err := conn.Channel()
	Expect(err).To(Succeed())

	ch = mq.Channel.Update(ch,
		mq.Channel.WithOnCancel(func(s string) {}),
		mq.Channel.WithOnClose(func(err error) {}),
		mq.Channel.WithOnFlow(func(b bool) {}),
	)

	onConsume := func(ctx context.Context, c *brock.AMQPDelivery, err error) error {
		return nil
	}
	err = mq.Consume(ctx, ch, onConsume,
		mq.WithQueueAndConsumer("", ""),
		mq.WithConsumeFlag(false, false, false, false),
		mq.WithConsumeArgs(nil),
	)

	onPublish := func(c *brock.AMQPConfirmation, r *brock.AMQPReturn, err error) error {
		return nil
	}
	err = mq.Publish(ctx, ch, onPublish,
		mq.WithExchangeAndKey("", ""),
		mq.WithPublishFlag(false, false),
		mq.WithPublishing(brock.AMQPPublishing{
			Body: []byte(`{}`),
		}),
	)
}
