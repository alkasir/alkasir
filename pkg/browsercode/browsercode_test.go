package browsercode

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBrowserCode(t *testing.T) {
	bc, err := New("localhost:5599")
	if err != nil {
		t.Error(err)
	}
	encoded, err := bc.Encode()
	if err != nil {
		t.Error(err)
	}
	expected := fmt.Sprintf("%s::5599", bc.Key)
	if encoded != expected {
		t.Errorf("not eq! %s %s", encoded, expected)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(bc, decoded) {
		t.Errorf("not eq! %s %s", bc, decoded)
	}

}

func TestBrowserCode2(t *testing.T) {
	bc, err := New(":5599")
	if err != nil {
		t.Error(err)
	}
	encoded, err := bc.Encode()
	if err != nil {
		t.Error(err)
	}
	expected := fmt.Sprintf("%s::5599", bc.Key)
	if encoded != expected {
		t.Errorf("not eq! %s %s", encoded, expected)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(bc, decoded) {
		t.Errorf("not eq! %s %s", bc, decoded)
	}

}

func TestBrowserCode4(t *testing.T) {
	bc, err := New("127.0.0.1:5599")
	if err != nil {
		t.Error(err)
	}
	encoded, err := bc.Encode()
	if err != nil {
		t.Error(err)
	}
	expected := fmt.Sprintf("%s::5599", bc.Key)
	if encoded != expected {
		t.Errorf("not eq! %s %s", encoded, expected)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Error(err)
	}

	if reflect.DeepEqual(bc, decoded) {
		t.Errorf("should not be equal: %s %s", bc, decoded)
	}

}

func TestBrowserCode3(t *testing.T) {
	bc, err := New("localhost:")
	if err == nil {
		t.Error("port is required")
	}

	_, err = bc.Encode()
	if err == nil {
		t.Error("encoding empty struct should fail")
	}

}
