package db

import "testing"

func TestAPICredentials(t *testing.T) {
	c := APICredentials{
		Username: "nameless",
		Enabled:  true,
	}

	ok, err := c.IsValid("error")
	if ok || err == nil {
		t.Error("error should error")
	}

	err = c.SetPassword("newpassword")
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	ok, err = c.IsValid("wrongpassword")
	if ok {
		t.Error("wrongpassword should not be ok")
	}
	if err != nil {
		t.Error("err should be nil")
	}

	c.Enabled = false
	ok, err = c.IsValid("newpassword")
	if ok {
		t.Error("newpassword should not be ok (not enabled)")
	}

}
