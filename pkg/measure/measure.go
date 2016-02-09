package measure

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/alkasir/alkasir/pkg/measure/sampletypes"
	"github.com/alkasir/alkasir/pkg/shared"
)

type Measurement interface {
	Marshal() ([]byte, error)
	Type() sampletypes.SampleType
	Host() string
}

type Measurer interface {
	Measure() (Measurement, error)
}

type ErrorHostNotAllowed struct {
	Hostname string
}

func (e ErrorHostNotAllowed) Error() string {
	return fmt.Sprintf("host %s not allowed", e.Hostname)
}

func DefaultMeasurements(URL string) ([]Measurer, error) {
	URL = strings.TrimSpace(URL)
	var result []Measurer
	if URL == "" {
		return nil, errors.New("emtpy url")
	}
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	if !shared.AcceptedURL(u) {
		return nil, ErrorHostNotAllowed{URL}
	}

	host := u.Host
	if strings.Contains(host, ":") {
		host, _, err = net.SplitHostPort(u.Host)
	}

	for _, resolver := range []string{"", "8.8.8.8:53"} {
		dnsm := DNSQuery{
			Hostname: host,
			Resolver: resolver,
		}
		result = append(result, dnsm)
	}

	httphm := HTTPHeader{
		URL: URL,
	}
	result = append(result, httphm)

	return result, nil
}

var defaultTimeout = time.Duration(45 * time.Second)
