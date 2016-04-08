package measure

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alkasir/alkasir/pkg/measure/sampletypes"
	"github.com/thomasf/lg"
)

// HTTPHeader .
type HTTPHeader struct {
	URL     string        `json:"url"` // The url to mesaure against
	Timeout time.Duration `json:"-"`   // Measurement timeout, defaults to 45 seconds unless specified.
}

// HTTPHeaderResult .
type HTTPHeaderResult struct {
	URL            string            `json:"url"`
	ResponseHeader map[string]string `json:"response_header"`
	Redirects      []Redirect        `json:"redirects,omitempty"`
	Error          string            `json:"error"`
	StatusCode     int               `json:"status_code"`
}

// Redirect is recorded on HTTP redirects.
type Redirect struct {
	StatusCode int               `json:"status_code"`
	Header     map[string]string `json:"header"`
	URL        string            `json:"url"`
}

func (h HTTPHeader) Measure() (Measurement, error) {
	timeout := h.Timeout
	if timeout == 0 {
		timeout = 45 * time.Second
	}

	resultchan := make(chan HTTPHeaderResult, 0)
	go func() {
		rr := &redirectRecorder{
			Transport: http.DefaultTransport.(*http.Transport),
		}
		client := http.Client{
			Timeout:   timeout,
			Transport: rr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
				return nil
			},
		}
		resp, err := client.Get(h.URL)
		if err != nil {
			resultchan <- HTTPHeaderResult{
				URL:   h.URL,
				Error: err.Error(),
			}
			return
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				// resp.Body should always have been closed at this point so an
				// error is expected.
				lg.V(20).Infoln(err)
			}
		}()
		respHeaders := make(map[string]string, 0)
		for k := range resp.Header {
			respHeaders[k] = resp.Header.Get(k)
		}
		if err := resp.Body.Close(); err != nil {
			lg.Errorln(err)
		}
		resultchan <- HTTPHeaderResult{
			URL:            h.URL,
			ResponseHeader: respHeaders,
			StatusCode:     resp.StatusCode,
			Redirects:      rr.Redirects,
		}
	}()
	select {
	case res := <-resultchan:
		return res, nil
	case <-time.After(timeout):
		return HTTPHeaderResult{
			URL:   h.URL,
			Error: "timeout: " + timeout.String(),
		}, nil
	}

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

func (h HTTPHeaderResult) Marshal() ([]byte, error) {
	return json.Marshal(h)
}

type redirectRecorder struct {
	*http.Transport
	Redirects []Redirect
}

func (t *redirectRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect:
		header := make(map[string]string, 0)
		for k := range resp.Header {
			header[k] = resp.Header.Get(k)
		}
		t.Redirects = append(
			t.Redirects, Redirect{
				StatusCode: resp.StatusCode,
				URL:        req.URL.String(),
				Header:     header,
			},
		)
	}
	return resp, err
}
