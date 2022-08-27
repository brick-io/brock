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

	mq := brock.AMQP

	load := func() (*brock.AMQPConnection, error) {
		return mq.Open("", brock.AMQPConfiguration{})
	}

	conn, err := load()
	Expect(err).To(Succeed())

	conn = mq.Connection.Update(conn,
		mq.Connection.WithOnInfo(func(major, minor int, properties map[string]any, locales ...string) {}),
	)

	ch, err := conn.Channel()
	Expect(err).To(Succeed())

	ch = mq.Channel.Update(ch,
		mq.Channel.WithOnCancel(func(s string) {}),
		mq.Channel.WithOnClose(func(err error) {
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
		}),
		mq.Channel.WithOnFlow(func(b bool) {}),
	)

	onConsume := brock.AMQPConsumeHandlerFunc(func(req *brock.AMQPConsumeRequest, res *brock.AMQPConsumeResponse) {
		ctx, err := req.Context(), res.Err()
		_, _ = ctx, err
	})
	err = mq.Consume(ctx, ch, onConsume, &brock.AMQPConsumeRequest{
		Queue:    "",
		Consumer: "",
	})

	onPublish := brock.AMQPPublishHandlerFunc(func(req *brock.AMQPPublishRequest, res *brock.AMQPPublishResponse) {
		ctx, err := req.Context(), res.Err()
		_, _ = ctx, err
	})
	err = mq.Publish(ctx, ch, onPublish, &brock.AMQPPublishRequest{
		Exchange: "",
		Key:      "",
	})
}
