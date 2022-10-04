package main

import (
	"context"
	"fmt"
	"log"

	"go.onebrick.io/brock"
	"golang.org/x/oauth2"
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

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(btoa(brock.Crypto.Nonce(24)), oauth2.AccessTypeOnline)
	fmt.Printf("Visit the URL for the auth dialog: \n%v\n", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	client.Get("...")
}
