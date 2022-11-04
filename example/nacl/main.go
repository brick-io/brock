package main

import (
	"context"
	"encoding/base64"
	"os"

	sdkcrypto "github.com/brick-io/brock/sdk/crypto"
	sdkotel "github.com/brick-io/brock/sdk/otel"
)

func main() {
	ctx := context.Background()
	log := sdkotel.Log(ctx, os.Stdout)
	cipher, _ := atob("J9w+EjXvsM2B1zws+75cp6GPWwz8T5M6IjJELQlBDGE4jRXrsYgfYbL5VIE/JwaOC/qFHRDxVi/tN3qE9L6uTXHUA4maM3" +
		"5bYvymyTkwGyLCm2JrrVGKBsLAUMCs9S39qmwAkotG1B8")
	key, _ := atob("T+YevCYFJ2+lAXdnxmqn3cLkiXUrRYIEP9VkDSl/lGQ")

	log.Log.Print("cipher: ", len(cipher), cipher)
	log.Log.Print("key: ", len(key), key)

	var k [32]byte

	copy(k[:], key)

	plain, ok := sdkcrypto.NaCl.Box.OpenWithSharedKey(cipher, &k)
	log.Log.Print("===\n", ok, string(plain), k[:])
}

//nolint:gochecknoglobals
var atob = base64.RawStdEncoding.DecodeString
