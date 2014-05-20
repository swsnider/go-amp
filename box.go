package amp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf8"
)

type Box map[string]string

const (
	MAX_KEY_LENGTH = 255
	MAX_VAL_LENGTH = 65535
)

// Serialize converts an AMP Box to the wire representation, if possible.
func (b Box) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	for k, v := range b {
		if !utf8.ValidString(k) {
			return nil, fmt.Errorf("key %q is not valid utf8", k)
		}
		if !utf8.ValidString(v) {
			return nil, fmt.Errorf("value %q is not valid utf8", k)
		}
		if len(k) > MAX_KEY_LENGTH {
			return nil, fmt.Errorf("key %q has length %v which is longer than max length %q", k, len(k), MAX_KEY_LENGTH)
		}
		if len(v) > MAX_VAL_LENGTH {
			return nil, fmt.Errorf("value %q has length %v which is longer than max length %q", v, len(v), MAX_VAL_LENGTH)
		}

		binary.Write(buf, binary.BigEndian, uint16(len(k)))
		binary.Write(buf, binary.BigEndian, []byte(k))
		binary.Write(buf, binary.BigEndian, uint16(len(v)))
		binary.Write(buf, binary.BigEndian, []byte(v))
	}
	buf.WriteString("\x00\x00")
	return buf.Bytes(), nil
}

// Deserialize converts a wire representation to an AMP box, if possible.
func Deserialize(b []byte) (Box, error) {
	if len(b) < 2 {
		return nil, fmt.Errorf("serialized AMP box length was %v, shorter than two bytes, which is illegal", len(b))
	}
	buf := bytes.NewBuffer(b)
	box := make(Box)
	for {
		var l uint16
		err := binary.Read(buf, binary.BigEndian, &l)
		if err != nil {
			return nil, fmt.Errorf("unable to read length of key from buffer: %v", err)
		}
		if l == 0 {
			break
		}
		kbytes := make([]byte, l)
		err = binary.Read(buf, binary.BigEndian, kbytes)
		if err != nil {
			return nil, fmt.Errorf("unable to read key of length %v from buffer: %v", l, err)
		}
		if !utf8.Valid(kbytes) {
			return nil, fmt.Errorf("key %q is an invalid utf-8 string", kbytes)
		}
		k := string(kbytes)

		err = binary.Read(buf, binary.BigEndian, &l)
		if err != nil {
			return nil, fmt.Errorf("unable to read length of value from buffer: %v", err)
		}
		vbytes := make([]byte, l)
		err = binary.Read(buf, binary.BigEndian, vbytes)
		if err != nil {
			return nil, fmt.Errorf("unable to read value of length %v from buffer: %v", l, err)
		}
		if !utf8.Valid(vbytes) {
			return nil, fmt.Errorf("value %q is an invalid utf-8 string", vbytes)
		}
		box[k] = string(vbytes)
	}
	return box, nil
}
