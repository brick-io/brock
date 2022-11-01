package brock_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"go.onebrick.io/brock"
)

func testHTTP(t *testing.T) {
	t.Parallel()

	_ = t.Run("middleware", testHTTPmiddleware)
	_ = t.Run("mux", testHTTPmux)
}

func testHTTPmiddleware(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	mw := brock.HTTP.Middleware
	str := "something in between me and you"

	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
	mw.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wr := mw.Wrap(w, r)
			wr.Next(nil)
		}),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wr := mw.Wrap(w, r)
			_ = wr.H2Push("/a.json", http.MethodGet, nil)
		}),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wr := mw.Wrap(w, r)
			if err := wr.Err(); err != nil {
				n, err := wr.Send(http.StatusInternalServerError, nil, nil)
				Expect(err).To(Succeed())
				Expect(n).To(BeNumerically(">", 1))

				return
			}

			ch := make(chan []byte)
			go func() {
				for _, v := range strings.Split(str, " ") {
					ch <- []byte(v + " ")
				}
				close(ch)
			}()
			for p := range ch {
				n, err := wr.Stream(p)
				Expect(err).To(Succeed())
				Expect(n).To(BeNumerically(">", 1))
			}
		}),
	).ServeHTTP(w, r)

	p, err := io.ReadAll(w.Result().Body)
	Expect(err).To(Succeed())
	Expect(string(p)).To(Equal(str + " "))
}

//nolint:funlen
func testHTTPmux(t *testing.T) {
	t.Parallel()

	Expect := NewWithT(t).Expect

	type test struct {
		call     func()
		recovery any
	}

	mux := brock.HTTP.Mux()
	tests := map[string]test{
		"method_1":  {func() { mux.Handle("", "", nil) }, "method: empty"},
		"method_2":  {func() { mux.Handle("XXX", "/", nil) }, "method: invalid: XXX"},
		"path_1":    {func() { mux.Handle("GET", "", nil) }, "path: empty"},
		"path_2":    {func() { mux.Handle("GET", "ABC", nil) }, "path: should be canonical: use \"/abc\" instead of \"ABC\""},
		"pattern_1": {func() { mux.Handle("GET", "/aku/{id}{v}", nil) }, "pattern: need separator"},
		"pattern_2": {func() { mux.Handle("GET", "/aku/{id}/mau/{}", nil) }, "pattern: empty key"},
		"pattern_3": {func() { mux.Handle("GET", "/aku/{id}/mau/{id}", nil) }, "pattern: duplicate key: id"},
		"ok":        {func() { mux.Handle("GET", "/aku/{id}/mau/{v}/makan/nasi/{tipe}", nil) }, ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.recovery == "" {
				Expect(test.call).NotTo(Panic())
			} else {
				Expect(test.call).To(PanicWith(test.recovery))
			}
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		http.Error(w, http.StatusText(code), code)
	})
	h := brock.HTTP.Header.Create(
		brock.HTTP.Header.WithKV("Content-Type", "text/plain; charset=utf-8"),
		brock.HTTP.Header.WithKV("X-Content-Type-Options", "nosniff"),
	)
	mux = brock.HTTP.Mux().
		Handle("GET,PUT,PATCH", "/aku/{id}_{v}/makan/{tipe}", handler).
		Handle("GET,PUT,PATCH", "/aku", handler)
	{
		w, r := newMockHandler("GET", "/aku/123_mau/makan/nasi/goreng", nil)
		mux.ServeHTTP(w, r)
		if ar := (assertResponse{w, http.StatusOK, h, []byte("OK\n")}); !ar.Equal() {
			t.Fatalf(ar.Format(), ar.Args()...)
		}
		u := brock.HTTP.NamedArgsFromRequest(r)
		Expect(u.Get("id")).To(Equal("123"))
		Expect(u.Get("v")).To(Equal("mau"))
		Expect(u.Get("tipe")).To(Equal("nasi/goreng"))
	}
	{
		w, r := newMockHandler("POST", "/aku/123_mau/makan/nasi/goreng", nil)
		mux.ServeHTTP(w, r)
		if ar := (assertResponse{w, http.StatusNotFound, h, []byte("Not Found\n")}); !ar.Equal() {
			t.Fatalf(ar.Format(), ar.Args()...)
		}
	}
	{
		w, r := newMockHandler("POST", "/aku/123mau/makan/nasi/goreng", nil)
		mux.ServeHTTP(w, r)
		if ar := (assertResponse{w, http.StatusNotFound, h, []byte("Not Found\n")}); !ar.Equal() {
			t.Fatalf(ar.Format(), ar.Args()...)
		}
	}
	{
		w, r := newMockHandler("GET", "/aku/?query=abc#hash=123", nil)
		mux.ServeHTTP(w, r)
		if ar := (assertResponse{w, http.StatusOK, h, []byte("OK\n")}); !ar.Equal() {
			t.Fatalf(ar.Format(), ar.Args()...)
		}
	}
}

func newMockHandler(method, target string, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
	return httptest.NewRecorder(), httptest.NewRequest(method, target, body)
}

type assertResponse struct {
	w      *httptest.ResponseRecorder
	code   int
	header http.Header
	body   []byte
}

func (x assertResponse) Equal() bool     { return x.EqualCode() && x.EqualBody() && x.EqualHeader() }
func (x assertResponse) EqualCode() bool { return x.code == x.w.Code }
func (x assertResponse) EqualBody() bool { return bytes.Equal(x.body, x.w.Body.Bytes()) }
func (x assertResponse) EqualHeader() bool {
	actual := x.w.Result().Header
	headerEq := len(x.header) == len(actual)

	for k := range actual {
		for i := range actual[k] {
			headerEq = headerEq && i < len(x.header[k])
			headerEq = headerEq && len(x.header[k]) == len(actual[k])
			headerEq = headerEq && x.header[k][i] == actual[k][i]
		}
	}

	return headerEq
}

func (x assertResponse) Format() string {
	format := ""
	if !x.EqualCode() {
		format += "\nCode Expect: %d\n     Actual: %d"
	}

	if !x.EqualHeader() {
		format += "\nHeader Expect: %s\n       Actual: %s"
	}

	if !x.EqualBody() {
		format += "\nBody Expect: %q\n     Actual: %q"
	}

	return format
}

func (x assertResponse) Args() []any {
	args := make([]any, 0)
	if !x.EqualCode() {
		args = append(args, x.code, x.w.Code)
	}

	if !x.EqualHeader() {
		args = append(args, x.header, x.w.Header())
	}

	if !x.EqualBody() {
		args = append(args, x.body, x.w.Body.Bytes())
	}

	return args
}
