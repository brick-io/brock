package sdkhttp

import (
	"net/url"

	"github.com/brick-io/brock/sdk"
)

//nolint:gochecknoglobals
var Query query

type query struct{}

// Create ...
func (query) Create(opts ...func(url.Values)) url.Values {
	return sdk.Apply(make(url.Values), opts...)
}

// WithMap ...
func (query) WithMap(m map[string]string) func(url.Values) {
	return func(h url.Values) {
		for key, value := range m {
			h.Add(key, value)
		}
	}
}

// WithKV ...
func (query) WithKV(key string, values ...string) func(url.Values) {
	return func(h url.Values) {
		for _, value := range values {
			h.Add(key, value)
		}
	}
}
