package measure

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDNSMeasurer(t *testing.T) {
	t.Parallel()
	dq := DNSQuery{
		Hostname: "google.com",
		Resolver: "8.8.8.8:53",
	}
	res, err := dq.Measure()
	if err != nil {
		t.Fatal(err)

	}
	log.Println(res)

}

func TestDNSMeasurer2(t *testing.T) {
	t.Parallel()
	dq := DNSQuery{
		Hostname: "google.com",
		Resolver: "localhost:39272",
	}
	res, err := dq.Measure()
	if err != nil {
		t.Fatal(err)

	}
	log.Println(res)

}

func TestHTTPHeader(t *testing.T) {
	t.Parallel()
	ht := HTTPHeader{
		URL: "https://google.com",
	}
	r, err := ht.Measure()
	if err != nil {
		t.Fatal(err)
	}
	_ = r
	// fmt.Printf("%+v", r)

}

func TestHTTPHeaderTimeout(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer ts.Close()
	ht := HTTPHeader{
		URL:     ts.URL,
		Timeout: 10 * time.Millisecond,
	}
	r, err := ht.Measure()
	if err != nil {
		t.Fatal(err)
	}
	switch d := r.(type) {
	case HTTPHeaderResult:
		if !strings.HasPrefix(d.Error, "timeout:") {
			t.Fatalf("expected timeout error, got %s", d.Error)
		}
	default:
		t.Fatal("expected httpheaderresult")
	}
}
