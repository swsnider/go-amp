package amp

import (
  "reflect"
  "testing"
)

func TestSerializationRoundTrip(t *testing.T) {
  box := Box{
    "_ask":     "1",
    "_command": "listpeer",
    "payload":  "{\"key\": \"value\"}",
  }
  bytes, err := box.Serialize()
  if err != nil {
    t.Fatalf("box.Serialize() returned error: %v", err)
  }
  t.Logf("%v", string(bytes))
  got, err := Deserialize(bytes)
  if err != nil {
    t.Fatalf("Deserialize() returned error: %v", err)
  }
  if !reflect.DeepEqual(got, box) {
    t.Fatalf("Got %+v, but was expecting %+v.", got, box)
  }
}

const pythonGeneratedWire = "\x00\x04_ask\x00\x011\x00\x08_command\x00\x08listpeer\x00\x00"

var pythonGeneratedBox = Box{"_ask": "1", "_command": "listpeer"}

func TestDeserializeFromPython(t *testing.T) {
  got, err := Deserialize([]byte(pythonGeneratedWire))
  if err != nil {
    t.Fatalf("Deserialize() returned error: %v", err)
  }
  if !reflect.DeepEqual(got, pythonGeneratedBox) {
    t.Fatalf("Got %+v, but was expecting %+v.", got, pythonGeneratedBox)
  }
}

func TestSerializeOrder(t *testing.T) {
  got, err := pythonGeneratedBox.Serialize()
  if err != nil {
    t.Fatal(err)
  }
  if string(got[2:6]) != "_ask" {
    t.Fatal("_ask key is not the first key in the serialization.")
  }
  if string(got[11:19]) != "_command" {
    t.Fatal("_command key is not the second key in the serialization.")
  }
}
