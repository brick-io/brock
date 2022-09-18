package brock

import (
	"context"
	"net/http"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

type o_t struct {
	Attribute o_t_attribute
	Code      o_t_code
}

// nolint: gochecknoglobals
var (
	OpenTelemetry o_t

	sOUT = os.Stdout
	sERR = os.Stderr
)

type o_t_attribute struct{}

func (o_t_attribute) Key(k string) attribute.Key { return attribute.Key(k) }

func (o_t_attribute) KeyValueHTTPServer(ctx context.Context, request *http.Request) []attribute.KeyValue {
	kvs, name := make([]attribute.KeyValue, 0), new(Metadata).Load(ctx).Namespace
	kvs = append(kvs, semconv.HTTPServerNameKey.String(name))
	kvs = append(kvs, semconv.HTTPServerAttributesFromHTTPRequest(name, request.URL.Path, request)...)
	kvs = append(kvs, semconv.HTTPServerMetricAttributesFromHTTPRequest(name, request)...)

	return kvs
}

type o_t_code struct{}

func (o_t_code) StatusUnset() codes.Code { return codes.Unset }

func (o_t_code) StatusError() codes.Code { return codes.Error }

func (o_t_code) StatusOk() codes.Code { return codes.Ok }
