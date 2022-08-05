package brock

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
)

var (
	HTTP brock_http
)

type brock_http struct {
	Header   brock_http_header
	Body     brock_http_body
	Response brock_http_response
}

// CancelRequest ...
func (brock_http) CancelRequest(r *http.Request) *http.Request {
	ctx, cancel := context.WithCancel(r.Context())
	cancel()

	*r = *(r.WithContext(ctx))

	return r
}

// IsRequestCancelled ...
func (brock_http) IsRequestCancelled(r *http.Request) bool {
	return errors.Is(r.Context().Err(), context.Canceled)
}

// =============================================================================

type brock_http_header struct{}

func (brock_http_header) Create(opts ...func(http.Header)) http.Header {
	return Apply(make(http.Header), opts...)
}

func (brock_http_header) WithMap(m map[string]string) func(http.Header) {
	return func(h http.Header) {
		for key, value := range m {
			h.Add(key, value)
		}
	}
}

func (brock_http_header) WithKV(key string, values ...string) func(http.Header) {
	return func(h http.Header) {
		for _, value := range values {
			h.Add(key, value)
		}
	}
}

// =============================================================================

type brock_http_body struct{}

func (brock_http_body) Create(opts func() io.Reader) io.ReadCloser {
	return io.NopCloser(opts())
}

func (brock_http_body) WithBytes(v []byte) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBuffer(v)
	}
}

func (brock_http_body) WithString(v string) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBufferString(v)
	}
}

func (brock_http_body) WithJSON(v any) func() io.Reader {
	return func() io.Reader {
		buf := new(bytes.Buffer)
		JSON.NewEncoder(buf).Encode(v)
		return buf
	}
}

// =============================================================================

type brock_http_response struct{}

func (brock_http_response) Create(opts ...func(*http.Response)) *http.Response {
	return Apply(new(http.Response))
}

func (brock_http_response) With(statusCode int, header http.Header, body io.Reader) func(*http.Response) {
	return func(r *http.Response) {
		HTTP.Response.WithStatusCode(statusCode)(r)
		HTTP.Response.WithHeader(header)(r)
		HTTP.Response.WithBody(body)(r)
	}
}

func (brock_http_response) WithStatusCode(statusCode int) func(*http.Response) {
	return func(r *http.Response) {
		r.StatusCode = statusCode
	}
}

func (brock_http_response) WithHeader(header http.Header) func(*http.Response) {
	return func(r *http.Response) {
		r.Header = header
	}
}

func (brock_http_response) WithBody(body io.Reader) func(*http.Response) {
	return func(r *http.Response) {
		r.Body = io.NopCloser(body)
	}
}
