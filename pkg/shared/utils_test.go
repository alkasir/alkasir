package shared

import (
	"log"
	"testing"
)

func TestIdGenSimple(t *testing.T) {
	ig, err := NewIDGen("HEPPO")
	if err != nil {
		log.Printf("err: %+v", err)
		t.Fail()
	}
	if ig == nil {
		log.Printf("no return value")
		t.Fail()
	}
	v1 := ig.New()
	v2 := ig.New()
	if v1 != "HEPPO-1" {
		t.Fail()
	}

	if v2 != "HEPPO-2" {
		t.Fail()
	}
}

func TestIdGenBadPrefix(t *testing.T) {
	ig, err := NewIDGen("h8@(!UJF)")
	if err == nil || ig != nil {
		log.Printf("Invalid name should fail")
		t.Fail()
	}

}

func TestSafeClean(t *testing.T) {
	s := scrubbed
	testData := map[string]string{
		"helloo 192.168.0.1 sddsd":           "helloo " + s + " sddsd",
		"helloo 12:12 sddsd":                 "helloo 12:12 sddsd",
		"helloo 192.168.0.1 sd192.168.0.1sd": "helloo " + s + " sd" + s + "sd",
	}
	for k, v := range testData {
		if SafeClean(k) != v {
			t.Errorf("''%s' is not equal to '%s'", k, v)
		}
	}
}

func TestClean(t *testing.T) {
	s := scrubbed
	testData := [][3]string{
		{"original secret string", "secret", "original " + s + " string"},
		{"original secrett string", "secret", "original " + s + "t string"},
		{"one two one two three five one", "one", s + " two " + s + " two three five " + s},
		{"original public string", "secret", "original public string"},
	}
	for _, v := range testData {
		if SafeCleanReplace(v[0], v[1]) != v[2] {
			t.Errorf("''%s' is not equal to '%s'", v[0], v[2])
		}
	}

}
