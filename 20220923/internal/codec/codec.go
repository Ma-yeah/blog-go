package codec

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/text/encoding/unicode"
	"io"
)

var Chinese = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)

type Unmarshaler interface {
	Unmarshal([]byte) error
}

func Unmarshal(b []byte, i any) error {
	switch i.(type) {
	case Unmarshaler:
		return i.(Unmarshaler).Unmarshal(b)
	default:
		return binary.Read(bytes.NewReader(b), binary.LittleEndian, i) // binary.BigEndian
	}
}

type Marshaler interface {
	Marshal() ([]byte, error)
}

func Marshal(i any) ([]byte, error) {
	switch i.(type) {
	case Marshaler:
		return i.(Marshaler).Marshal()
	default:
		var p bytes.Buffer
		if err := binary.Write(&p, binary.LittleEndian, i); err != nil {
			return nil, err
		}
		return p.Bytes(), nil
	}
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

func (e *Encoder) Encode(data ...any) error {
	for _, v := range data {
		if err := binary.Write(e.w, binary.LittleEndian, v); err != nil {
			return err
		}
	}
	return nil
}

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (d *Decoder) Decode(data ...any) error {
	for _, v := range data {
		if err := binary.Read(d.r, binary.LittleEndian, v); err != nil {
			return err
		}
	}
	return nil
}
