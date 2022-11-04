package sdkhttp

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"

	"github.com/brick-io/brock/sdk"
	sdkcrypto "github.com/brick-io/brock/sdk/crypto"
)

func Mux() *mux {
	return &mux{
		entries: make(map[string]muxEntry),
		panicHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b64 := base64.RawStdEncoding
			btoa := func(b []byte) []byte {
				enc := make([]byte, b64.EncodedLen(len(b)))
				b64.Encode(enc, b)

				return enc
			}

			pub, pvt, _ := sdkcrypto.NaCl.Box.Generate()
			key := sdkcrypto.NaCl.Box.SharedKey(pub, pvt)
			stack := sdk.Sprint(PanicRecoveryFromRequest(r)) + string(debug.Stack())
			cipher := sdkcrypto.NaCl.Box.SealWithSharedKey([]byte(stack), key)
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

type muxEntry struct {
	parts []string
	http.Handler
}

type mux struct {
	entries         map[string]muxEntry
	panicHandler    http.Handler
	notFoundHandler http.Handler
}

// Handle register http.Handler based on the given pattern.
func (x *mux) Handle(method, pattern string, h http.Handler) *mux {
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

	x.entries[method+" "+pattern] = muxEntry{x.parts(pattern), h}

	return x
}

// HandleNotFound register http.Handler that called when no matches request.
func (x *mux) HandleNotFound(h http.Handler) *mux {
	x.notFoundHandler = h

	return x
}

// HandlePanic register http.Handler that called when panic occurred, to access the recovered value
//
//	brock.HTTP.PanicRecoveryFromRequest(r)
func (x *mux) HandlePanic(h http.Handler) *mux {
	x.panicHandler = h

	return x
}

// ServeHTTP implement the http.Handler.
func (x *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var _ http.Handler = x

	defer func() {
		if rcv := recover(); rcv != nil {
			Request.Set(r, ctxKeyPanicRecovery{}, rcv)
			x.panicHandler.ServeHTTP(w, Request.Cancel(r))
		}
	}()

	key := x.requestKey(r)
	if e, ok := x.entries[key]; len(key) > 0 && ok && e.Handler != nil {
		e.ServeHTTP(w, r)

		return
	}

	x.notFoundHandler.ServeHTTP(w, r)
}

func (x *mux) parts(pattern string) []string {
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

func (x *mux) requestKey(r *http.Request) string {
	pat, n, u, m := x.canonicalPath(r.URL.String()), 0, make(url.Values), r.Method
	k := m + " " + pat

	if _, ok := x.entries[k]; ok &&
		!strings.Contains(pat, "{") &&
		!strings.Contains(pat, "}") && ok {
		return k // match with exact entry
	}

	for k, entry := range x.entries {
		if strings.Contains(k, m) && len(entry.parts) > 0 {
			n, u = x.parse(pat, n, u, k)
			if len(u) > 0 {
				Request.Set(r, ctxKeyNamedArguments{}, u)
			}

			_ = n

			return k // match with variables
		}
	}

	return "" // no match
}

func (x *mux) canonicalPath(s string) string {
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
		if !strings.Contains(s[0:h], ".") {
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

func (x *mux) parse(pattern string, n int, u url.Values, k string) (int, url.Values) {
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

func (x *mux) setKV(u url.Values, key, val string) {
	if len(val) > 0 {
		u.Set(key, val)
	}
}

func (x *mux) isStatic(part string) bool {
	return len(part) > 0 && part[0] != '{' && part[len(part)-1] != '}'
}

func (x *mux) isVars(part string) bool {
	return len(part) > 2 && part[0] == '{' && part[len(part)-1] == '}'
}
