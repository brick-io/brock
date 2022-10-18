package main

import (
	"encoding/base64"

	"go.onebrick.io/brock"
)

func main() {
	cipher, _ := atob("J9w+EjXvsM2B1zws+75cp6GPWwz8T5M6IjJELQlBDGE4jRXrsYgfYbL5VIE/JwaOC/qFHRDxVi/tN3qE9L6uTXHUA4maM35bYvymyTkwGyLCm2JrrVGKBsLAUMCs9S39qmwAkotG1B8")
	key, _ := atob("T+YevCYFJ2+lAXdnxmqn3cLkiXUrRYIEP9VkDSl/lGQ")
	_, _ = brock.Println("cipher: ", len(cipher), cipher)
	_, _ = brock.Println("key: ", len(key), key)
	var k [32]byte
	copy(k[:], key)
	plain, ok := brock.Crypto.NaCl.Box.OpenWithSharedKey(cipher, &k)
	_, _ = brock.Println("===\n", ok, string(plain), k[:])
}

var (
	atob = base64.RawStdEncoding.DecodeString
)
