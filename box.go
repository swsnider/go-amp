package amp

import (
  "bytes"
  "encoding/binary"
  "fmt"
  "io"
  "reflect"
  "strconv"
  "unicode/utf8"
)

type Box map[string]string

const (
  MAX_KEY_LENGTH = 255
  MAX_VAL_LENGTH = 65535
)

func writeKV(buf *bytes.Buffer, k string, v string) error {
  if err := binary.Write(buf, binary.BigEndian, uint16(len(k))); err != nil {
    return fmt.Errorf("unable to write length for key %q: %v", k, err)
  }
  if err := binary.Write(buf, binary.BigEndian, []byte(k)); err != nil {
    return fmt.Errorf("unable to write key %q: %v", k, err)
  }
  if err := binary.Write(buf, binary.BigEndian, uint16(len(v))); err != nil {
    return fmt.Errorf("unable to write length of value for key %q: %v", k, err)
  }
  if err := binary.Write(buf, binary.BigEndian, []byte(v)); err != nil {
    return fmt.Errorf("unable to write value for key %q: %v", k, err)
  }
  return nil
}

// Serialize converts an AMP Box to the wire representation, if possible.
func (b Box) Serialize() ([]byte, error) {
  var err error
  buf := new(bytes.Buffer)

  if ask, ok := b["_ask"]; ok {
    err = writeKV(buf, "_ask", ask)
    if err != nil {
      return nil, err
    }
  } else if answer, ok := b["_answer"]; ok {
    err = writeKV(buf, "_answer", answer)
    if err != nil {
      return nil, err
    }
  }
  if cmd, ok := b["_command"]; ok {
    err = writeKV(buf, "_command", cmd)
    if err != nil {
      return nil, err
    }
  }
  for k, v := range b {
    if k == "_ask" || k == "_command" || k == "_answer" {
      continue
    }
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
    err = writeKV(buf, k, v)
    if err != nil {
      return nil, err
    }
  }
  buf.WriteString("\x00\x00")
  return buf.Bytes(), nil
}

// Deserialize converts a wire representation to an AMP box, if possible.
func Deserialize(b []byte) (Box, error) {
  if len(b) < 2 {
    return nil, fmt.Errorf("serialized AMP box length was %v, shorter than two bytes, which is illegal", len(b))
  }
  return Decode(bytes.NewBuffer(b))
}

func Decode(buf io.Reader) (Box, error) {
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

func (box Box) Unmarshal(x interface{}) error {
  if x == nil {
    return nil
  }
  rv := reflect.ValueOf(x).Elem()
  rt := rv.Type()
  l := rv.NumField()
  for i := 0; i < l; i++ {
    f := rv.Field(i)
    if !f.IsValid() {
      continue
    }
    if !f.CanSet() {
      continue
    }
    nameTag := rt.Field(i).Tag.Get("amp")
    if nameTag == "" {
      nameTag = rt.Field(i).Name
    }
    if _, ok := box[nameTag]; !ok {
      continue
    }
    switch f.Type().Name() {
    case "string":
      f.SetString(box[nameTag])
    case "int":
      i, err := strconv.Atoi(box[nameTag])
      if err != nil {
        return err
      }
      f.SetInt(int64(i))
    case "float32", "float64":
      var width int
      if f.Type().Name() == "float32" {
        width = 32
      } else {
        width = 64
      }
      fl, err := strconv.ParseFloat(box[nameTag], width)
      if err != nil {
        return err
      }
      f.SetFloat(fl)
    case "bool":
      if box[nameTag] == "True" {
        f.SetBool(true)
      } else if box[nameTag] == "False" {
        f.SetBool(false)
      } else {
        return fmt.Errorf("%q is not a valid boolean.", box[nameTag])
      }
    }
  }
  return nil
}
