package brock_test

import (
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"go.onebrick.io/brock"
)

func test_http(t *testing.T) {
	t.Parallel()

	_ = t.Run("middleware", test_http_middleware)
}

func test_http_middleware(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	mw := brock.HTTP.Middleware
	str := "something in between me and you"

	w, r :=
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil)
	mw.Chain(
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

func NErr[T int | int64](n T, err error) { print(n, err) }
