package sdkotel

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	aggregator "go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	aggregation "go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

// -----------------------------------------------------------------------------
// Meter
// -----------------------------------------------------------------------------

type MeterConfiguration struct {
	Name       string
	Prometheus struct {
		HTTPHandlerCallback func(http.Handler)
	}
	OTLP struct {
		GRPC struct{ URL string }
		HTTP struct{ URL string }
	}
}

func Metric(ctx context.Context, c ...*MeterConfiguration) *Meter {
	if m, ok := ctx.Value(meterCtxKey{}).(*Meter); ok && m != nil {
		return m
	}

	opts := make([]controller.Option, 0)
	promConfig := prometheus.Config{}

	if len(c) < 1 {
		c = append(c, nil)
	}

	c0 := c[0]
	if c0 == nil {
		c0 = new(MeterConfiguration)
	}

	switch {
	case c0.OTLP.GRPC.URL != "":
		opts = append(opts, controller.WithExporter(otlpmetric.NewUnstarted(
			otlpmetricgrpc.NewClient(
				otlpmetricgrpc.WithInsecure(),
				otlpmetricgrpc.WithEndpoint(c0.OTLP.GRPC.URL),
				otlpmetricgrpc.WithReconnectionPeriod((10)*time.Second),
			),
		)))
	case c0.OTLP.HTTP.URL != "":
		opts = append(opts, controller.WithExporter(otlpmetric.NewUnstarted(
			otlpmetrichttp.NewClient(
				otlpmetrichttp.WithInsecure(),
				otlpmetrichttp.WithEndpoint(c0.OTLP.HTTP.URL),
				otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
			),
		)))
	}

	ctrl := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				aggregator.WithExplicitBoundaries(promConfig.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
		opts...,
	)

	var mp metric.MeterProvider = ctrl

	if c0.Prometheus.HTTPHandlerCallback != nil {
		if exporter, err := prometheus.New(promConfig, ctrl); err == nil && exporter != nil {
			c0.Prometheus.HTTPHandlerCallback(exporter)

			mp = exporter.MeterProvider()
		}
	}

	global.SetMeterProvider(mp)

	return &Meter{mp.Meter(c0.Name), nil}
}

type meterCtxKey struct{}

type Meter struct {
	metric.Meter
	_ any
}

func (m *Meter) WithContext(ctx context.Context) context.Context {
	if m_, ok := ctx.Value(meterCtxKey{}).(*Meter); ok {
		if m == m_ {
			return ctx
		}
	}

	return context.WithValue(ctx, meterCtxKey{}, m)
}
