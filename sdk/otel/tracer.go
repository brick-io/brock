package sdkotel

import (
	"context"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

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

	"github.com/brick-io/brock/sdk"
)

// -----------------------------------------------------------------------------
// Tracer
// -----------------------------------------------------------------------------

type TracerConfiguration struct {
	Name   string
	Jaeger struct{ URL string }
	OTLP   struct {
		GRPC struct {
			Compression        string
			Endpoint           string
			Headers            map[string]string
			ReconnectionPeriod time.Duration
			Retry              struct {
				// Enabled indicates whether to not retry sending batches in case of
				// export failure.
				Enabled bool
				// InitialInterval the time to wait after the first failure before
				// retrying.
				InitialInterval time.Duration
				// MaxInterval is the upper bound on backoff interval. Once this value is
				// reached the delay between consecutive retries will always be
				// `MaxInterval`.
				MaxInterval time.Duration
				// MaxElapsedTime is the maximum amount of time (including retries) spent
				// trying to send a request/batch.  Once this value is reached, the data
				// is discarded.
				MaxElapsedTime time.Duration
			}
			// Timeout sets the max amount of time a client will attempt to export a
			// batch of spans. This takes precedence over any retry settings defined with
			// WithRetry, once this time limit has been reached the export is abandoned
			// and the batch of spans is dropped.
			//
			// If unset, the default timeout will be set to 10 seconds.
			Timeout time.Duration
		}
		HTTP struct {
			Compression string
			Endpoint    string
			Headers     map[string]string
			Retry       struct {
				// Enabled indicates whether to not retry sending batches in case of
				// export failure.
				Enabled bool
				// InitialInterval the time to wait after the first failure before
				// retrying.
				InitialInterval time.Duration
				// MaxInterval is the upper bound on backoff interval. Once this value is
				// reached the delay between consecutive retries will always be
				// `MaxInterval`.
				MaxInterval time.Duration
				// MaxElapsedTime is the maximum amount of time (including retries) spent
				// trying to send a request/batch.  Once this value is reached, the data
				// is discarded.
				MaxElapsedTime time.Duration
			}
			// Timeout tells the driver the max waiting time for the backend to process
			// each spans batch.  If unset, the default will be 10 seconds.
			Timeout time.Duration

			// URLPath allows one to override the default URL path used
			// for sending traces. If unset, default ("/v1/traces") will be used.
			URLPath string
		}
	}
}

//nolint:funlen
func Trace(ctx context.Context, c ...*TracerConfiguration) *Tracer {
	if t, ok := ctx.Value(tracerCtxKey{}).(*Tracer); ok && t != nil {
		return t
	}

	if len(c) < 1 {
		c = append(c, nil)
	}

	c0 := c[0]
	if c0 == nil {
		c0 = new(TracerConfiguration)
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
		case c0.OTLP.GRPC.Endpoint != "":
			spanExporter = otlptrace.NewUnstarted(otlptraceClientGRPC(c0))
		case c0.OTLP.HTTP.Endpoint != "":
			spanExporter = otlptrace.NewUnstarted(otlptraceClientHTTP(c0))
		case c0.Jaeger.URL != "":
			spanExporter, _ = jaeger.New(
				jaeger.WithCollectorEndpoint(
					jaeger.WithEndpoint(
						c0.Jaeger.URL,
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
	_ = err

	res, err = sdk_resource.Merge(res, sdk_resource.Default())
	_ = err

	res, err = sdk_resource.Merge(res, sdk_resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.TelemetrySDKLanguageGo,
		semconv.HostArchKey.String(runtime.GOARCH),
		semconv.ServiceNameKey.String(c0.Name),
		semconv.NetHostNameKey.String(netHostName),
		semconv.NetHostIPKey.StringSlice(netHostIPList),
		semconv.OSNameKey.String(runtime.GOOS),
		// semconv.DeploymentEnvironmentKey.String(new(Metadata).Load(ctx).Environment),
	))
	_ = err

	tp := sdk_trace.NewTracerProvider(
		sdk_trace.WithBatcher(spanExporter),
		sdk_trace.WithSampler(sdk_trace.AlwaysSample()),
		sdk_trace.WithResource(res),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tp)

	return &Tracer{tp.Tracer(c0.Name), tracerProviderWrap{tp}}
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

func otlptraceClientGRPC(c *TracerConfiguration) otlptrace.Client {
	opts := make([]otlptracegrpc.Option, 0)

	return otlptracegrpc.NewClient(append(opts,
		otlptracegrpc.WithCompressor(c.OTLP.GRPC.Compression),
		otlptracegrpc.WithEndpoint(c.OTLP.GRPC.Endpoint),
		otlptracegrpc.WithDialOption(),
		otlptracegrpc.WithHeaders(c.OTLP.GRPC.Headers),
		otlptracegrpc.WithReconnectionPeriod(c.OTLP.GRPC.ReconnectionPeriod),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         c.OTLP.GRPC.Retry.Enabled,
			InitialInterval: c.OTLP.GRPC.Retry.InitialInterval,
			MaxInterval:     c.OTLP.GRPC.Retry.MaxInterval,
			MaxElapsedTime:  c.OTLP.GRPC.Retry.MaxElapsedTime,
		}),
		otlptracegrpc.WithTimeout(c.OTLP.GRPC.Timeout),
	)...)
}

func otlptraceClientHTTP(c *TracerConfiguration) otlptrace.Client {
	opts := make([]otlptracehttp.Option, 0)

	for k := range c.OTLP.HTTP.Headers {
		switch strings.ToLower(k) {
		case
			"content-length",
			"content-encoding",
			"content-type":
			delete(c.OTLP.HTTP.Headers, k)
		}
	}

	switch {
	case c.OTLP.HTTP.Timeout == 0:
		c.OTLP.HTTP.Timeout = (10) * time.Second
	case c.OTLP.HTTP.URLPath == "":
		c.OTLP.HTTP.URLPath = "v1/traces"
	}

	return otlptracehttp.NewClient(append(opts,
		otlptracehttp.WithCompression(sdk.IfThenElse(
			strings.ToLower(c.OTLP.HTTP.Compression) == "gzip",
			otlptracehttp.GzipCompression,
			otlptracehttp.NoCompression,
		)),
		otlptracehttp.WithEndpoint(c.OTLP.HTTP.Endpoint),
		otlptracehttp.WithHeaders(c.OTLP.HTTP.Headers),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         c.OTLP.HTTP.Retry.Enabled,
			InitialInterval: c.OTLP.HTTP.Retry.InitialInterval,
			MaxInterval:     c.OTLP.HTTP.Retry.MaxInterval,
			MaxElapsedTime:  c.OTLP.HTTP.Retry.MaxElapsedTime,
		}),
		otlptracehttp.WithTimeout(c.OTLP.HTTP.Timeout),
		otlptracehttp.WithURLPath(c.OTLP.HTTP.URLPath),
	)...)
}
