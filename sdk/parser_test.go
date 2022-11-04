package sdk_test

import (
	"bytes"
	"testing"

	. "github.com/brick-io/brock/sdk"
	. "github.com/onsi/gomega"
)

//nolint:funlen
func Test_sdkparser(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		type root struct {
			A string `json:"a"`
			B string `json:"b"`
		}
		const raw = `{"a":"one","b":"two"}` + "\n"

		t.Run("Encode", func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := JSON.NewEncoder(buf).Encode(root{A: "one", B: "two"})
			Expect(err).To(Succeed())
			Expect(buf.String()).To(Equal(raw))
		})

		t.Run("Decode", func(t *testing.T) {
			var root root
			err := JSON.NewDecoder(bytes.NewBufferString(raw)).Decode(&root)
			Expect(err).To(Succeed())
			Expect(root.A).To(Equal("one"))
			Expect(root.B).To(Equal("two"))
		})
	})

	t.Run("XML", func(t *testing.T) {
		t.Parallel()
		type root struct {
			XMLName XMLName `xml:"root"`
			A       string  `xml:"a"`
			B       string  `xml:"b"`
		}
		const raw = `<root><a>one</a><b>two</b></root>`

		t.Run("Marshal", func(t *testing.T) {
			t.Parallel()
			p, err := XML.Marshal(root{A: "one", B: "two"})
			Expect(err).To(Succeed())
			Expect(string(p)).To(Equal(raw))
		})

		t.Run("Unmarshal", func(t *testing.T) {
			t.Parallel()
			var root root
			err := XML.Unmarshal([]byte(raw), &root)
			Expect(err).To(Succeed())
			Expect(root.A).To(Equal("one"))
			Expect(root.B).To(Equal("two"))
		})

		t.Run("Encode", func(t *testing.T) {
			t.Parallel()
			buf := new(bytes.Buffer)
			err := XML.NewEncoder(buf).Encode(root{A: "one", B: "two"})
			Expect(err).To(Succeed())
			Expect(buf.String()).To(Equal(raw))
		})

		t.Run("Decode", func(t *testing.T) {
			t.Parallel()
			var root root
			err := XML.NewDecoder(bytes.NewBufferString(raw)).Decode(&root)
			Expect(err).To(Succeed())
			Expect(root.A).To(Equal("one"))
			Expect(root.B).To(Equal("two"))
		})
	})

	t.Run("YAML", func(t *testing.T) { t.Parallel() })

	t.Run("TOML", func(t *testing.T) { t.Parallel() })
}
