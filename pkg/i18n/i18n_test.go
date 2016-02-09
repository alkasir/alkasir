package i18n

import (
	"io/ioutil"
	"testing"
)

func TestAddBundle(t *testing.T) {
	data, err := ioutil.ReadFile("../../res/messages/en/messages.json")
	if err != nil {
		t.Fail()
	}
	AddBundle("en-us", data)

	T, err := Tfunc("en-us")
	if err != nil {
		t.Fail()
	}
	v := T("action_ok")
	if v != "Ok" {
		t.Fail()
	}
}
