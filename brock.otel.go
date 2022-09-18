package brock

import (
	"context"
	"net/http"
	"os"

	a "go.opentelemetry.io/otel/attribute"
	c "go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

type open_telemetry struct{}

// nolint: gochecknoglobals
var (
	OTel, OpenTelemetry open_telemetry

	// logID = strconv.FormatInt(time.Now().UnixMicro(), 36)
	sOUT = os.Stdout
	sERR = os.Stderr
)

type ot = open_telemetry

func (ot) StatusUnset() c.Code { return c.Unset }
func (ot) StatusError() c.Code { return c.Error }
func (ot) StatusOk() c.Code    { return c.Ok }

func (ot) AttributeKey(k string) a.Key { return a.Key(k) }

func (ot) AttributeKVHTTPServer(ctx context.Context, request *http.Request) []a.KeyValue {
	kvs, name := make([]a.KeyValue, 0), new(Metadata).Load(ctx).Namespace
	kvs = append(kvs, semconv.HTTPServerNameKey.String(name))
	kvs = append(kvs, semconv.HTTPServerAttributesFromHTTPRequest(name, request.URL.Path, request)...)
	kvs = append(kvs, semconv.HTTPServerMetricAttributesFromHTTPRequest(name, request)...)

	return kvs
}
