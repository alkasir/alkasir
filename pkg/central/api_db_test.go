// +build databases

package central

import (
	"log"
	"net"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alkasir/alkasir/pkg/central/client"
	"github.com/alkasir/alkasir/pkg/central/db"
)

var testclient *client.Client

func TestMain(m *testing.M) {
	err := Init()
	if err != nil {
		panic(err)
	}
	clients := db.Clients{
		DB:       sqlDB,
		Internet: db.NewInternetClient(redisPool),
		Maxmind:  db.NewMaxmindClient(mmCountryDB, mmCityDB),
	}
	mux, err := apiMux(clients)
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(mux)
	testclient = client.NewClient(ts.URL, nil)
	defer ts.Close()
	os.Exit(m.Run())
}

func TestRequestToken(t *testing.T) {
	suggestion := client.NewSuggestion("http://google.com")
	resp, err := suggestion.RequestToken(testclient, net.IPv4(85, 225, 60, 122), "SE")
	if err != nil {
		t.Error(err)
	}
	if !resp.Ok {
		t.Error(resp)
	}
}

func TestSendSample(t *testing.T) {
	suggestion := client.NewSuggestion("http://google2.com")
	resp, err := suggestion.RequestToken(testclient, net.IPv4(85, 225, 60, 122), "SE")
	if err != nil {
		t.Error(err)
	}
	if !resp.Ok {
		t.Error(resp)
	}
	log.Printf("resp: %+v", resp)
	suggestion.AddSample("DNSQuery", "{}")
	n, err := suggestion.SendSamples(testclient)
	if err != nil {
		t.Error(err)
	}

	if n != 1 {
		t.Errorf("expected n==1, n=%d", n)
	}

}
