package sdkotel

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

//nolint:gochecknoglobals
var (
	Attribute attr
	Attr      attr
	Code      code
)

type attr struct{}

func (attr) Key(k string) attribute.Key { return attribute.Key(k) }

func (attr) KeyValueHTTPRequest(request *http.Request) []attribute.KeyValue {
	kvs, name := make([]attribute.KeyValue, 0), ""
	kvs = append(kvs, semconv.HTTPServerNameKey.String(name))
	kvs = append(kvs, semconv.HTTPServerAttributesFromHTTPRequest(name, request.URL.Path, request)...)
	kvs = append(kvs, semconv.HTTPServerMetricAttributesFromHTTPRequest(name, request)...)

	return kvs
}

type code struct{}

func (code) StatusUnset() codes.Code { return codes.Unset }

func (code) StatusError() codes.Code { return codes.Error }

func (code) StatusOk() codes.Code { return codes.Ok }
