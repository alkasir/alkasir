package client

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

import "reflect"

func reset() {
	resetSuggestions <- true
}

func TestCreateSuggestion(t *testing.T) {
	reset()
	// add a local suggestion
	s := NewSuggestion("http://google.com")
	r, ok := GetSuggestion(s.ID)
	if !ok {
		t.Error("could not get suggestion")
	}
	// should be the same value
	if !reflect.DeepEqual(r, s) {
		t.Error("r != s")
	}

	// add another one
	s2 := NewSuggestion("http://google.com")
	r2, ok := GetSuggestion(s2.ID)
	if !ok {
		t.Error("could not get suggestion")
	}
	// should be the same value
	if !reflect.DeepEqual(r2, s2) {
		t.Error("r2 != s2")
	}

	// should not be the same value
	if reflect.DeepEqual(r, r2) {
		t.Error("r == s")
	}

	// count suggestions
	suggestions := AllSuggestions()
	if len(suggestions) != 2 {
		t.Errorf("expected 2 suggestions, found %d", len(suggestions))
	}
}

func TestRequestTokenSuccess(t *testing.T) {
	reset()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/suggestions/new/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{
  "Token": "c1c2813416e6da508f4d5b3998956ace6ab97f726915fead6430b724c8b57e28",
  "URL": "http:\/\/data",
  "Ok": true }`)
		}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := NewClient(ts.URL, nil)
	suggestion := NewSuggestion("http://data")
	tokenResponse, err := suggestion.RequestToken(client, net.IPv4(130, 234, 12, 2), "SE")
	if err != nil {
		t.Error(err)
	}

	if tokenResponse.Token != "c1c2813416e6da508f4d5b3998956ace6ab97f726915fead6430b724c8b57e28" {
		t.Error("got wrong token back")
	}

	suggestionCheck, ok := GetSuggestion(suggestion.ID)
	if !ok {
		t.Error("could not get suggestion")
	}

	if tokenResponse.Token != suggestionCheck.Token {
		t.Errorf("token mismatch [%s] [%s]", suggestion.Token, suggestionCheck.Token)
	}

}

func TestRequestTokenFail(t *testing.T) {
	reset()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/suggestions/new/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{
  "Token": "THISSHOULDNOTBEADDED",
  "URL": "http:\/\/data",
  "Ok": false }`)
		}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := NewClient(ts.URL, nil)

	suggestion := NewSuggestion("http://data")
	tokenResponse, err := suggestion.RequestToken(client, net.IPv4(130, 234, 12, 2), "SE")
	if err != nil {
		t.Error(err)
	}

	suggestionCheck, ok := GetSuggestion(suggestion.ID)
	if !ok {
		t.Error("could not get suggestion")
	}

	if tokenResponse.Token == suggestionCheck.Token {
		t.Errorf("token mismatch [%s] [%s]", tokenResponse.Token, suggestionCheck.Token)
	}

	if suggestion.Token == tokenResponse.Token {
		t.Errorf("token mismatch [%s] [%s]", suggestion.Token, tokenResponse.Token)
	}

	if tokenResponse.Ok {
		t.Fail()
	}

	// count suggestions
	suggestions := AllSuggestions()
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestions, found %d", len(suggestions))
	}
}

func TestSubmitSample(t *testing.T) {
	reset()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/suggestions/new/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{
  "Token": "TestSubmitSamtpleSuccessToken",
  "URL": "http:\/\/TestSubmitSample",
  "Ok": true }`)
		}))

	var storeSampleResponse string
	mux.HandleFunc("/v1/samples/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if storeSampleResponse == "" {
				fmt.Fprintln(w, `{"Ok": true}`)
			} else {
				fmt.Fprintln(w, storeSampleResponse)
			}
		}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := NewClient(ts.URL, nil)

	// create local suggestion
	s := NewSuggestion("http://TestSubmitSample")
	// add sample right away
	s.AddSample("BrowserExtensionSample", "{}")

	// create session token
	{
		tokenResponse, err := s.RequestToken(client, net.IPv4(133, 23, 123, 21), "US")
		if err != nil || !tokenResponse.Ok {
			t.Fail()
		}
	}

	// send the previously created sample
	{
		n, err := s.SendSamples(client)
		if err != nil {
			t.Error(err)
		}
		if n != 1 {
			t.Errorf("expected 1 sent, got %d", n)
		}
	}
	// send two more samples
	{
		s.AddSample("BrowserExtensionSample", "{}")
		s.AddSample("BrowserExtensionSample", "{}")
		n, err := s.SendSamples(client)
		if err != nil {
			t.Error(err)
		}
		if n != 2 {
			t.Errorf("expected 2 sent, got %d", n)
		}

		n, err = s.SendSamples(client)
		if err != nil {
			t.Error(err)
		}
		if n != 0 {
			t.Errorf("expected 0 sent, got %d", n)
		}
	}

	for _, errresp := range []string{`{"Ok": false}`, `{"Ok": false, "Error": "errnsg"}`} {
		// test failed store sample request
		{
			s, _ = GetSuggestion(s.ID)
			storeSampleResponse = errresp

			if len(s.samples) != 0 {
				t.Error("ss")
			}
			s.AddSample("BrowserExtensionSample", "{}")
			s, _ = GetSuggestion(s.ID)
			if len(s.samples) != 1 {
				t.Error("wrong amount of samples")
			}
			n, err := s.SendSamples(client)
			if err == nil {
				t.Error("was expecting error")
			}
			if n != 0 {
				t.Errorf("expected 0 sent, got %d", n)
			}
			s, _ = GetSuggestion(s.ID)
			if len(s.samples) != 1 {
				t.Errorf("expected len(s.samples) == 1, got %d", n)
			}

		}
		// resent the failed samples (problems magically went away)
		{
			storeSampleResponse = ""
			n, err := s.SendSamples(client)
			if err != nil {
				t.Error(err)
			}
			if n != 1 {
				t.Errorf("expected 1 sent, got %d", n)
			}
		}
	}

}
