package sdkcrypto

import (
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

//nolint:gochecknoglobals
var NaCl wrapNaCl

type wrapNaCl struct {
	Box       wrapNaClBox
	SecretBox wrapNaClSecretBox
}

func (wrapNaCl) extract(ciphertext []byte) ([]byte, [24]byte) {
	var nonce [24]byte

	_, rest := copy(nonce[:], ciphertext[:24]), ciphertext[24:]

	return rest, nonce
}

func (wrapNaCl) nonce() [24]byte {
	var nonce [24]byte

	_, _ = io.ReadFull(rand.Reader, nonce[:])

	return nonce
}

func (wrapNaCl) secretHash(secret []byte) *[32]byte {
	var out [32]byte

	hash := sha256.New()
	_, _ = hash.Write(secret)
	_ = copy(out[:], hash.Sum(nil))

	return &out
}

type wrapNaClBox struct{}

func (wrapNaClBox) Generate() (publicKey, privateKey *[32]byte, err error) {
	return box.GenerateKey(rand.Reader)
}

func (wrapNaClBox) Seal(plaintext []byte, peersPublicKey, privateKey *[32]byte) (ciphertext []byte) {
	nonce := NaCl.nonce()

	return box.Seal(nonce[:], plaintext, &nonce, peersPublicKey, privateKey)
}

func (wrapNaClBox) Open(ciphertext []byte, peersPublicKey, privateKey *[32]byte) (plaintext []byte, ok bool) {
	rest, nonce := NaCl.extract(ciphertext)

	return box.Open(nil, rest, &nonce, peersPublicKey, privateKey)
}

func (wrapNaClBox) SealWithSharedKey(plaintext []byte, sharedKey *[32]byte) (ciphertext []byte) {
	nonce := NaCl.nonce()

	return box.SealAfterPrecomputation(nonce[:], plaintext, &nonce, sharedKey)
}

func (wrapNaClBox) OpenWithSharedKey(ciphertext []byte, sharedKey *[32]byte) (plaintext []byte, ok bool) {
	rest, nonce := NaCl.extract(ciphertext)

	return box.OpenAfterPrecomputation(nil, rest, &nonce, sharedKey)
}

func (wrapNaClBox) SharedKey(peersPublicKey, privateKey *[32]byte) (_ *[32]byte) {
	var sharedKey [32]byte

	box.Precompute(&sharedKey, peersPublicKey, privateKey)

	return &sharedKey
}

type wrapNaClSecretBox struct{}

func (wrapNaClSecretBox) Seal(plaintext, secret []byte) (ciphertext []byte) {
	nonce := NaCl.nonce()

	return secretbox.Seal(nonce[:], plaintext, &nonce, NaCl.secretHash(secret))
}

func (wrapNaClSecretBox) Open(ciphertext, secret []byte) (plaintext []byte, ok bool) {
	rest, nonce := NaCl.extract(ciphertext)

	return secretbox.Open(nil, rest, &nonce, NaCl.secretHash(secret))
}
