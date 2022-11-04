package sdkcrypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"
)

//nolint:gochecknoglobals
var ED25519 wrapED25519

type wrapED25519 struct{}

func (wrapED25519) Generate() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func (wrapED25519) ReadKeypair(keyR, pubR io.Reader) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	key, err := read[ed25519.PrivateKey](keyR)
	if err != nil {
		return nil, nil, err
	}

	pub, err := read[ed25519.PublicKey](pubR)
	if err != nil {
		return nil, nil, err
	}

	return key, pub, nil
}

func (wrapED25519) WriteKeypair(keyW, pubW io.Writer, key ed25519.PrivateKey, pub ed25519.PublicKey) error {
	err := wrapED25519{}.Validate(key, pub)
	if err == nil {
		err = write(keyW, key)
	}

	if err == nil {
		err = write(pubW, pub)
	}

	return err
}

func (wrapED25519) Validate(key ed25519.PrivateKey, pub ed25519.PublicKey) error {
	msg := Nonce((24))

	sig := ed25519.Sign(key, msg)
	if !ed25519.Verify(pub, msg, sig) {
		return ErrCryptoInvalidKeypair
	}

	return nil
}
