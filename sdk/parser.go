package sdk

import (
	"bytes"
	"encoding/xml"
	"io"

	toml "github.com/BurntSushi/toml"
	json "github.com/json-iterator/go"
	yaml "gopkg.in/yaml.v3"
)

type Decoder interface{ Decode(v any) error }

type Encoder interface{ Encode(v any) error }

type Parser interface {
	Marshal(v any) (p []byte, err error)
	Unmarshal(p []byte, v any) (err error)
	NewEncoder(w io.Writer) Encoder
	NewDecoder(r io.Reader) Decoder
}

//nolint:gochecknoglobals
var (
	// JSON ...
	JSON Parser = _json{json.ConfigFastest}
	// XML ...
	XML Parser = _xml{}
	// YAML ...
	YAML Parser = _yaml{}
	// TOML ...
	TOML Parser = _toml{}
)

// =============================================================================

// _json implementation.
type _json struct{ json.API }

func (json _json) NewDecoder(r io.Reader) Decoder { return json.API.NewDecoder(r) }

func (json _json) NewEncoder(w io.Writer) Encoder { return json.API.NewEncoder(w) }

// =============================================================================

type XMLName = xml.Name

// _xml implementation.
type _xml struct{}

func (_xml) Marshal(v any) ([]byte, error) { return xml.Marshal(v) }

func (_xml) Unmarshal(data []byte, v any) error { return xml.Unmarshal(data, v) }

func (_xml) NewEncoder(w io.Writer) Encoder { return xml.NewEncoder(w) }

func (_xml) NewDecoder(r io.Reader) Decoder { return xml.NewDecoder(r) }

// =============================================================================

// _yaml implementation.
type _yaml struct{}

func (_yaml) Marshal(v any) ([]byte, error) { return yaml.Marshal(v) }

func (_yaml) Unmarshal(data []byte, v any) error { return yaml.Unmarshal(data, v) }

func (_yaml) NewEncoder(w io.Writer) Encoder { return yaml.NewEncoder(w) }

func (_yaml) NewDecoder(r io.Reader) Decoder { return yaml.NewDecoder(r) }

// =============================================================================

// _toml implementation.
type _toml struct{ r io.Reader }

func (_toml) Marshal(v any) ([]byte, error) {
	b := new(bytes.Buffer)
	err := toml.NewEncoder(b).Encode(v)

	return b.Bytes(), err
}

func (_toml) Unmarshal(data []byte, v any) error { return toml.Unmarshal(data, v) }

func (_toml) NewEncoder(w io.Writer) Encoder { return toml.NewEncoder(w) }

func (_toml) NewDecoder(r io.Reader) Decoder { return _toml{r} }

func (t _toml) Decode(v any) error {
	if p, err := io.ReadAll(t.r); err != nil {
		return err
	} else {
		return toml.Unmarshal(p, v)
	}
}
