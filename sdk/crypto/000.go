package sdkcrypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"strings"

	"github.com/brick-io/brock/sdk"
)

var (
	ErrCryptoInvalidPEMFormat    = sdk.Errorf("brock: crypto: invalid pem format")
	ErrCryptoInvalidKeypair      = sdk.Errorf("brock: crypto: invalid keypair")
	ErrCryptoUnsupportedKeyTypes = sdk.Errorf("brock: crypto: unsupported key type")
)

func Nonce(n int) []byte {
	nonce := make([]byte, n)
	_, _ = io.ReadFull(rand.Reader, nonce)

	return nonce
}

type allKeyTypes interface {
	privateKeyTypes | publicKeyTypes
}
type privateKeyTypes interface {
	*rsa.PrivateKey | *ecdsa.PrivateKey | ed25519.PrivateKey
}
type publicKeyTypes interface {
	*rsa.PublicKey | *ecdsa.PublicKey | ed25519.PublicKey
}

func read[T allKeyTypes](r io.Reader) (T, error) {
	buf := &bytes.Buffer{}

	_, err := io.Copy(buf, io.LimitReader(r, (1e9)))
	if err != nil {
		return nil, err
	}

	p, rest := pem.Decode(buf.Bytes())
	if p == nil {
		err = sdk.Errorf("%w: pem:[%v] rest:[%s]", ErrCryptoInvalidPEMFormat, p, rest)

		return nil, err
	}

	var k0 any

	switch {
	case strings.Contains(p.Type, "PRIVATE KEY"):
		k0, err = x509.ParsePKCS8PrivateKey(p.Bytes)
	case strings.Contains(p.Type, "PUBLIC KEY"):
		k0, err = x509.ParsePKIXPublicKey(p.Bytes)
	}

	if err != nil {
		return nil, err
	}

	k, ok := k0.(T)
	if !ok || k == nil {
		return nil, sdk.Errorf("%w: %T", ErrCryptoUnsupportedKeyTypes, k0)
	}

	return k, nil
}

func write[T allKeyTypes](w io.Writer, k T) error {
	var (
		kt  string
		err error
		p   []byte
	)

	switch any(k).(type) {
	case *rsa.PrivateKey:
		kt = "RSA PRIVATE KEY"
		p, err = x509.MarshalPKCS8PrivateKey(k)
	case *ecdsa.PrivateKey:
		kt = "EC PRIVATE KEY"
		p, err = x509.MarshalPKCS8PrivateKey(k)
	case ed25519.PrivateKey:
		kt = "OPENSSH PRIVATE KEY"
		p, err = x509.MarshalPKCS8PrivateKey(k)
	case *rsa.PublicKey:
		kt = "RSA PUBLIC KEY"
		p, err = x509.MarshalPKIXPublicKey(k)
	case *ecdsa.PublicKey:
		kt = "EC PUBLIC KEY"
		p, err = x509.MarshalPKIXPublicKey(k)
	case ed25519.PublicKey:
		kt = "OPENSSH PUBLIC KEY"
		p, err = x509.MarshalPKIXPublicKey(k)
	default:
		err = sdk.Errorf("%w: %T", ErrCryptoUnsupportedKeyTypes, k)
	}

	if err != nil {
		return err
	}

	return pem.Encode(w, &pem.Block{Type: kt, Bytes: p})
}
