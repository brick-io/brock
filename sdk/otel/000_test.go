package sdkotel_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/brick-io/brock/sdk"
	sdkotel "github.com/brick-io/brock/sdk/otel"
)

func Test_sdkotel(t *testing.T) {
	ctx := context.Background()

	test_logger(ctx, t)
	test_meter(ctx, t)
	test_tracer(ctx, t)
}

func test_logger(ctx context.Context, t *testing.T) {
	f, err := os.CreateTemp("", "test-*")
	o := os.Stdout

	f, o = o, f

	buf := new(bytes.Buffer)
	_ = sdkotel.Log(ctx)
	_ = sdkotel.Log(ctx, nil)
	_ = sdkotel.Log(ctx, buf)
	ctx = sdkotel.Log(ctx, buf, f, o, io.Discard).Context(ctx)
	log := sdkotel.Log(ctx)
	ctx = log.Context(ctx)
	_ = ctx

	log.Info().Err(err).Int("3", 3).Msg("msg")
	log.Warn().Err(err).Int("3", 3).Msg("msg")

	f, o = o, f
	_, _ = f, o
}

func test_meter(ctx context.Context, t *testing.T) {
	Expect := NewWithT(t).Expect

	cfg := &sdkotel.MeterConfiguration{}
	cfg.OTLP.GRPC.URL = "0.0.0.0"

	mtr := sdkotel.Metric(ctx, cfg)
	ctx = mtr.WithContext(ctx)
	mtr2 := sdkotel.Metric(ctx, nil)
	Expect(mtr).To(Equal(mtr2))

	g1, err := mtr.AsyncInt64().Gauge("gauge-1")
	Expect(err).To(Succeed())
	g1.Observe(ctx, 120,
		sdkotel.Attr.Key("institution").String("bca"),
		sdkotel.Attr.Key("source").String("dashboard"),
	)
}

func test_tracer(ctx context.Context, t *testing.T) {
	Expect := NewWithT(t).Expect

	cfg := &sdkotel.TracerConfiguration{}
	cfg.OTLP.GRPC.Endpoint = "0.0.0.0"

	trc := sdkotel.Trace(ctx, cfg)
	ctx = trc.WithContext(ctx)
	trc2 := sdkotel.Trace(ctx, nil)
	Expect(trc).To(Equal(trc2))

	count := 3
	err := sdk.Errorf("some error")
	ctx, span := trc.Start(ctx, "first")

	for i := 0; i < count; i++ {
		<-time.After(time.Second * time.Duration(i))
	}
	span.End()

	ctx, span = trc.Start(ctx, "second")

	for i := 0; i < count; i++ {
		<-time.After(time.Second * time.Duration(i))
	}

	span.AddEvent("second event")
	span.SetAttributes(sdkotel.Attr.KeyValueHTTPRequest(new(http.Request))...)

	if err != nil {
		span.SetStatus(sdkotel.Code.StatusError(), err.Error()+": [error on this func]")
		span.RecordError(err)
	} else {
		span.SetStatus(sdkotel.Code.StatusOk(), "")
	}

	span.End()

	_, _ = ctx, span
}
