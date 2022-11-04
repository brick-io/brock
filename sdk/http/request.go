package sdkhttp

import (
	"context"
	"errors"
	"net/http"
)

//nolint:gochecknoglobals
var Request request

type request struct{}

// Cancel ...
func (request) Cancel(r *http.Request) *http.Request {
	ctx, cancel := context.WithCancel(r.Context())
	cancel()

	*r = *(r.WithContext(ctx))

	return r
}

// IsCancelled ...
func (request) IsCancelled(r *http.Request) bool {
	return errors.Is(r.Context().Err(), context.Canceled)
}

// Set ...
func (request) Set(r *http.Request, key, val any) *http.Request {
	*r = *(r.WithContext(context.WithValue(r.Context(), key, val)))

	return r
}

// Get ...
func (request) Get(r *http.Request, key any) any {
	return r.Context().Value(key)
}
