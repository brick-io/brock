package sdkcrypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"

	"github.com/brick-io/brock/sdk"
)

//nolint:gochecknoglobals
var RSA wrapRSA

type wrapRSA struct{}

func (wrapRSA) Generate(bits int) (*rsa.PrivateKey, error) {
	switch bits {
	default:
		bits = 4096
	case (2048), (4096):
	}

	return rsa.GenerateKey(rand.Reader, bits)
}

func (wrapRSA) ReadKeypair(keyReader, pubReader io.Reader) (*rsa.PrivateKey, error) {
	key, err := read[*rsa.PrivateKey](keyReader)
	if err != nil {
		return nil, err
	}

	pub, err := read[*rsa.PublicKey](pubReader)
	if err != nil {
		return nil, err
	}

	key.PublicKey = *pub

	return key, nil
}

func (wrapRSA) WriteKeypair(keyWriter, pubWriter io.Writer, key *rsa.PrivateKey) error {
	err := wrapRSA{}.Validate(key)
	if err == nil {
		err = write(keyWriter, key)
	}

	if err == nil {
		err = write(pubWriter, &key.PublicKey)
	}

	return err
}

func (wrapRSA) Validate(key *rsa.PrivateKey) error {
	if key == nil {
		key = &rsa.PrivateKey{}
	}

	if err := key.Validate(); err != nil {
		return err
	}

	hash, msg, lbl := sha256.New(), []byte("test valid"), []byte("label 123")

	e1, err := rsa.EncryptOAEP(hash, rand.Reader, &key.PublicKey, msg, lbl)
	if err != nil {
		return err
	}

	d1, err := rsa.DecryptOAEP(hash, rand.Reader, key, e1, lbl)
	if err != nil {
		return err
	} else if !bytes.Equal(msg, d1) {
		return sdk.Errorf("%w: msg:[%s] d1:[%s]", ErrCryptoInvalidKeypair, msg, d1)
	}

	e2, err := rsa.EncryptPKCS1v15(rand.Reader, &key.PublicKey, msg)
	if err != nil {
		return err
	}

	d2, err := rsa.DecryptPKCS1v15(rand.Reader, key, e2)
	if err != nil {
		return err
	} else if !bytes.Equal(msg, d2) {
		return sdk.Errorf("%w: msg:[%s] d2:[%s]", ErrCryptoInvalidKeypair, msg, d2)
	}

	return nil
}
