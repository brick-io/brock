package main

import (
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"os"

	"github.com/brick-io/brock/sdk"
	sdkotel "github.com/brick-io/brock/sdk/otel"
)

func main() {
	ctx := context.Background()
	log := sdkotel.Log(ctx, os.Stdout)

	nonce := sdk.Sprintf("%x", Nonce((24)))
	srv := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				p, err := io.ReadAll(r.Body)
				log.Log.Print("----:----------------------------------------------")
				log.Log.Print("ERRS:", err)
				log.Log.Print("HEAD:", r.Header)
				log.Log.Print("BODY:", string(p))
			}
			ok := http.StatusOK
			http.Error(w, http.StatusText(ok)+"with nonce="+nonce, ok)
		}),
	}
	log.Log.Printf("running on %s with nonce=%s", srv.Addr, nonce)
	log.Log.Print(srv.ListenAndServe())
}

func Nonce(n int) []byte {
	nonce := make([]byte, n)
	_, _ = io.ReadFull(rand.Reader, nonce)

	return nonce
}

// c5d3c974bf85c2472b1d56facaf4a527282f2126fc427104
