package brock_test

import (
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.onebrick.io/brock"
)

func test_http(t *testing.T) {
	t.Parallel()

	_ = t.Run("test_http_middleware", test_http_middleware)
}

func test_http_middleware(t *testing.T) {
	t.Parallel()

	mw := brock.HTTP.Middleware

	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
	mw.Create(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wr := mw.Wrap(w, r)
			if rand.Intn(1) > 0 {
				wr.Next(brock.ErrUnimplemented)
			} else {
				wr.Next(nil)
			}
		}),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wr := mw.Wrap(w, r)
			wr.H2Push("/a.json", http.MethodGet, nil)
		}),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wr := mw.Wrap(w, r)
			if err := wr.Err(); err != nil {
				NErr(wr.Send(http.StatusInternalServerError, nil, nil))
				return
			}

			ch := make(chan []byte)
			go func() {
				str := "something in between me and you"
				for _, v := range strings.Split(str, " ") {
					ch <- []byte(v + " ")
				}
				close(ch)
			}()
			for p := range ch {
				NErr(wr.Stream(p))
			}
		}),
	).ServeHTTP(w, r)

	p, err := io.ReadAll(w.Result().Body)
	t.Log("\n"+string(p), err)
}

func NErr[T int | int64](n T, err error) { print(n, err) }
