package sdkhttp

import (
	"net/http"

	"github.com/brick-io/brock/sdk"
)

//nolint:gochecknoglobals
var Header header

type header struct{}

// Create ...
func (header) Create(opts ...func(http.Header)) http.Header {
	return sdk.Apply(make(http.Header), opts...)
}

// WithMap ...
func (header) WithMap(m map[string]string) func(http.Header) {
	return func(h http.Header) {
		for key, value := range m {
			h.Add(key, value)
		}
	}
}

// WithKV ...
func (header) WithKV(key string, values ...string) func(http.Header) {
	return func(h http.Header) {
		for _, value := range values {
			h.Add(key, value)
		}
	}
}
