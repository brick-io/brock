package main

import (
	"fmt"
	"go.onebrick.io/brock"
	"net/http"
	"os"
)

var middleware = brock.HTTP.MiddlewareFunc(func(request *http.Request) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: brock.HTTP.Header.Create(brock.HTTP.Header.WithMap(map[string]string{
			"content-type": "application/json",
		})),
		Body: brock.HTTP.Body.Create(brock.HTTP.Body.WithJSON(map[string]string{
			"result": "ok",
		})),
	}
})

var checkTokenMiddleware = brock.HTTP.MiddlewareFunc(func(request *http.Request) *http.Response {
	if request.Header.Get("token") == "" {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header: brock.HTTP.Header.Create(brock.HTTP.Header.WithMap(map[string]string{
				"content-type": "application/json",
			})),
			Body: brock.HTTP.Body.Create(brock.HTTP.Body.WithJSON(map[string]string{
				"result": "Unauthorized",
			})),
		}
	}
	return nil
})

var nilMiddleware = brock.HTTP.MiddlewareFunc(func(request *http.Request) *http.Response {
	return nil
})

func main() {
	server := &http.Server{
		Addr: os.Args[1],
		Handler: brock.HTTP.Handle(func(n int, err error) {
			fmt.Println(n, err)
		}, nilMiddleware),
	}

	server.ListenAndServe()

}
