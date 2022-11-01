package brock

import (
	"context"
	"net"
	"os"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdk_resource "go.opentelemetry.io/otel/sdk/resource"
	sdk_trace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

// -----------------------------------------------------------------------------
// Tracer
// -----------------------------------------------------------------------------

type TracerConfiguration struct {
	Name   string
	Jaeger struct{ URL string }
	OTLP   struct {
		GRPC struct{ URL string }
		HTTP struct{ URL string }
	}
}

//nolint:funlen
func (o_t) NewTracer(ctx context.Context, c *TracerConfiguration) *Tracer {
	if t, ok := ctx.Value(tracerCtxKey{}).(*Tracer); ok && t != nil {
		t.tracerProviderWrap.sdk = nil

		return t
	}

	netHostName := func() string {
		hostname, _ := os.Hostname()

		return hostname
	}()
	netHostIPList := func() []string {
		ipList := make([]string, 0)
		addrs, _ := net.InterfaceAddrs()

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ipList = append(ipList, ipnet.IP.String())
			}
		}

		return ipList
	}()
	spanExporter := func() sdk_trace.SpanExporter {
		var spanExporter sdk_trace.SpanExporter

		switch {
		case c.OTLP.GRPC.URL != "":
			spanExporter = otlptrace.NewUnstarted(
				otlptracegrpc.NewClient(
					otlptracegrpc.WithInsecure(),
					otlptracegrpc.WithEndpoint(c.OTLP.GRPC.URL),
				),
			)
		case c.OTLP.HTTP.URL != "":
			spanExporter = otlptrace.NewUnstarted(
				otlptracehttp.NewClient(
					otlptracehttp.WithInsecure(),
					otlptracehttp.WithEndpoint(c.OTLP.HTTP.URL),
					otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
				),
			)
		case c.Jaeger.URL != "":
			spanExporter, _ = jaeger.New(
				jaeger.WithCollectorEndpoint(
					jaeger.WithEndpoint(
						c.Jaeger.URL,
					),
				),
			)
		}

		return spanExporter
	}()

	if spanExporter == nil {
		return nil
	}

	res, err := sdk_resource.New(ctx,
		sdk_resource.WithFromEnv(),
		sdk_resource.WithHost(),
		sdk_resource.WithTelemetrySDK(),
		sdk_resource.WithOS(),
		sdk_resource.WithProcess(),
		sdk_resource.WithContainer(),
	)
	IfThenElse(err != nil, func() { Nop() }, func() { panic(err) })()

	res, err = sdk_resource.Merge(res, sdk_resource.Default())
	IfThenElse(err != nil, func() { Nop() }, func() { panic(err) })()

	res, err = sdk_resource.Merge(res, sdk_resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.TelemetrySDKLanguageGo,
		semconv.HostArchKey.String(runtime.GOARCH),
		semconv.ServiceNameKey.String(c.Name),
		semconv.NetHostNameKey.String(netHostName),
		semconv.NetHostIPKey.StringSlice(netHostIPList),
		semconv.OSNameKey.String(runtime.GOOS),
		// semconv.DeploymentEnvironmentKey.String(new(Metadata).Load(ctx).Environment),
	))

	IfThenElse(err != nil, func() { Nop() }, func() { panic(err) })()

	tp := sdk_trace.NewTracerProvider(
		sdk_trace.WithBatcher(spanExporter),
		sdk_trace.WithSampler(sdk_trace.AlwaysSample()),
		sdk_trace.WithResource(res),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tp)

	return &Tracer{tp.Tracer(c.Name), tracerProviderWrap{tp}}
}

type tracerCtxKey struct{}

type tracerProviderWrap struct{ sdk *sdk_trace.TracerProvider }

func (x tracerProviderWrap) RegisterSpanProcessor(s sdk_trace.SpanProcessor) {
	if x.sdk != nil {
		x.sdk.RegisterSpanProcessor(s)
	}
}

func (x tracerProviderWrap) UnregisterSpanProcessor(s sdk_trace.SpanProcessor) {
	if x.sdk != nil {
		x.sdk.UnregisterSpanProcessor(s)
	}
}

func (x tracerProviderWrap) ForceFlush(ctx context.Context) error {
	if x.sdk != nil {
		return x.sdk.ForceFlush(ctx)
	}

	return nil
}

func (x tracerProviderWrap) Shutdown(ctx context.Context) error {
	if x.sdk != nil {
		return x.sdk.Shutdown(ctx)
	}

	return nil
}

type Tracer struct {
	trace.Tracer
	tracerProviderWrap
}

func (t *Tracer) WithContext(ctx context.Context) context.Context {
	if t_, ok := ctx.Value(tracerCtxKey{}).(*Tracer); ok {
		if t == t_ {
			return ctx
		}
	}

	return context.WithValue(ctx, tracerCtxKey{}, t)
}
