package main

import (
	"log"
	"net/http"
	"os"

	"go.onebrick.io/brock"
)

type dict map[string]any

var (
	ok     struct{}
	mw     = brock.HTTP.Middleware
	header = brock.HTTP.Header
	body   = brock.HTTP.Body

	mw1 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := mw.Wrap(w, r)
		err := wr.Err()
		c := http.StatusOK
		if err != nil {
			c = http.StatusInternalServerError
		}
		h := header.Create(header.WithKV("content-type", "application/json"))
		p := dict{"result": http.StatusText(c)}
		if err != nil {
			p["errors"] = []map[string]string{
				{"message": err.Error()},
			}
		}
		b := body.Create(body.WithJSON(p))
		_, _ = wr.Send(c, h, b)
	})

	mw2 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := mw.Wrap(w, r)
		if r.Header.Get("token") == "" {
			c := http.StatusUnauthorized
			// h := header.Create(header.WithKV("content-type", "application/json"))
			// b := body.Create(body.WithJSON(dict{"result": http.StatusText(c)}))
			wr.Next(brock.Errorf(http.StatusText(c)))
			// _, _ = wr.Send(c, h, b)
		}
		wr.Next(nil)
	})

	mw3 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := mw.Wrap(w, r)
		wr.Next(nil)
	})
)

func main() {
	server := &http.Server{
		Addr:              os.Args[1],
		Handler:           mw.Chain(mw3, mw2, mw1),
		TLSConfig:         nil,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          log.Default(),
		BaseContext:       nil,
		ConnContext:       nil,
	}

	print("server run on port ", server.Addr)

	_ = brock.Must(ok, server.ListenAndServe())
}
