package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alkasir/alkasir/pkg/measure"
	"github.com/alkasir/alkasir/pkg/measure/sampletypes"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/alkasir/alkasir/pkg/shared/apierrors"
)

func NewClient(baseurl string, httpclient *http.Client) *Client {
	if httpclient == nil {
		httpclient = http.DefaultClient
	}
	return &Client{
		httpcli: httpclient,
		baseurl: strings.TrimRight(baseurl, "/"),
	}
}

// Client .
type Client struct {
	baseurl string
	httpcli *http.Client
}

// CreateSuggestionToken requests an new suggestion token from central.
func (c *Client) CreateSuggestionToken(request shared.SuggestionTokenRequest) (shared.SuggestionTokenResponse, error) {
	data, err := json.Marshal(&request)
	var response shared.SuggestionTokenResponse
	if err != nil {
		return response, err
	}
	resp, err := c.post("suggestions/new/", bytes.NewBuffer(data))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response,
			fmt.Errorf("createSuggestionToken status: %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}
	if response.Error != "" {
		response.Ok = false
	}
	return response, nil
}

// CreateSample posts an sample to central.
func (c *Client) CreateSample(request shared.StoreSampleRequest) (shared.SampleResponse, error) {
	data, err := json.Marshal(&request)
	if err != nil {
		return shared.SampleResponse{}, err
	}
	resp, err := c.post("samples/", bytes.NewBuffer(data))
	if err != nil {
		return shared.SampleResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return shared.SampleResponse{},
			fmt.Errorf("CreateSample http status response: %d", resp.StatusCode)
	}
	var response shared.SampleResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return shared.SampleResponse{}, err
	}
	return response, nil
}

// UpdateHostlist posts an sample to central.
func (c *Client) UpdateHostlist(request shared.UpdateHostlistRequest) (shared.UpdateHostlistResponse, error) {
	data, err := json.Marshal(&request)
	if err != nil {
		return shared.UpdateHostlistResponse{}, err
	}
	resp, err := c.post("hosts/", bytes.NewBuffer(data))
	if err != nil {
		return shared.UpdateHostlistResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return shared.UpdateHostlistResponse{},
			fmt.Errorf("Updatehostlist http status response: %d", resp.StatusCode)
	}
	var response shared.UpdateHostlistResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return shared.UpdateHostlistResponse{}, err
	}

	return response, nil
}

// CheckBinaryUpgrade returns upgrade meta data if an upgrade is available.
func (c *Client) CheckBinaryUpgrade(request shared.BinaryUpgradeRequest) (shared.BinaryUpgradeResponse, bool, error) {
	data, err := json.Marshal(&request)
	var response shared.BinaryUpgradeResponse
	if err != nil {
		return response, false, err
	}
	resp, err := c.post("upgrades/", bytes.NewBuffer(data))
	if err != nil {
		return response, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return response, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return response, false,
			fmt.Errorf("Upgradeversion http status response: %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, false, err
	}

	return response, true, nil
}

func (c *Client) url(resource string) string {
	return fmt.Sprintf("%s/v1/%s", c.baseurl, resource)
}

func (c *Client) get(resource string) (resp *http.Response, err error) {
	return c.httpcli.Get(c.url(resource))
}

func (c *Client) post(resource string, body io.Reader) (resp *http.Response, err error) {
	return c.httpcli.Post(c.url(resource), "application/json", body)
}

// sample .
type sample struct {
	id         string
	sampleType string    // sample type
	data       string    // json encoded sample data (should be []byte)
	createdAt  time.Time // Timestamp.
}

// Suggestion keeps local and remote suggestions together.
type Suggestion struct {
	ID        string                 // ID is assigned by the client on creation.
	Token     shared.SuggestionToken // SuggestionToken session key from central.
	URL       string                 // The origin url for the suggestion session.
	CreatedAt time.Time              // Timestamp.
	samples   []sample               // For keeping local cache of samples before submitting to central.
}

// RequestToken requests a suggestion session from central. On success the
// Suggestion is updated the token id. Check SuggestionTokenResponse.Ok for
// validity.
func (s *Suggestion) RequestToken(client *Client, clientAddr net.IP, countryCode string) (shared.SuggestionTokenResponse, error) {
	r, err := client.CreateSuggestionToken(shared.SuggestionTokenRequest{
		URL:         s.URL,
		ClientAddr:  clientAddr,
		CountryCode: countryCode,
	})
	if err != nil {
		return r, err
	}
	if r.Ok {
		modifySuggestion <- setToken(*s, r.Token)
		s.Token = r.Token
	}
	return r, err
}

var supportedTypes = map[sampletypes.SampleType]bool{
	sampletypes.DNSQuery:   true,
	sampletypes.HTTPHeader: true,
}

func (s *Suggestion) AddMeasurement(m measure.Measurement) error {
	b, err := m.Marshal()
	if err != nil {
		return err
	}
	t := m.Type()
	if _, ok := supportedTypes[t]; !ok {
		return fmt.Errorf("unsupported sampletype: %s", t.String())
	}
	s.AddSample(m.Type().String(), string(b))
	return nil
}

// AddSample adds a sample to the local cache ready to be sent to central using SendSamples.
func (s *Suggestion) AddSample(sampleType, data string) {
	ID, err := shared.SecureRandomString(32)
	if err != nil {
		panic("could not generate random number")
	}
	ss := sample{
		id:         ID,
		sampleType: sampleType,
		data:       data,
		createdAt:  time.Now(),
	}
	modifySuggestion <- addSample(*s, ss)
}

// AddSample adds a sample to the local cache ready to be sent to central using SendSamples.
func (sugg *Suggestion) SendSamples(client *Client) (int, error) {
	s, ok := GetSuggestion(sugg.ID)
	if !ok {
		return 0, apierrors.NewNotFound("suggestion", sugg.ID)
	}
	if s.Token == "" {
		return 0, apierrors.NewConflict("suggestion", "", errors.New("No session token associated with suggestion"))
	}

	n := 0
	var rerr error

	for _, v := range s.samples {
		r, err := client.CreateSample(shared.StoreSampleRequest{
			Sample: &shared.Sample{
				Token:      s.Token,
				SampleType: v.sampleType,
				Data:       v.data,
				URL:        s.URL,
			},
			ClientAddr: shared.GetPublicIPAddr().String(),
		})
		if err != nil {
			rerr = errors.New("error sending sample: " + err.Error())
			continue
		}

		if !r.Ok {
			rerr = errors.New("error sending sample, not accepted")
			continue

		}
		n++
		modifySuggestion <- removeSample(s, v)
	}

	return n, rerr
}

// NewSuggestion creates a new local suggestion
func NewSuggestion(URL string) Suggestion {
	ID, err := shared.SecureRandomString(32)
	if err != nil {
		panic("could not generate random number")
	}
	item := Suggestion{
		ID:        ID,
		URL:       URL,
		CreatedAt: time.Now(),
		samples:   make([]sample, 0),
	}
	addSuggestion <- item
	return item
}

func AllSuggestions() []Suggestion {
	c := make(chan []Suggestion, 0)
	allSuggestions <- c
	res := <-c
	return res
}

// GetSuggestion returns a suggestion by ID.
func GetSuggestion(ID string) (Suggestion, bool) {
	recv := make(chan Suggestion, 0)
	oneSuggestion <- get{
		id: ID,
		c:  recv,
	}
	v := <-recv
	ok := true
	if v.ID == "" {
		ok = false
	}
	return v, ok
}

func addSample(suggestion Suggestion, sample sample) mod {
	return mod{
		id: suggestion.ID,
		fn: func(s *Suggestion) {
			s.samples = append(s.samples, sample)
		},
	}
}

func removeSample(suggestion Suggestion, samp sample) mod {
	return mod{
		id: suggestion.ID,
		fn: func(s *Suggestion) {
			var samples []sample
			for _, v := range s.samples {
				if v.id != samp.id {
					samples = append(samples, v)
				}
			}
			s.samples = samples
		},
	}
}

func setToken(suggestion Suggestion, token shared.SuggestionToken) mod {
	return mod{
		id: suggestion.ID,
		fn: func(s *Suggestion) {
			s.Token = token
		},
	}
}

// modifySuggestion .
type mod struct {
	id string
	fn func(*Suggestion)
}

type get struct {
	id string
	c  chan Suggestion
}

var (
	modifySuggestion = make(chan mod, 0)
	addSuggestion    = make(chan Suggestion, 0)
	allSuggestions   = make(chan chan []Suggestion, 0)
	oneSuggestion    = make(chan get, 0)
	resetSuggestions = make(chan bool, 0)
	sendSamples      = make(chan int, 0)
)

func init() {
	go func() {
		var suggestions = make(map[string]Suggestion)
		for {
			select {
			case <-resetSuggestions:
				suggestions = make(map[string]Suggestion)

			case one := <-oneSuggestion:
				s, _ := suggestions[one.id]
				go func(s Suggestion, sc chan Suggestion) {
					sc <- s
				}(s, one.c)

			case all := <-allSuggestions:
				var res []Suggestion
				for _, v := range suggestions {
					res = append(res, v)
				}
				go func(s []Suggestion, sc chan []Suggestion) {
					sc <- s
				}(res, all)

			case add := <-addSuggestion:
				suggestions[add.ID] = add

			case modify := <-modifySuggestion:
				s, ok := suggestions[modify.id]
				if ok {
					modify.fn(&s)
					suggestions[modify.id] = s
				} else {
					panic("noooo")
				}
			}
		}
	}()

}
