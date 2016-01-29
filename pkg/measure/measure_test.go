package measure

import (
	"reflect"
	"testing"
)

func TestMarshalHTTPHeader(t *testing.T) {
	h := HTTPHeaderResult{
		URL: "http://testtest",
		ResponseHeader: map[string]string{
			"status": "200",
		},
	}
	var b1, b2 []byte
	{
		var err error
		b1, err = h.Marshal()
		if err != nil {
			t.Error(err)
		}

	}
	{
		var err error
		var m Measurement
		m = h
		b2, err = m.Marshal()
		if err != nil {
			t.Error(err)
		}

	}
	if !reflect.DeepEqual(b1, b2) {
		t.Fail()
	}
}

func TestMarshalDNSQuery(t *testing.T) {
	var d DNSQueryResult = DNSQueryResult{
		Addrs:    []string{"1.1.1.1"},
		Hostname: "test",
		Resolver: "",
	}
	var b1, b2 []byte
	{
		var err error
		b1, err = d.Marshal()
		if err != nil {
			t.Error(err)
		}

	}
	{
		var err error
		var m Measurement
		m = d
		b2, err = m.Marshal()
		if err != nil {
			t.Error(err)
		}

	}
	if !reflect.DeepEqual(b1, b2) {
		t.Fail()
	}
}
