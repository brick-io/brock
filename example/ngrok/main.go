package main

import (
	"crypto/rand"
	"io"
	"net/http"

	"go.onebrick.io/brock"
)

func main() {
	nonce := brock.Sprintf("%x", Nonce((24)))
	srv := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				p, err := io.ReadAll(r.Body)
				_, _ = brock.Println("----:----------------------------------------------")
				_, _ = brock.Println("ERRS:", err)
				_, _ = brock.Println("HEAD:", r.Header)
				_, _ = brock.Println("BODY:", string(p))
			}
			ok := http.StatusOK
			http.Error(w, http.StatusText(ok)+"with nonce="+nonce, ok)
		}),
	}
	_, _ = brock.Printf("running on %s with nonce=%s", srv.Addr, nonce)
	_, _ = brock.Println(srv.ListenAndServe())
}

func Nonce(n int) []byte {
	nonce := make([]byte, n)
	_, _ = io.ReadFull(rand.Reader, nonce)

	return nonce
}

// c5d3c974bf85c2472b1d56facaf4a527282f2126fc427104
