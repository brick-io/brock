package brock_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"go.onebrick.io/brock"
)

func test_http_middleware(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	header, body := brock.HTTP.Header, brock.HTTP.Body

	mw1 := brock.HTTP.MiddlewareFunc(func(r *http.Request) *http.Response {
		switch r.URL.Query().Get("n") {
		case "1":
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: header.Create(header.WithKV(
					"Server", "brock by brick",
				)),
				Body: body.Create(body.WithJSON(map[string]string{
					"data": "ok",
				})),
			}
		default:
			return nil
		}
	})
	mw2 := brock.HTTP.MiddlewareFunc(func(r *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header: header.Create(header.WithKV(
				"Server", "brock by brick v2",
			)),
			Body: body.Create(body.WithJSON(map[string]string{
				"data": "okok",
			})),
		}
	})
	mw3 := brock.HTTP.MiddlewareFunc(func(r *http.Request) *http.Response {
		brock.HTTP.CancelRequest(r)
		return nil
	})

	t.Run("intercepted", func(t *testing.T) {
		t.Parallel()
		w, r :=
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/endpoint?n=1", nil)

		brock.HTTP.Handle(nil, mw1, mw2).ServeHTTP(w, r)
		Expect(w.Header().Get("Server")).To(Equal("brock by brick"))
		Expect(w.Body.String()).To(Equal(`{"data":"ok"}` + "\n"))
		Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("relayed", func(t *testing.T) {
		t.Parallel()
		w, r :=
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/endpoint", nil)

		brock.HTTP.Handle(func(n int, err error) {
			Expect(err).To(Succeed())
		}, mw1, mw2).ServeHTTP(w, r)
		Expect(w.Header().Get("Server")).To(Equal("brock by brick v2"))
		Expect(w.Body.String()).To(Equal(`{"data":"okok"}` + "\n"))
		Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("invalid-middleware", func(t *testing.T) {
		t.Parallel()
		w, r :=
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/endpoint", nil)

		brock.HTTP.Handle(func(n int, err error) {
			Expect(n).To(Equal(0))
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(brock.ErrEmptyResponse))
		}, nil).ServeHTTP(w, r)
		Expect(w.Header().Get("Server")).To(Equal(""))
		Expect(w.Body.String()).To(Equal(``))
		Expect(w.Result().StatusCode).To(Equal(http.StatusInternalServerError))
	})

	t.Run("cancelled-request", func(t *testing.T) {
		t.Parallel()
		w, r :=
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/endpoint", nil)

		brock.HTTP.Handle(func(n int, err error) {
			Expect(n).To(Equal(0))
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(brock.ErrRequestCancelled))
		}, mw3).ServeHTTP(w, r)
		Expect(w.Header().Get("Server")).To(Equal(""))
		Expect(w.Body.String()).To(Equal(``))
		Expect(w.Result().StatusCode).To(Equal(http.StatusInternalServerError))
	})
}
