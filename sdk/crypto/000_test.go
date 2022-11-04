package sdkcrypto_test

import (
	"testing"

	. "github.com/onsi/gomega"

	sdkcrypto "github.com/brick-io/brock/sdk/crypto"
)

func Test_sdkcrypto(t *testing.T) {
	const (
		msg = "To infinity, and beyond!"
		sig = "Buzz Lightyear, not Aldrin"
	)

	Expect := NewWithT(t).Expect

	t.Run("NaCl", func(t *testing.T) {
		t.Run("box", func(t *testing.T) {
			t.Parallel()
			box := sdkcrypto.NaCl.Box

			// 1st person
			pub1, key1, err := box.Generate()
			Expect(err).To(Succeed())

			// 2nd person
			pub2, key2, err := box.Generate()
			Expect(err).To(Succeed())

			// transfer to 2nd person
			ciphertext := box.Seal([]byte(msg), pub2, key1)
			plaintext, ok := box.Open(ciphertext, pub1, key2)
			Expect(ok).To(BeTrue())
			Expect(string(plaintext)).To(Equal(msg))

			sk12 := box.SharedKey(pub1, key2)
			plaintext, ok = box.OpenWithSharedKey(ciphertext, sk12)
			Expect(ok).To(BeTrue())
			Expect(string(plaintext)).To(Equal(msg))

			ciphertext = box.SealWithSharedKey([]byte(msg), sk12)
			plaintext, ok = box.Open(ciphertext, pub2, key1)
			Expect(ok).To(BeTrue())
			Expect(string(plaintext)).To(Equal(msg))
		})
		t.Run("secretbox", func(t *testing.T) {
			t.Parallel()
			secretbox := sdkcrypto.NaCl.SecretBox

			ciphertext := secretbox.Seal([]byte(msg), []byte(sig))
			plaintext, ok := secretbox.Open(ciphertext, []byte(sig))
			Expect(ok).To(BeTrue())
			Expect(string(plaintext)).To(Equal(msg))
		})
	})
}
