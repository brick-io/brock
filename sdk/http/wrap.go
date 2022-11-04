package sdkhttp

import (
	"bytes"
	"io"
	"net/http"

	"github.com/brick-io/brock/sdk"
)

var (
	ErrAlreadySent     = sdk.Errorf("brock/sdkhttp: already sent to the client")
	ErrAlreadyStreamed = sdk.Errorf("brock/sdkhttp: already streamed to the client")
	ErrUnimplemented   = sdk.Errorf("brock/sdkhttp: unimplemented")
)

type (
	ctxKeyMiddlewareNextErr         struct{}
	ctxKeyMiddlewareAlreadySent     struct{}
	ctxKeyMiddlewareAlreadyStreamed struct{}
)

//nolint:gochecknoglobals
var Wrap wrap

type wrap struct{}

// Middleware multiple handlers as one http.Handler.
func (wrap) Middleware(handlers ...http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range handlers {
			if h == nil {
				continue
			}
			h.ServeHTTP(w, r)

			if Request.IsCancelled(r) {
				break
			} else if Request.Get(r, ctxKeyMiddlewareAlreadySent{}) != nil {
				break
			}
		}
	})
}

// WrapHandler helper contract.
type WrapHandler interface {
	// Err get any error passed from the previous handler
	Err() error
	// Next pass the error to the next handler
	Next(err error)
	// Send is a shorthand for set the statusCode, header & body
	Send(statusCode int, header http.Header, body io.Reader) (int, error)
	// Stream is used for streaming response to the client
	Stream(p []byte) (int, error)
	// H2Push initiate a HTTP/2 server push
	H2Push(target, method string, header http.Header) error
}

// Handler the middleware helper from http.ResponseWriter and *http.Request.
func (wrap) Handler(w http.ResponseWriter, r *http.Request) WrapHandler {
	return &handler{w, r}
}

type handler struct {
	w http.ResponseWriter
	r *http.Request
}

// Err get any error passed from the previous handler.
func (x *handler) Err() error {
	err, _ := Request.Get(x.r, ctxKeyMiddlewareNextErr{}).(error)

	return err
}

// Next pass the error to the next handler.
func (x *handler) Next(err error) {
	if err != nil {
		*x.r = *(Request.Set(x.r, ctxKeyMiddlewareNextErr{}, err))
	}
}

// Send is a shorthand for set the statusCode, header & body.
func (x *handler) Send(statusCode int, header http.Header, body io.Reader) (int, error) {
	if http.StatusText(statusCode) == "" {
		return 0, nil
	} else if Request.Get(x.r, ctxKeyMiddlewareAlreadySent{}) != nil {
		return 0, ErrAlreadySent
	} else if Request.Get(x.r, ctxKeyMiddlewareAlreadyStreamed{}) != nil {
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
	*x.r = *(Request.Set(x.r, ctxKeyMiddlewareAlreadySent{}, sdk.NonNil))

	return int(n), err
}

// Stream is used for streaming response to the client.
func (x *handler) Stream(p []byte) (int, error) {
	if len(p) < 1 {
		return 0, nil
	} else if Request.Get(x.r, ctxKeyMiddlewareAlreadySent{}) != nil {
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

	*x.r = *(Request.Set(x.r, ctxKeyMiddlewareAlreadyStreamed{}, sdk.NonNil))

	return n, err
}

// H2Push initiate a HTTP/2 server push.
func (x *handler) H2Push(target, method string, header http.Header) error {
	if target == "" {
		return nil
	} else if Request.Get(x.r, ctxKeyMiddlewareAlreadySent{}) != nil {
		return ErrAlreadySent
	} else if Request.Get(x.r, ctxKeyMiddlewareAlreadyStreamed{}) != nil {
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
