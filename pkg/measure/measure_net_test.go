// +build net

package measure

import (
	"fmt"
	"log"
	"testing"
)

func TestDNSMeasurer(t *testing.T) {
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

	ht := HTTPHeader{
		URL: "https://google.com",
	}
	r, err := ht.Measure()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v", r)

}
