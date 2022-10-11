package brock

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
)

var (
	HTTP _http
)

type _http struct {
	Header     _http_header
	Query      _http_query
	Multipart  _http_multipart
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

type _http_query struct{}

// Create ...
func (_http_query) Create(opts ...func(url.Values)) url.Values {
	return Apply(make(url.Values), opts...)
}

// WithMap ...
func (_http_query) WithMap(m map[string]string) func(url.Values) {
	return func(h url.Values) {
		for key, value := range m {
			h.Add(key, value)
		}
	}
}

// WithKV ...
func (_http_query) WithKV(key string, values ...string) func(url.Values) {
	return func(h url.Values) {
		for _, value := range values {
			h.Add(key, value)
		}
	}
}

// =============================================================================

type _http_header struct{}

// Create ...
func (_http_header) Create(opts ...func(http.Header)) http.Header {
	return Apply(make(http.Header), opts...)
}

// WithMap ...
func (_http_header) WithMap(m map[string]string) func(http.Header) {
	return func(h http.Header) {
		for key, value := range m {
			h.Add(key, value)
		}
	}
}

// WithKV ...
func (_http_header) WithKV(key string, values ...string) func(http.Header) {
	return func(h http.Header) {
		for _, value := range values {
			h.Add(key, value)
		}
	}
}

// =============================================================================

type _http_multipart struct{}

func (_http_multipart) Create(opts ...func(*multipart.Writer)) *multipart.Writer {
	return Apply(new(multipart.Writer), opts...)
}

func (_http_multipart) WithWriter(w io.Writer) func(*multipart.Writer) {
	return func(mw *multipart.Writer) {
		*mw = *(multipart.NewWriter(w))
	}
}

func (_http_multipart) WithField(key, value string) func(*multipart.Writer) {
	return func(mw *multipart.Writer) {
		_ = mw.WriteField(key, value)
	}
}

func (_http_multipart) WithFile(key, filename string, r io.Reader) func(*multipart.Writer) {
	return func(mw *multipart.Writer) {
		w, err := mw.CreateFormFile(key, filename)
		if err == nil {
			_, _ = io.Copy(w, r)
		}
	}
}

// =============================================================================

type _http_body struct{}

// Create ...
func (_http_body) Create(opts func() io.Reader) io.ReadCloser {
	return io.NopCloser(opts())
}

// WithBytes ...
func (_http_body) WithBytes(v []byte) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBuffer(v)
	}
}

// WithString ...
func (_http_body) WithString(v string) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBufferString(v)
	}
}

// WithJSON ...
func (_http_body) WithJSON(v any) func() io.Reader {
	return func() io.Reader {
		buf := new(bytes.Buffer)
		JSON.NewEncoder(buf).Encode(v)
		return buf
	}
}

type _http_middleware struct{}
type ctx_key_http_middleware_next_err struct{}
type ctx_key_http_middleware_already_sent struct{}
type ctx_key_http_middleware_already_streamed struct{}

// Chain multiple handlers as one http.Handler
func (_http_middleware) Chain(handlers ...http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range handlers {
			if h == nil {
				continue
			}
			h.ServeHTTP(w, r)

			if HTTP.Request.IsCancelled(r) {
				break
			} else if HTTP.Request.Get(r, ctx_key_http_middleware_already_sent{}) != nil {
				break
			}
		}
	})
}

// MiddlewareHTTP helper contract
type MiddlewareHTTP interface {
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

// Wrap the middleware helper from http.ResponseWriter and *http.Request
func (_http_middleware) Wrap(w http.ResponseWriter, r *http.Request) MiddlewareHTTP {
	return &_http_middleware_wrap{w, r}
}

type _http_middleware_wrap struct {
	w http.ResponseWriter
	r *http.Request
}

// Err get any error passed from the previous handler
func (x *_http_middleware_wrap) Err() error {
	err, _ := HTTP.Request.Get(x.r, ctx_key_http_middleware_next_err{}).(error)
	return err
}

// Next pass the error to the next handler
func (x *_http_middleware_wrap) Next(err error) {
	if err != nil {
		*x.r = *(HTTP.Request.Set(x.r, ctx_key_http_middleware_next_err{}, err))
	}
}

// Send is a shorthand for set the statusCode, header & body
func (x *_http_middleware_wrap) Send(statusCode int, header http.Header, body io.Reader) (int, error) {
	if http.StatusText(statusCode) == "" {
		return 0, nil
	} else if HTTP.Request.Get(x.r, ctx_key_http_middleware_already_sent{}) != nil {
		return 0, ErrHTTPAlreadySent
	} else if HTTP.Request.Get(x.r, ctx_key_http_middleware_already_streamed{}) != nil {
		return 0, ErrHTTPAlreadyStreamed
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
	*x.r = *(HTTP.Request.Set(x.r, ctx_key_http_middleware_already_sent{}, NonNil))
	return int(n), err
}

// Stream is used for streaming response to the client
func (x *_http_middleware_wrap) Stream(p []byte) (int, error) {
	if len(p) < 1 {
		return 0, nil
	} else if HTTP.Request.Get(x.r, ctx_key_http_middleware_already_sent{}) != nil {
		return 0, ErrHTTPAlreadySent

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
	*x.r = *(HTTP.Request.Set(x.r, ctx_key_http_middleware_already_streamed{}, NonNil))
	return n, err
}

// H2Push initiate a HTTP/2 server push
func (x *_http_middleware_wrap) H2Push(target, method string, header http.Header) error {
	if target == "" {
		return nil
	} else if HTTP.Request.Get(x.r, ctx_key_http_middleware_already_sent{}) != nil {
		return ErrHTTPAlreadySent
	} else if HTTP.Request.Get(x.r, ctx_key_http_middleware_already_streamed{}) != nil {
		return ErrHTTPAlreadyStreamed
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

type ctx_key_http_mux_named_arguments struct{}

// NamedArgsFromRequest is a helper function that extract url.Values that have
// been parsed using MuxMatcherPattern, url.Values should not be empty if
// parsing is successful and should be able to extract further following
// url.Values, same keys in the pattern result in new value added in url.Values.
func (_http) NamedArgsFromRequest(r *http.Request) url.Values {
	u, _ := HTTP.Request.Get(r, ctx_key_http_mux_named_arguments{}).(url.Values)

	return u
}

type ctx_key_http_mux_panic_recovery struct{}

// PanicRecoveryFromRequest is a helper function that extract error value
// when panic occurred, the value is saved to *http.Request after recovery
// process and right before calling mux.PanicHandler.
func (_http) PanicRecoveryFromRequest(r *http.Request) any {
	return HTTP.Request.Get(r, ctx_key_http_mux_panic_recovery{})
}

func (_http) Mux() *_http_mux {
	return &_http_mux{
		entries: make(map[string]_http_mux_entry),
		panicHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b64 := base64.RawStdEncoding
			btoa := func(b []byte) []byte {
				enc := make([]byte, b64.EncodedLen(len(b)))
				b64.Encode(enc, b)

				return enc
			}

			pub, pvt, _ := Crypto.NaCl.Box.Generate()
			key := Crypto.NaCl.Box.SharedKey(pub, pvt)
			stack := Sprint(HTTP.PanicRecoveryFromRequest(r)) + string(debug.Stack())
			cipher := Crypto.NaCl.Box.SealWithSharedKey([]byte(stack), key)
			w.Header().Set("X-Recovery-Code", string(btoa(key[:])))

			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code)+"\n"+string(btoa(cipher)), code)
		}),
		notFoundHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code := http.StatusNotFound
			http.Error(w, http.StatusText(code), code)
		}),
	}
}

type _http_mux_entry struct {
	parts []string
	http.Handler
}

type _http_mux struct {
	entries map[string]_http_mux_entry
	panicHandler,
	notFoundHandler http.Handler
}

// HandlePanic register http.Handler that called when panic occured,
// to access the recovered value
//
//	brock.HTTP.PanicRecoveryFromRequest(r)
func (x *_http_mux) HandlePanic(h http.Handler) *_http_mux { x.panicHandler = h; return x }

// HandleNotFound register http.Handler that called when no matches request
func (x *_http_mux) HandleNotFound(h http.Handler) *_http_mux { x.notFoundHandler = h; return x }

// Handle register http.Handler based on the given pattern
func (x *_http_mux) Handle(method, pattern string, h http.Handler) *_http_mux {
	if ms := strings.Split(method, ","); len(ms) > 1 {
		for _, method := range ms {
			x.Handle(method, pattern, h)
		}
		return x
	}

	switch method {
	default:
		panic("method: invalid: " + method)
	case "":
		panic("method: empty")
	case
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace:
	}

	if len(pattern) < 1 {
		panic("path: empty")
	} else if pattern != x.canonicalPath(pattern) {
		panic("path: should be canonical: use \"" + x.canonicalPath(pattern) + "\" instead of \"" + pattern + "\"")
	}

	x.entries[method+" "+pattern] = _http_mux_entry{x.parts(pattern), h}
	return x
}

func (x *_http_mux) parts(pattern string) []string {
	parts, keys := make([]string, 0), make(map[string]struct{})
	for i, p := 0, 0; i < len(pattern); i++ {
		if pattern[i] != '{' {
			continue
		}
		parts = append(parts, pattern[p:i])

		// previous rune is '}'
		if i > 0 && pattern[i-1] == '}' {
			panic("pattern: need separator")
		}

		// next rune is '}'
		if i+1 < len(pattern) && pattern[i+1] == '}' {
			panic("pattern: empty key")
		}

		if i+1 < len(pattern) {
			s := strings.Index(pattern[i+1:], "}")
			key := pattern[i+1:][:s]
			if _, ok := keys[key]; ok {
				panic("pattern: duplicate key: " + key)
			}
			i = i + 1 + s
			p = i + 1
			keys[key] = struct{}{}
			parts = append(parts, "{"+key+"}")
		}
	}

	return parts
}

// ServeHTTP implement the http.Handler
func (x *_http_mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var _ http.Handler = x

	defer func() {
		if rcv := recover(); rcv != nil {
			HTTP.Request.Set(r, ctx_key_http_mux_panic_recovery{}, rcv)
			x.panicHandler.ServeHTTP(w, HTTP.Request.Cancel(r))
		}
	}()

	key := x.requestKey(r)
	if e, ok := x.entries[key]; len(key) > 0 && ok && e.Handler != nil {
		e.ServeHTTP(w, r)
		return
	}
	x.notFoundHandler.ServeHTTP(w, r)
}

func (x *_http_mux) requestKey(r *http.Request) string {
	pat, n, u, m := x.canonicalPath(r.URL.String()), 0, make(url.Values), r.Method
	k := m + " " + pat

	if _, ok := x.entries[k]; ok &&
		strings.Index(pat, "{") < 0 &&
		strings.Index(pat, "}") < 0 && ok {
		return k // match with exact entry
	}

	for k, entry := range x.entries {
		if strings.Index(k, m) >= 0 && len(entry.parts) > 0 {
			n, u = x.parse(pat, n, u, k)
			if len(u) > 0 {
				HTTP.Request.Set(r, ctx_key_http_mux_named_arguments{}, u)
			}
			return k // match with variables
		}
	}

	return "" // no match
}

func (x *_http_mux) canonicalPath(s string) string {
	if h := strings.Index(s, "?"); h > 0 {
		s = s[0:h]
	}

	if h := strings.Index(s, "#"); h > 0 {
		s = s[0:h]
	}

	if h := strings.Index(s, "://"); h >= 0 {
		s = s[h+len("://"):]
	}

	if h := strings.Index(s, "/") + 1; h > 0 {
		if strings.Index(s[0:h], ".") < 0 {
			h = 0
		}
		s = s[h:]
	}

	if s[0] != '/' {
		s = "/" + s
	}

	for len(s) > 1 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}

	return strings.ToLower(s)
}

func (x *_http_mux) parse(pattern string, n int, u url.Values, k string) (int, url.Values) {
	e := x.entries[k]
	for i, part := range e.parts {
		switch {
		case x.isStatic(part):
			nn := strings.Index(pattern[n:], part)
			n = n + len(part) + nn
		case x.isVars(part):
			key, val := part[1:len(part)-1], ""
			if i < len(e.parts)-1 {
				next := e.parts[i+1]
				nn := strings.Index(pattern[n:], next)
				if nn > 0 {
					val = pattern[n:][:nn]
					n = nn
					x.setKV(u, key, val)
				}
			} else if n < len(pattern) {
				val = pattern[n:]
				x.setKV(u, key, val)
			}
		}
	}
	return n, u
}

func (x *_http_mux) setKV(u url.Values, key, val string) {
	if len(val) > 0 {
		u.Set(key, val)
	}
}

func (x *_http_mux) isStatic(part string) bool {
	return len(part) > 0 && part[0] != '{' && part[len(part)-1] != '}'
}

func (x *_http_mux) isVars(part string) bool {
	return len(part) > 2 && part[0] == '{' && part[len(part)-1] == '}'
}
