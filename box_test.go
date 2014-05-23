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
  box := Box{"_ask": "1", "_command": "listpeer", "21": "notfirst"}
  got, err := box.Serialize()
  if err != nil {
    t.Fatal(err)
  }
  if string(got[2:6]) != "_ask" {
    t.Fatal("_ask key is not the first key in the serialization.")
  }
  if string(got[11:19]) != "_command" {
    t.Fatal("_command key is not the second key in the serialization.")
  }

  delete(box, "_ask")
  box["_answer"] = "2"
  got, err = box.Serialize()
  if err != nil {
    t.Fatal(err)
  }
  if string(got[2:9]) != "_answer" {
    t.Fatal("_answer key is not the first key in the serialization.")
  }
}

type unmarshalTest struct {
  SomeValue   string `amp:"some_value"`
  SomeBool    bool   `amp:"some_bool"`
  SomeInt     int
  SomeFloat   float64
  AnotherBool bool `amp:"another_bool"`
}

func TestUnmarshal(t *testing.T) {
  x := &unmarshalTest{}
  box := Box{"some_value": "a value", "some_bool": "True", "SomeInt": "7", "SomeFloat": "8.2", "another_bool": "False"}
  err := box.Unmarshal(x)
  if err != nil {
    t.Fatal(err)
  }
  if x.SomeValue != "a value" {
    t.Errorf("x.SomeValue = %q, expected \"a value\"", x.SomeValue)
  }
  if !x.SomeBool {
    t.Errorf("x.SomeBool = %v, expected true", x.SomeBool)
  }
  if x.AnotherBool {
    t.Errorf("x.AnotherBool = %v, expected false", x.AnotherBool)
  }
  if x.SomeInt != 7 {
    t.Errorf("x.SomeInt = %v, expected 7", x.SomeInt)
  }
  if x.SomeFloat != 8.2 {
    t.Errorf("x.SomeFloat = %v, expected 8.2", x.SomeFloat)
  }
}
