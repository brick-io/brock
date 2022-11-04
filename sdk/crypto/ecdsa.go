package sdkcrypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/brick-io/brock/sdk"
)

//nolint:gochecknoglobals
var ECDSA wrapECDSA

type wrapECDSA struct{}

func (wrapECDSA) Generate(curv elliptic.Curve) (*ecdsa.PrivateKey, error) {
	if curv == nil {
		curv = elliptic.P521()
	}

	return ecdsa.GenerateKey(curv, rand.Reader)
}

func (wrapECDSA) ReadKeypair(keyReader, pubReader io.Reader) (*ecdsa.PrivateKey, error) {
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

func (wrapECDSA) WriteKeypair(keyWriter, pubWriter io.Writer, key *ecdsa.PrivateKey) error {
	err := wrapECDSA{}.Validate(key)
	if err == nil {
		err = write(keyWriter, key)
	}

	if err == nil {
		err = write(pubWriter, &key.PublicKey)
	}

	return err
}

func (wrapECDSA) Validate(key *ecdsa.PrivateKey) error {
	if key == nil {
		key = &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{}}
	}

	onCurve, h := true, sha256.Sum256(Nonce((24)))
	onCurve = onCurve && key.Curve != nil
	onCurve = onCurve && key.PublicKey.Curve != nil
	onCurve = onCurve && key.IsOnCurve(key.X, key.Y)
	onCurve = onCurve && key.PublicKey.IsOnCurve(key.PublicKey.X, key.PublicKey.Y)

	if !onCurve {
		return sdk.Errorf("pub not on curve: key_x:[%v] key_y:[%v] pub_x:[%v] pub_y:[%v]",
			key.X, key.Y, key.PublicKey.X, key.PublicKey.Y,
		)
	}

	sig, err := ecdsa.SignASN1(rand.Reader, key, h[:])
	if err != nil {
		return err
	} else if !ecdsa.VerifyASN1(&key.PublicKey, h[:], sig) {
		return sdk.Errorf("mismatch: sig:[%s] ", sig)
	}

	r, s, err := ecdsa.Sign(rand.Reader, key, h[:])
	if err != nil {
		return err
	} else if !ecdsa.Verify(&key.PublicKey, h[:], r, s) {
		return sdk.Errorf("mismatch: r:[%v] s:[%v]", r, s)
	}

	return nil
}
