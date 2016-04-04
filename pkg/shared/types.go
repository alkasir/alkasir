package shared

import (
	"net"
	"net/http"
	"strconv"
	"time"
)


// BrowserSuggestionTokenRequest goes from browser extension to client
type BrowserSuggestionTokenRequest struct {
	URL string
}

// SuggestionTokenRequest is sent from the client to central to initiate a url suggestion flow.
type SuggestionTokenRequest struct {
	URL         string // The url in question
	ClientAddr  net.IP // the public ip address of the client
	CountryCode string // the client country code setting
}

// SuggestionTokenResponse is sent back to the client after processing the SuggestionTokenRequest.
type SuggestionTokenResponse struct {
	*Status
	Ok    bool            // false if the server denies the submitted url to be added.
	URL   string          // The url that was sent in the request, possibly altered.
	Token SuggestionToken // The token to send with all data sampling results for this session.
	Error string
}

// SuggestionToken is a session ID.
type SuggestionToken string

// Sample is the core data structure representing a network test.
type Sample struct {
	Token      SuggestionToken // The Suggestion Token.
	URL        string          // The same URL as the initial submission token session.
	SampleType string          // The sample type.
	Data       string          // The payload in the form of serialized JSON.
}

// StoreSampleRequest for receiving StoreSample API requests.
type StoreSampleRequest struct {
	*Sample
	ClientAddr string // the public ip address of the client
}

// StoreSampleRequest for receiving StoreSample API requests.
type StoreSampleResponse struct {
	Ok    bool
	Error string
}

// SampleResponse is sent to the client for any kind of samples sent to the server.
type SampleResponse struct {
	Ok    bool // True if the sample was accepted by the server.
	Error string
}

// NewClientTokenSample is stored when central accepts a client suggestion token request.
// Central is the only allowed origin for NewClientToken samples.
type NewClientTokenSample struct {
	URL         string // The url
	CountryCode string // The clients configured country code, ie. not the one derived by geoip.
}

// BrowserExtensionSample represents how the browser saw the url and related when it was submitted.
type BrowserExtensionSample struct {
	Title           string              // Page title
	HTTPStatusCode  int                 // the http status code
	ResponseHeaders []map[string]string // HTTP response headers
	RelatedHosts    []string            // Hosts which has been used to requested load page.
	LinkHosts       []string            // Hosts linked to via anchor tags on the page
}

// IPExtraData is the default struct of the IP.
type IPExtraData struct {
	CityGeoNameID uint // The city genoname ID derived from the client's IP address using a geoip lookups.
}

// UpdateHostlistRequest .
type UpdateHostlistRequest struct {
	ClientAddr    net.IP // the public ip address of the client
	UpdateID      string `json:"update_id"`  // unique installation identifer
	ClientVersion string `json:"client_ver"` // client version
}

// UpdateHostlistRequest .
type UpdateHostlistResponse struct {
	Ok    bool
	Error string
	Hosts []string // All hosts listed as blocked in the current region
}

// BlockedContentRequest .
type BlockedContentRequest struct {
	IDMax int `json:"id_max"` // TODO: maybe move
}

func (b BlockedContentRequest) AddParams(req *http.Request) {
	q := req.URL.Query()

	if b.IDMax != 0 {
		q.Add("id_max", strconv.Itoa(b.IDMax))
	}
	req.URL.RawQuery = q.Encode()
}

type HostsPublishLog struct {
	ID          string    `json:"id"`
	Host        string    `json:"host"`
	CountryCode string    `json:"country_code"`
	ASN         string    `json:"asn"`
	CreatedAt   time.Time `json:"created_at"`
	Sticky      bool      `json:"sticky"`
	Action      string    `json:"action"`
}

// SampleRequest .
type ExportSampleRequest struct {
	IDMax int `json:"id_max"` // TODO: maybe move
}

type ExportSimpleSampleRequest struct {
	IDMax int `json:"id_max"` // TODO: maybe move
}

// Sample is the core data structure representing a network test.
type ExportSampleEntry struct {
	ID          string    `json:"id"`
	Host        string    `json:"host"`
	CountryCode string    `json:"country_code"`
	ASN         string    `json:"asn"`
	CreatedAt   time.Time `json:"created_at"`
	Origin      string    `json:"origin"`
	Type        string    `json:"type"`
	Token       string    `json:"token"`
	Data        string    `json:"data"`
	ExtraData   string    `json:"extra_data"`
}

// Sample is the core data structure representing a network test.
type ExportSimpleSampleEntry struct {
	ID          string    `json:"id"`
	CountryCode string    `json:"country_code"`
	ASN         string    `json:"asn"`
	CreatedAt   time.Time `json:"created_at"`
	Type        string    `json:"type"`
	OriginID    string    `json:"origin_id"`
	Data        string    `json:"data"`
}

// BinaryUpgradeRequest .
type BinaryUpgradeRequest struct {
	Artifact    string `json:"artifact"`
	FromVersion string `json:"fromVersion"`
}

// UpgradeMeta .
type BinaryUpgradeResponse struct {
	Artifact         string    `json:"artifact"`
	Version          string    `json:"version"`
	CreatedAt        time.Time `json:"createdAt"`
	SHA256Sum        string    `json:"sha256Sum"`
	ED25519Signature string    `json:"ed25519Sig"`
}
