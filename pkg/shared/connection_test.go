package shared

import (
	"reflect"
	"testing"
)

func TestConnectionEncodingAndDecoding(t *testing.T) {
	conn1 := Connection{
		Secret:    "A very secret key",
		Addr:      "server.domain",
		Transport: "TestTransport",
	}

	conn2 := Connection{
		Secret:    "Yet another secret key",
		Addr:      "server.domain",
		Transport: "TestTransport",
	}
	enc1, _ := conn1.Encode()
	enc2, _ := conn2.Encode()

	dec1, err := DecodeConnection(enc1)
	if err != nil {
		t.Error(err)
	}

	dec2, err := DecodeConnection(enc2)
	if err != nil {
		t.Error(err)
	}

	dec1.ID = ""
	dec2.ID = ""

	if !reflect.DeepEqual(conn1, dec1) {
		t.Error("input and output not equal")
	}
	if !reflect.DeepEqual(conn2, dec2) {
		t.Error("input and output not equal")
	}
	if reflect.DeepEqual(dec1, dec2) {
		t.Error("dec1 and dec2 should not be equal")
	}
}
