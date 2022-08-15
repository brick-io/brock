package brock_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"go.onebrick.io/brock"
)

func test_httpmux(t *testing.T) {
	t.Parallel()

	Expect := NewWithT(t).Expect
	host := "http://example.com"
	root := host + "/"

	matcher := brock.HTTP.MuxMatcher
	header := brock.HTTP.Header
	body := brock.HTTP.Body
	mw := brock.HTTP.Middleware

	headerError := header.Create(header.WithMap(map[string]string{
		"Content-Type":           "text/plain; charset=utf-8",
		"X-Content-Type-Options": "nosniff",
	}))

	code200, body200 := 200, []byte(http.StatusText(200))
	code404, body404 := 404, []byte(http.StatusText(404)+"\n")
	code500, body500 := 500, []byte(http.StatusText(500)+"\n")

	mw200 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw.Wrap(w, r).Send(code200, nil, body.Create(body.WithBytes(body200)))
	})
	mw500 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw.Wrap(w, r).Send(code500, headerError, body.Create(body.WithBytes(body500)))
	})

	t.Run("middleware", func(t *testing.T) {
		w, r := newMockHandler("", root, nil)
		brock.HTTP.Middleware.Chain(
			mw200,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { *r = *brock.HTTP.Request.Cancel(r) }),
			mw500,
		).ServeHTTP(w, r)
		Expect(testResponse(t, w, code200, nil, body200)).To(BeTrue())
	})
	t.Run("mux", func(t *testing.T) {
		t.Run("without-panic-handler", func(t *testing.T) {
			w, r := newMockHandler("", root, nil)
			brock.HTTP.Mux().
				With(
					http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(0) }),
					matcher.Mock(0, true, true)).
				ServeHTTP(w, r)
			Expect(testResponse(t, w, code500, headerError, body500)).To(BeTrue())
		})
		t.Run("without-notfound-handler", func(t *testing.T) {
			w, r := newMockHandler("", root, nil)
			brock.HTTP.Mux().
				ServeHTTP(w, r)
			Expect(testResponse(t, w, code404, headerError, body404)).To(BeTrue())
		})
		t.Run("with-panic-handler", func(t *testing.T) {
			w, r := newMockHandler("", root, nil)

			mux := brock.HTTP.Mux().
				With(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic(99) }),
					matcher.Mock(0, true, true)).
				With(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic(99) }),
					matcher.Mock(0, true, false))
			mux.PanicHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(brock.HTTP.PanicRecoveryFromRequest(r)).To(Equal(99))
				mw500.ServeHTTP(w, r)
			})
			mux.ServeHTTP(w, r)

			Expect(testResponse(t, w, code500, headerError, body500)).To(BeTrue())
		})
		t.Run("with-notfound-handler", func(t *testing.T) {
			w, r := newMockHandler("", root, nil)
			mux := brock.HTTP.Mux()
			mux.NotFoundHandler = mw500
			mux.ServeHTTP(w, r)
			Expect(testResponse(t, w, code500, headerError, body500)).To(BeTrue())
			Expect(testResponse(nil, w, code200, nil, body200)).To(BeFalse())
		})
	})
	t.Run("mock", func(t *testing.T) {
		w, r := newMockHandler("", root, nil)
		brock.HTTP.Mux().
			With(mw200, matcher.Or(0,
				matcher.Mock(0, true, true),
				matcher.Mock(0, true, true),
			)).
			With(mw200, matcher.Or(0,
				matcher.Mock(.1, true, true),
				matcher.Mock(.2, true, true),
			)).
			ServeHTTP(w, r)
		Expect(testResponse(t, w, code200, nil, body200)).To(BeTrue())
	})
	t.Run("methods", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			w, r := newMockHandler("", host+"/x/yyy/z", nil)
			brock.HTTP.Mux().
				With(mw200, matcher.Or(0,
					matcher.Methods(0),
					matcher.Methods(0, "GET"),
					matcher.Methods(0,
						"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE", "*"),
				)).
				ServeHTTP(w, r)
			Expect(testResponse(t, w, code200, nil, body200)).To(BeTrue())
		})
		t.Run("fail", func(t *testing.T) {
			w, r := newMockHandler("", host+"/x/yyy/z", nil)
			brock.HTTP.Mux().
				// With(nil, nil).
				// With(mw200, matcher.Methods(0, "XXX")).
				// With(mw200, nil).
				With(mw200, matcher.Methods(0, "GET")).
				ServeHTTP(w, r)
			Expect(testResponse(t, w, code200, nil, body200)).To(BeTrue())
		})
	})
	t.Run("pattern", func(t *testing.T) {
		t.Run("colon-start", func(t *testing.T) {
			w, r := newMockHandler("", host+"/x/yyy/z", nil)
			mux := brock.HTTP.Mux().
				Handle("*", "/makan", mw200).
				Handle("*", "/:args1", mw200).
				Handle("*", "/:args1/:args2", mw200).
				Handle("GET", "/:args1/:args2/:args3", mw200)
			mux.ServeHTTP(w, r)
			Expect(testResponse(t, w, code200, nil, body200)).To(BeTrue())
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("colon-both", func(t *testing.T) {
			w, r := newMockHandler("", host+"/x/yyy/z", nil)
			mux := brock.HTTP.Mux().
				With(mw200,
					matcher.Pattern(0, "/:args1:/:args2:/:args3:", ":", ":", false))
			mux.ServeHTTP(w, r)
			Expect(testResponse(t, w, code200, nil, body200)).To(BeTrue())
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("double-curly-braces", func(t *testing.T) {
			w, r := newMockHandler("", host+"/x/yyy/z", nil)
			mux := brock.HTTP.Mux().
				With(mw200,
					matcher.Pattern(0, "/{{args1}}/{{args2}}/{{args3}}", "{{", "}}", false))
			mux.ServeHTTP(w, r)
			Expect(testResponse(t, w, code200, nil, body200)).To(BeTrue())
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("exact-no-pattern", func(t *testing.T) {
			w, r := newMockHandler("", host+"/x/yyy/z", nil)
			mux := brock.HTTP.Mux().
				With(mw200,
					matcher.Pattern(0, "/x/yyy/z", "", "", false))
			mux.ServeHTTP(w, r)
			Expect(testResponse(t, w, code200, nil, body200)).To(BeTrue())
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args1")).To(Equal(""))
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args2")).To(Equal(""))
			Expect(brock.HTTP.NamedArgsFromRequest(r).Get("args3")).To(Equal(""))
		})
	})
	t.Run("test response", func(t *testing.T) {
		w, r := newMockHandler("", root, nil)
		Expect(w).NotTo(BeNil())
		Expect(r).NotTo(BeNil())
		brock.HTTP.Mux().ServeHTTP(w, r)
		Expect(testResponse(t, w, code404, headerError, body404)).To(BeTrue())
	})
}

func newMockHandler(method, target string, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
	return httptest.NewRecorder(), httptest.NewRequest(method, target, body)
}

// testResponse check w so that it fulfill code, header & body accordingly.
func testResponse(tb testing.TB, w *httptest.ResponseRecorder, code int, header http.Header, body []byte) bool {
	if len(header) < 1 {
		header = http.Header{}
	}

	codeEq := w.Code == code
	bodyEq := bytes.Equal(w.Body.Bytes(), body)
	headerEq := len(w.Header()) == len(header)

	for k := range w.Header() {
		for i := range w.Header()[k] {
			headerEq = headerEq && i < len(header[k])
			headerEq = headerEq && len(header[k]) == len(w.Header()[k])
			headerEq = headerEq && header[k][i] == w.Header()[k][i]
		}
	}

	logf := func(string, ...any) {}
	if tb != nil {
		logf = tb.Logf
	}

	if !headerEq {
		logf("\nHeader:\n"+
			"    Expect: %s\n"+
			"    Actual: %s\n", header, w.Header())
	}

	if !codeEq {
		logf("\nCode:\n"+
			"    Expect: %d\n"+
			"    Actual: %d\n", code, w.Code)
	}

	if !bodyEq {
		logf("\nBody:\n"+
			"    Expect: %q\n"+
			"    Actual: %q\n", body, w.Body.Bytes())
	}

	return headerEq && codeEq && bodyEq
}
