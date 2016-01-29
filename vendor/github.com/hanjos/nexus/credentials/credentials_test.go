package credentials_test

import (
	"github.com/hanjos/nexus/credentials"
	"net/http"
	"testing"
)

func TestNoneImplementsCredentials(t *testing.T) {
	if _, ok := interface{}(credentials.None).(credentials.Credentials); !ok {
		t.Errorf("credentials.None doesn't implement credentials.Credentials!")
	}
}

func TestOrZeroReturnsTheGivenNonNilArgument(t *testing.T) {
	c := credentials.BasicAuth("", "")
	if v := credentials.OrZero(c); v != c {
		t.Errorf("credentials.OrZero(%v) should've returned %v, not %v!", c, c, v)
	}
}

func TestOrZeroReturnsNoneOnNil(t *testing.T) {
	if v := credentials.OrZero(nil); v != credentials.None {
		t.Errorf("credentials.OrZero(nil) should've returned credentials.None, not %v!", v)
	}
}

func TestNoneSignDoesntBarfOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%v", r)
		}
	}()

	credentials.None.Sign(nil)
}

func TestBasicAuthSignDoesntBarfOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%v", r)
		}
	}()

	credentials.BasicAuth("u", "p").Sign(nil)
}

func TestBasicAuthAddAuthorizationDataToTheRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "http://www.google.com", nil)
	if err != nil {
		t.Errorf("Error creating request: %v", err)
		return
	}

	credentials.BasicAuth("username", "password").Sign(req)

	if req.Header.Get("Authorization") != "Basic dXNlcm5hbWU6cGFzc3dvcmQ=" {
		t.Errorf("Expected Basic dXNlcm5hbWU6cGFzc3dvcmQ=, got %v", req.Header.Get("Authorization"))
	}
}

func TestNoneRemovesAuthorizationDataFromTheRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "http://www.google.com", nil)
	if err != nil {
		t.Errorf("Error creating request: %v", err)
		return
	}

	req.SetBasicAuth("username", "password")

	credentials.None.Sign(req)

	if req.Header.Get("Authorization") != "" {
		t.Errorf("Expected \"\", got %v", req.Header.Get("Authorization"))
	}
}
