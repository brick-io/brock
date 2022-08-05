package brock

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

type MiddlewareHTTP interface {
	MiddlewareHTTP(req *http.Request) (res *http.Response)
}

type http_mw_fn func(*http.Request) *http.Response

func (mw http_mw_fn) MiddlewareHTTP(req *http.Request) *http.Response {
	return mw(req)
}

// MiddlewareFunc ...
func (brock_http) MiddlewareFunc(fn func(*http.Request) *http.Response) MiddlewareHTTP {
	return http_mw_fn(fn)
}

// Handle is a transformation method from the given *http.Request into *http.Response.
func (brock_http) Handle(cb func(n int, err error), middlewares ...MiddlewareHTTP) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { HTTP.CancelRequest(req) }()

		cbOnce := func(n int64, err error) {}
		if cb != nil {
			once := new(sync.Once)
			cbOnce = func(n int64, err error) { once.Do(func() { cb(int(n), err) }) }
		}

		var res *http.Response
		for _, mw := range middlewares {
			if mw == nil {
				continue
			} else if res = mw.MiddlewareHTTP(req); res != nil {
				break
			}
		}

		if HTTP.IsRequestCancelled(req) {
			cbOnce(0, ErrRequestCancelled)
		}

		if res == nil {
			cbOnce(0, ErrEmptyResponse)
			res = &http.Response{}
		}

		if http.StatusText(res.StatusCode) == "" {
			res.StatusCode = http.StatusInternalServerError
		}

		if res.Body == nil {
			res.Body = io.NopCloser(new(bytes.Buffer))
		}

		// 1. Set Header
		for k, vs := range res.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}

		// 2. Set Status
		w.WriteHeader(res.StatusCode)

		// 3. Write
		cbOnce(io.Copy(w, res.Body))
	})
}
