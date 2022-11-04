package sdkhttp

import (
	"io"
	"mime/multipart"

	"github.com/brick-io/brock/sdk"
)

//nolint:gochecknoglobals
var MultipartForm multipartForm

type multipartForm struct{}

func (multipartForm) Create(opts ...func(*multipart.Writer)) *multipart.Writer {
	return sdk.Apply(new(multipart.Writer), opts...)
}

func (multipartForm) WithWriter(w io.Writer) func(*multipart.Writer) {
	return func(mw *multipart.Writer) {
		*mw = *(multipart.NewWriter(w))
	}
}

func (multipartForm) WithField(key, value string) func(*multipart.Writer) {
	return func(mw *multipart.Writer) {
		_ = mw.WriteField(key, value)
	}
}

func (multipartForm) WithFile(key, filename string, r io.Reader) func(*multipart.Writer) {
	return func(mw *multipart.Writer) {
		w, err := mw.CreateFormFile(key, filename)
		if err == nil {
			_, _ = io.Copy(w, r)
		}
	}
}
