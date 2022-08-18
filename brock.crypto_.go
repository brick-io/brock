package brock

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io"
	"strings"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

// nolint: gochecknoglobals
var (
	Crypto crypto
)

type crypto struct {
	RSA     crypto_rsa
	ECDSA   crypto_ecdsa
	ED25519 crypto_ed25519
	NaCl    crypto_nacl
}

func (crypto) Nonce(n int) []byte {
	nonce := make([]byte, n)
	_, _ = io.ReadFull(rand.Reader, nonce)

	return nonce
}

// =============================================================================

type crypto_rsa struct{}

func (crypto_rsa) Generate(bits int) (*rsa.PrivateKey, error) {
	switch bits {
	default:
		bits = 4096
	case (2048), (4096):
	}

	return rsa.GenerateKey(rand.Reader, bits)
}

func (crypto_rsa) ReadKeypair(keyReader, pubReader io.Reader) (*rsa.PrivateKey, error) {
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

func (crypto_rsa) WriteKeypair(keyWriter, pubWriter io.Writer, key *rsa.PrivateKey) error {
	err := crypto_rsa{}.Validate(key)
	if err == nil {
		err = write(keyWriter, key)
	}

	if err == nil {
		err = write(pubWriter, &key.PublicKey)
	}

	return err
}

func (crypto_rsa) Validate(key *rsa.PrivateKey) error {
	if key == nil {
		key = &rsa.PrivateKey{}
	}

	err := key.Validate()
	if err != nil {
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
		return Errorf("%w: msg:[%s] d1:[%s]", ErrCryptoInvalidKeypair, msg, d1)
	}

	e2, err := rsa.EncryptPKCS1v15(rand.Reader, &key.PublicKey, msg)
	if err != nil {
		return err
	}

	d2, err := rsa.DecryptPKCS1v15(rand.Reader, key, e2)
	if err != nil {
		return err
	} else if !bytes.Equal(msg, d2) {
		return Errorf("%w: msg:[%s] d2:[%s]", ErrCryptoInvalidKeypair, msg, d2)
	}

	return nil
}

// =============================================================================

type crypto_ecdsa struct{}

func (crypto_ecdsa) Generate(curv elliptic.Curve) (*ecdsa.PrivateKey, error) {
	if curv == nil {
		curv = elliptic.P521()
	}

	return ecdsa.GenerateKey(curv, rand.Reader)
}

func (crypto_ecdsa) ReadKeypair(keyReader, pubReader io.Reader) (*ecdsa.PrivateKey, error) {
	key, err := read[*ecdsa.PrivateKey](keyReader)
	if err != nil {
		return nil, err
	}

	pub, err := read[*ecdsa.PublicKey](pubReader)
	if err != nil {
		return nil, err
	}

	key.PublicKey = *pub

	return key, nil
}

func (crypto_ecdsa) WriteKeypair(keyWriter, pubWriter io.Writer, key *ecdsa.PrivateKey) error {
	err := crypto_ecdsa{}.Validate(key)
	if err == nil {
		err = write(keyWriter, key)
	}

	if err == nil {
		err = write(pubWriter, &key.PublicKey)
	}

	return err
}

func (crypto_ecdsa) Validate(key *ecdsa.PrivateKey) error {
	if key == nil {
		key = &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{}}
	}

	onCurve, h := true, sha256.Sum256(crypto{}.Nonce(24))
	onCurve = onCurve && key.Curve != nil
	onCurve = onCurve && key.PublicKey.Curve != nil
	onCurve = onCurve && key.IsOnCurve(key.X, key.Y)
	onCurve = onCurve && key.PublicKey.IsOnCurve(key.PublicKey.X, key.PublicKey.Y)

	if !onCurve {
		return Errorf("pub not on curve: key_x:[%v] key_y:[%v] pub_x:[%v] pub_y:[%v]",
			key.X, key.Y, key.PublicKey.X, key.PublicKey.Y,
		)
	}

	sig, err := ecdsa.SignASN1(rand.Reader, key, h[:])
	if err != nil {
		return err
	} else if !ecdsa.VerifyASN1(&key.PublicKey, h[:], sig) {
		return Errorf("mismatch: sig:[%s] ", sig)
	}

	r, s, err := ecdsa.Sign(rand.Reader, key, h[:])
	if err != nil {
		return err
	} else if !ecdsa.Verify(&key.PublicKey, h[:], r, s) {
		return Errorf("mismatch: r:[%v] s:[%v]", r, s)
	}

	return nil
}

// =============================================================================

type crypto_ed25519 struct{}

func (crypto_ed25519) Generate() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func (crypto_ed25519) ReadKeypair(keyReader, pubReader io.Reader) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	key, err := read[ed25519.PrivateKey](keyReader)
	if err != nil {
		return nil, nil, err
	}

	pub, err := read[ed25519.PublicKey](pubReader)
	if err != nil {
		return nil, nil, err
	}

	return key, pub, nil
}

func (crypto_ed25519) WriteKeypair(keyWriter, pubWriter io.Writer, key ed25519.PrivateKey, pub ed25519.PublicKey) error {
	err := crypto_ed25519{}.Validate(key, pub)
	if err == nil {
		err = write(keyWriter, key)
	}

	if err == nil {
		err = write(pubWriter, pub)
	}

	return err
}

func (crypto_ed25519) Validate(key ed25519.PrivateKey, pub ed25519.PublicKey) error {
	msg := crypto{}.Nonce(24)

	sig := ed25519.Sign(key, msg)
	if !ed25519.Verify(pub, msg, sig) {
		return ErrCryptoInvalidKeypair
	}

	return nil
}

// =============================================================================

type crypto_nacl struct {
	Box       crypto_nacl_box
	SecretBox crypto_nacl_secretbox
}

func (crypto_nacl) extract(ciphertext []byte) ([]byte, [24]byte) {
	var nonce [24]byte

	_, rest := copy(nonce[:], ciphertext[:24]), ciphertext[24:]

	return rest, nonce
}

func (crypto_nacl) nonce() [24]byte {
	var nonce [24]byte

	_, _ = io.ReadFull(rand.Reader, nonce[:])

	return nonce
}

func (crypto_nacl) secretHash(secret []byte) *[32]byte {
	var out [32]byte

	hash := sha256.New()
	_, _ = hash.Write(secret)
	_ = copy(out[:], hash.Sum(nil))

	return &out
}

type crypto_nacl_box struct{}

// nolint: nonamedreturns
func (crypto_nacl_box) Generate() (publicKey, privateKey *[32]byte, err error) {
	return box.GenerateKey(rand.Reader)
}

// nolint: nonamedreturns
func (crypto_nacl_box) Seal(plaintext []byte, peersPublicKey, privateKey *[32]byte) (ciphertext []byte) {
	nonce := crypto_nacl{}.nonce()

	return box.Seal(nonce[:], plaintext, &nonce, peersPublicKey, privateKey)
}

// nolint: nonamedreturns
func (crypto_nacl_box) Open(ciphertext []byte, peersPublicKey, privateKey *[32]byte) (plaintext []byte, ok bool) {
	rest, nonce := crypto_nacl{}.extract(ciphertext)

	return box.Open(nil, rest, &nonce, peersPublicKey, privateKey)
}

// nolint: nonamedreturns
func (crypto_nacl_box) SealWithSharedKey(plaintext []byte, sharedKey *[32]byte) (ciphertext []byte) {
	nonce := crypto_nacl{}.nonce()

	return box.SealAfterPrecomputation(nonce[:], plaintext, &nonce, sharedKey)
}

// nolint: nonamedreturns
func (crypto_nacl_box) OpenWithSharedKey(ciphertext []byte, sharedKey *[32]byte) (plaintext []byte, ok bool) {
	rest, nonce := crypto_nacl{}.extract(ciphertext)

	return box.OpenAfterPrecomputation(nil, rest, &nonce, sharedKey)
}

// nolint: nonamedreturns
func (crypto_nacl_box) SharedKey(peersPublicKey, privateKey *[32]byte) (_ *[32]byte) {
	var sharedKey [32]byte

	box.Precompute(&sharedKey, peersPublicKey, privateKey)

	return &sharedKey
}

type crypto_nacl_secretbox struct{}

// nolint: nonamedreturns
func (crypto_nacl_secretbox) Seal(plaintext, secret []byte) (ciphertext []byte) {
	nonce := crypto_nacl{}.nonce()

	return secretbox.Seal(nonce[:], plaintext, &nonce, crypto_nacl{}.secretHash(secret))
}

// nolint: nonamedreturns
func (crypto_nacl_secretbox) Open(ciphertext, secret []byte) (plaintext []byte, ok bool) {
	rest, nonce := crypto_nacl{}.extract(ciphertext)

	return secretbox.Open(nil, rest, &nonce, crypto_nacl{}.secretHash(secret))
}

// =============================================================================

type crypto_key_types interface {
	crypto_private_key_types | crypto_public_key_types
}
type crypto_private_key_types interface {
	*rsa.PrivateKey | *ecdsa.PrivateKey | ed25519.PrivateKey
}
type crypto_public_key_types interface {
	*rsa.PublicKey | *ecdsa.PublicKey | ed25519.PublicKey
}

func read[T crypto_key_types](r io.Reader) (T, error) {
	buf := &bytes.Buffer{}

	_, err := io.Copy(buf, io.LimitReader(r, 1e9))
	if err != nil {
		return nil, err
	}

	p, rest := pem.Decode(buf.Bytes())
	if p == nil {
		err = Errorf("%w: pem:[%v] rest:[%s]", ErrCryptoInvalidPEMFormat, p, rest)

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
		return nil, Errorf("%w: %T", ErrCryptoUnsupportedKeyTypes, k0)
	}

	return k, nil
}

func write[T crypto_key_types](w io.Writer, k T) error {
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
		err = Errorf("%w: %T", ErrCryptoUnsupportedKeyTypes, k)
	}

	if err != nil {
		return err
	}

	return pem.Encode(w, &pem.Block{Type: kt, Bytes: p})
}
