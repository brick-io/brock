//go:build integration

package sdkamqp_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/brick-io/brock/sdk"
	sdkamqp "github.com/brick-io/brock/sdk/amqp"
)

func Test_sdkamqp(t *testing.T) {
	t.SkipNow()
	if t.Skipped() {
		return
	}

	Expect := NewWithT(t).Expect
	ctx := context.Background()

	url, cfg := "amqps://", sdkamqp.Configuration{}
	load := func() (*sdkamqp.Connection, error) { return sdkamqp.Open(url, cfg) }

	conn, err := load()
	Expect(err).To(Succeed())

	sdk.Nop(conn.Major, conn.Minor, conn.Properties, conn.Locales)

	ch, err := conn.Channel()
	Expect(err).To(Succeed())
	onCancel := func(s string) { sdk.Nop(s) }
	onClose := func(err error) {
		var e *sdkamqp.Error
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
	onFlow := func(b bool) { sdk.Nop(b) }

	ch = sdk.Apply(ch,
		sdkamqp.WithOnCancel(onCancel),
		sdkamqp.WithOnClose(onClose),
		sdkamqp.WithOnFlow(onFlow),
		sdkamqp.Consume(ctx,
			&sdkamqp.ConsumeRequest{
				//
			},
			sdkamqp.ConsumeHandlerFunc(func(ctx context.Context, d *sdkamqp.Delivery, err error) {
				sdk.Nop(ctx, d, err)
			}),
		),
		sdkamqp.Consume(ctx,
			&sdkamqp.ConsumeRequest{
				//
			},
			sdkamqp.ConsumeHandlerFunc(func(ctx context.Context, d *sdkamqp.Delivery, err error) {
				sdk.Nop(ctx, d, err)
			}),
		),
	)

	c, r, err := sdkamqp.Publish(ctx, ch, &sdkamqp.PublishRequest{
		//
	})
	sdk.Nop(c, r, err)
	c, r, err = sdkamqp.Publish(ctx, ch, &sdkamqp.PublishRequest{
		//
	})
	sdk.Nop(c, r, err)
}
