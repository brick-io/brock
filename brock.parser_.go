package brock

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

//nolint: gochecknoglobals
var (
	// JSON
	JSON Parser = brock_json{json.ConfigFastest}
	// XML
	XML Parser = brock_xml{}
	// YAML
	YAML Parser = brock_yaml{}
	// TOML
	TOML Parser = brock_toml{}
)

// =============================================================================

// brock_json implementation.
type brock_json struct{ json.API }

func (json brock_json) NewDecoder(r io.Reader) Decoder { return json.API.NewDecoder(r) }

func (json brock_json) NewEncoder(w io.Writer) Encoder { return json.API.NewEncoder(w) }

// =============================================================================

type XMLName = xml.Name

// brock_xml implementation.
type brock_xml struct{}

func (brock_xml) Marshal(v any) ([]byte, error) { return xml.Marshal(v) }

func (brock_xml) Unmarshal(data []byte, v any) error { return xml.Unmarshal(data, v) }

func (brock_xml) NewEncoder(w io.Writer) Encoder { return xml.NewEncoder(w) }

func (brock_xml) NewDecoder(r io.Reader) Decoder { return xml.NewDecoder(r) }

// =============================================================================

// brock_yaml implementation.
type brock_yaml struct{}

func (brock_yaml) Marshal(v any) ([]byte, error) { return yaml.Marshal(v) }

func (brock_yaml) Unmarshal(data []byte, v any) error { return yaml.Unmarshal(data, v) }

func (brock_yaml) NewEncoder(w io.Writer) Encoder { return yaml.NewEncoder(w) }

func (brock_yaml) NewDecoder(r io.Reader) Decoder { return yaml.NewDecoder(r) }

// =============================================================================

// brock_toml implementation.
type brock_toml struct{ r io.Reader }

func (brock_toml) Marshal(v any) ([]byte, error) {
	b := new(bytes.Buffer)
	err := toml.NewEncoder(b).Encode(v)

	return b.Bytes(), err
}

func (brock_toml) Unmarshal(data []byte, v any) error { return toml.Unmarshal(data, v) }

func (brock_toml) NewEncoder(w io.Writer) Encoder { return toml.NewEncoder(w) }

func (brock_toml) NewDecoder(r io.Reader) Decoder { return brock_toml{r} }

func (t brock_toml) Decode(v any) error {
	if p, err := io.ReadAll(t.r); err != nil {
		return err
	} else {
		return toml.Unmarshal(p, v)
	}
}
