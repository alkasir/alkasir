package sampletypes

type SampleType int

//go:generate stringer -type=SampleType
const (
	None SampleType = iota
	HTTPHeader
	BrowserExtension
	NewClientToken
	DNSQuery
)
