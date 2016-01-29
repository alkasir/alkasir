package measure

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thomasf/lg"

	"github.com/alkasir/alkasir/pkg/measure/sampletypes"
)

// HTTPHeader .
type HTTPHeader struct {
	URL string `json:"url"`
}

// HTTPHeaderResult .
type HTTPHeaderResult struct {
	URL            string            `json:"url"`
	ResponseHeader map[string]string `json:"response_header"`
	Error          string            `json:"error"`
}

func (h HTTPHeaderResult) Type() sampletypes.SampleType {
	return sampletypes.HTTPHeader
}

func (h HTTPHeaderResult) Host() string {
	u, err := url.Parse(h.URL)
	if err != nil {
		lg.Fatalf("invalid url: %s %s", h.URL, err.Error())
	}

	host := u.Host
	if strings.Contains(host, ":") {
		host, _, err = net.SplitHostPort(u.Host)
		if err != nil {
			lg.Fatalf("invalid url: %s %s", h.URL, err.Error())
		}
	}
	return host
}

func (h HTTPHeader) Measure() (Measurement, error) {
	resultchan := make(chan HTTPHeaderResult, 0)
	go func() {
		client := http.Client{
			Timeout: defaultTimeout + 5*time.Second,
		}
		resp, err := client.Get(h.URL)
		if err != nil {
			resultchan <- HTTPHeaderResult{
				URL:   h.URL,
				Error: err.Error(),
			}
			return
		}
		defer resp.Body.Close()
		respHeaders := make(map[string]string, 0)
		for k := range resp.Header {
			respHeaders[k] = resp.Header.Get(k)
		}
		resp.Body.Close()
		resultchan <- HTTPHeaderResult{
			URL:            h.URL,
			ResponseHeader: respHeaders,
		}
	}()
	select {
	case res := <-resultchan:
		return res, nil
	case <-time.After(defaultTimeout):
		return HTTPHeaderResult{
			URL:   h.URL,
			Error: "timeout: " + defaultTimeout.String(),
		}, nil
	}

}

func (h HTTPHeaderResult) Marshal() ([]byte, error) {
	return json.Marshal(h)
}
