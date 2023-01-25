package sdkotel

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// -----------------------------------------------------------------------------
// Meter
// -----------------------------------------------------------------------------

type MeterConfiguration struct {
	Name string
	OTLP struct {
		ServiceNameKey         string
		ServiceVersionKey      string
		TelemetrySDKVersionKey string
		CollectPeriod          time.Duration
		GRPC                   struct {
			URL     string
			Timeout time.Duration
		}
		HTTP struct {
			URL     string
			Timeout time.Duration
		}
	}
}

func MetricMeter(ctx context.Context, c ...*MeterConfiguration) *Meter {
	if m, ok := ctx.Value(meterCtxKey{}).(*Meter); ok && m != nil {
		return m
	}

	if len(c) < 1 {
		c = append(c, nil)
	}

	c0 := c[0]
	if c0 == nil {
		c0 = new(MeterConfiguration)
	}

	attr := []attribute.KeyValue{
		semconv.ServiceNameKey.String(c0.OTLP.ServiceNameKey),
		semconv.ServiceVersionKey.String(c0.OTLP.ServiceVersionKey),
		semconv.TelemetrySDKVersionKey.String(c0.OTLP.TelemetrySDKVersionKey),
		semconv.TelemetrySDKLanguageGo,
	}

	res0urce, _ := resource.New(ctx, resource.WithAttributes(attr...))

	var metricOpts []otlpmetricgrpc.Option

	switch {
	case c0.OTLP.GRPC.URL != "":
		metricOpts = []otlpmetricgrpc.Option{otlpmetricgrpc.WithTimeout(c0.OTLP.GRPC.Timeout)}
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
		metricOpts = append(metricOpts, otlpmetricgrpc.WithEndpoint(c0.OTLP.GRPC.URL))
	case c0.OTLP.HTTP.URL != "":
		metricOpts = []otlpmetricgrpc.Option{otlpmetricgrpc.WithTimeout(c0.OTLP.HTTP.Timeout)}
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
		metricOpts = append(metricOpts, otlpmetricgrpc.WithEndpoint(c0.OTLP.HTTP.URL))
	}

	metricExporter, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create exporter", err)
	}

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			metricExporter,
		),
		controller.WithResource(res0urce),
		controller.WithExporter(metricExporter),
		controller.WithCollectPeriod(c0.OTLP.CollectPeriod),
	)

	err = pusher.Start(ctx)
	if err != nil {
		log.Fatalf("%s: %v", "failed to start the pusher", err)
	}

	global.SetMeterProvider(pusher)

	return &Meter{global.Meter("io.opentelemetry.metrics.boiva"), nil}
}

type meterCtxKey struct{}

type Meter struct {
	metric.Meter
	_ any
}

type MetricConfiguration struct {
	MetricName        string
	MetricDescription string
}

func (cfg MetricConfiguration) MetricCounter(ctx context.Context, m *Meter) (syncint64.Counter, error) {
	metric, err := m.SyncInt64().Counter(
		cfg.MetricName,
		instrument.WithDescription(cfg.MetricDescription),
	)
	if err != nil {
		return nil, err
	}

	return metric, err
}

func (m *Meter) WithContext(ctx context.Context) context.Context {
	if m_, ok := ctx.Value(meterCtxKey{}).(*Meter); ok {
		if m == m_ {
			return ctx
		}
	}

	return context.WithValue(ctx, meterCtxKey{}, m)
}
