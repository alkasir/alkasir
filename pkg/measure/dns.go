package measure

import (
	"net"
	"strings"
	"time"

	"github.com/alkasir/alkasir/pkg/measure/sampletypes"

	"encoding/json"

	"github.com/miekg/dns"
)

type DNSQuery struct {
	Hostname string `json:"hostname"` // Hostnames to test against
	Resolver string `json:"resolver"` // Resolver to use, an empty string means to use the default system resolver.
}

func (d DNSQueryResult) Type() sampletypes.SampleType {
	return sampletypes.DNSQuery
}

func (d DNSQueryResult) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

func (d DNSQueryResult) Host() string {
	return d.Hostname
}

// DNSQueryResult .
type DNSQueryResult struct {
	// Resolver net.IP
	Addrs    []string `json:"addrs"`
	Error    string   `json:"error"`
	Hostname string   `json:"hostname"`
	Resolver string   `json:"resolver"`
}

func lookup(lookupaddr, nameserveraddr string) ([]string, error) {

	// if no nameserver addr is supplied, use the built in resolver config
	if nameserveraddr == "" {
		return net.LookupHost(lookupaddr)
	}
	var results []string
	if !strings.HasSuffix(lookupaddr, ".") {
		lookupaddr += "."
	}
	m := new(dns.Msg)
	m.SetQuestion(lookupaddr, dns.TypeA)
	ret, err := dns.Exchange(m, nameserveraddr)
	if err != nil {
		return nil, err
	}
	for _, a := range ret.Answer {
		if t, ok := a.(*dns.A); ok {
			results = append(results, t.A.String())
		}
	}
	return results, nil
}

func (d DNSQuery) Measure() (Measurement, error) {
	qr := DNSQueryResult{
		Hostname: d.Hostname,
		Addrs:    []string{},
		Resolver: d.Resolver,
	}
	lookupresult := make(chan []string, 0)
	lookuperror := make(chan error, 0)
	go func(d DNSQuery) {
		addrs, err := lookup(d.Hostname, d.Resolver)
		if err != nil {
			lookuperror <- err
			return
		}
		lookupresult <- addrs
	}(d)
	select {
	case lr := <-lookupresult:
		qr.Addrs = lr
	case err := <-lookuperror:
		qr.Error = err.Error()
	case <-time.After(defaultTimeout):
		qr.Error = "timeout: " + defaultTimeout.String()
	}

	close(lookupresult)
	close(lookuperror)

	return qr, nil

}
