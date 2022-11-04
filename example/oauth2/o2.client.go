package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"

	sdkcrypto "github.com/brick-io/brock/sdk/crypto"
	sdkotel "github.com/brick-io/brock/sdk/otel"
)

func doclient() {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     pub_client_b64,
		ClientSecret: pvt_client_b64,
		Scopes:       []string{"SCOPE1", "SCOPE2"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:9096/oauth2/authorization",
			TokenURL: "http://localhost:9096/oauth2/access_token",
		},
		RedirectURL: "https://example.com/oauth2/myapp/callback",
	}
	log := sdkotel.Log(ctx, os.Stdout)

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(btoa(sdkcrypto.Nonce((24))), oauth2.AccessTypeOnline)
	log.Log.Printf("Visit the URL for the auth dialog: \n%v\n", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Log.Fatal(err)
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	_, _ = client.Get("...")
}
