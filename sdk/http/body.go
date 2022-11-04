package sdkhttp

import (
	"bytes"
	"io"

	"github.com/brick-io/brock/sdk"
)

//nolint:gochecknoglobals
var Body body

type body struct{}

// Create ...
func (body) Create(opts func() io.Reader) io.ReadCloser {
	return io.NopCloser(opts())
}

// WithBytes ...
func (body) WithBytes(v []byte) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBuffer(v)
	}
}

// WithString ...
func (body) WithString(v string) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBufferString(v)
	}
}

// WithJSON ...
func (body) WithJSON(v any) func() io.Reader {
	return func() io.Reader {
		buf := new(bytes.Buffer)
		_ = sdk.JSON.NewEncoder(buf).Encode(v)

		return buf
	}
}
