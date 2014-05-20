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
		t.Errorf("box.Serialize() returned error: %v", err)
		t.FailNow()
	}
	t.Logf("%v", string(bytes))
	got, err := Deserialize(bytes)
	if err != nil {
		t.Errorf("Deserialize() returned error: %v", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(got, box) {
		t.Errorf("Got %+v, but was expecting %+v.", got, box)
	}
}

const pythonGeneratedWire = "\x00\x04_ask\x00\x011\x00\x08_command\x00\x08listpeer\x00\x00"

var pythonGeneratedBox = Box{"_ask": "1", "_command": "listpeer"}

func TestDeserializeFromPython(t *testing.T) {
	got, err := Deserialize([]byte(pythonGeneratedWire))
	if err != nil {
		t.Errorf("Deserialize() returned error: %v", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(got, pythonGeneratedBox) {
		t.Errorf("Got %+v, but was expecting %+v.", got, pythonGeneratedBox)
	}
}
