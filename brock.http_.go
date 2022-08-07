package brock

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
)

var (
	HTTP _http
)

type _http struct {
	Header     _http_header
	Body       _http_body
	Middleware _http_middleware
	Request    _http_request
}

type _http_request struct{}

// Cancel ...
func (_http_request) Cancel(r *http.Request) *http.Request {
	ctx, cancel := context.WithCancel(r.Context())
	cancel()

	*r = *(r.WithContext(ctx))

	return r
}

// IsRequestCancelled ...
func (_http_request) IsCancelled(r *http.Request) bool {
	return errors.Is(r.Context().Err(), context.Canceled)
}

// Set ...
func (_http_request) Set(r *http.Request, key, val any) *http.Request {
	*r = *(r.WithContext(context.WithValue(r.Context(), key, val)))

	return r
}

// Get ...
func (_http_request) Get(r *http.Request, key any) any {
	return r.Context().Value(key)
}

// =============================================================================

type _http_header struct{}

func (_http_header) Create(opts ...func(http.Header)) http.Header {
	return Apply(make(http.Header), opts...)
}

func (_http_header) WithMap(m map[string]string) func(http.Header) {
	return func(h http.Header) {
		for key, value := range m {
			h.Add(key, value)
		}
	}
}

func (_http_header) WithKV(key string, values ...string) func(http.Header) {
	return func(h http.Header) {
		for _, value := range values {
			h.Add(key, value)
		}
	}
}

// =============================================================================

type _http_body struct{}

func (_http_body) Create(opts func() io.Reader) io.ReadCloser {
	return io.NopCloser(opts())
}

func (_http_body) WithBytes(v []byte) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBuffer(v)
	}
}

func (_http_body) WithString(v string) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBufferString(v)
	}
}

func (_http_body) WithJSON(v any) func() io.Reader {
	return func() io.Reader {
		buf := new(bytes.Buffer)
		JSON.NewEncoder(buf).Encode(v)
		return buf
	}
}

type _http_middleware struct{}

func (_http_middleware) Create(h ...http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range h {
			if h == nil {
				continue
			}
			h.ServeHTTP(w, r)

			if HTTP.Request.IsCancelled(r) {
				break
			} else if HTTP.Request.Get(r, ctxKeyAlreadySent) != nil {
				break
			}
		}
	})
}

type ctx_key_http_middleware struct{ string }

var (
	ctxKeyNextErr         = ctx_key_http_middleware{"next_err"}
	ctxKeyAlreadySent     = ctx_key_http_middleware{"already_sent"}
	ctxKeyAlreadyStreamed = ctx_key_http_middleware{"already_streamed"}
)

func (_http_middleware) Wrap(w http.ResponseWriter, r *http.Request) *_http_middleware_wrap {
	return &_http_middleware_wrap{w, r}
}

type _http_middleware_wrap struct {
	w http.ResponseWriter
	r *http.Request
}

func (x *_http_middleware_wrap) Err() error {
	err, _ := HTTP.Request.Get(x.r, ctxKeyNextErr).(error)
	return err
}

func (x *_http_middleware_wrap) Next(err error) {
	if err != nil {
		*x.r = *(HTTP.Request.Set(x.r, ctxKeyNextErr, err))
	}
}

func (x *_http_middleware_wrap) Send(statusCode int, header http.Header, body io.Reader) (int, error) {
	if http.StatusText(statusCode) == "" {
		return 0, nil
	} else if HTTP.Request.Get(x.r, ctxKeyAlreadySent) != nil {
		return 0, ErrAlreadySent
	} else if HTTP.Request.Get(x.r, ctxKeyAlreadyStreamed) != nil {
		return 0, ErrAlreadyStreamed
	}

	for k, vs := range header {
		for _, v := range vs {
			x.w.Header().Add(k, v)
		}
	}
	x.w.WriteHeader(statusCode)
	if body == nil {
		body = new(bytes.Buffer)
	}
	n, err := io.Copy(x.w, body)
	*x.r = *(HTTP.Request.Set(x.r, ctxKeyAlreadySent, NonNil))
	return int(n), err
}

func (x *_http_middleware_wrap) Stream(p []byte) (int, error) {
	if len(p) < 1 {
		return 0, nil
	} else if HTTP.Request.Get(x.r, ctxKeyAlreadySent) != nil {
		return 0, ErrAlreadySent

	}

	type streamer interface {
		http.Flusher
		io.Writer
	}

	w, ok := (x.w).(streamer)
	if !ok {
		return 0, ErrUnimplemented
	}

	n, err := w.Write(p)
	w.Flush()
	*x.r = *(HTTP.Request.Set(x.r, ctxKeyAlreadyStreamed, NonNil))
	return n, err
}

func (x *_http_middleware_wrap) H2Push(target, method string, header http.Header) error {
	if target == "" {
		return nil
	} else if HTTP.Request.Get(x.r, ctxKeyAlreadySent) != nil {
		return ErrAlreadySent
	} else if HTTP.Request.Get(x.r, ctxKeyAlreadyStreamed) != nil {
		return ErrAlreadyStreamed
	}

	w, ok := x.w.(http.Pusher)
	if !ok {
		return ErrUnimplemented
	}

	var opts *http.PushOptions
	if method != "" && header != nil {
		opts = &http.PushOptions{Method: method, Header: header}
	}
	return w.Push(target, opts)
}
