package service

// These are the (currently) allowed protocols metods that are exposed to the
// end user web browser/os.
const (
	SOCKS5 BrowserProtocols = 1 << iota
	SOCKS4
	SOCKS4A
	HTTP
	HTTPS
)

type BrowserProtocols int

// Returns a string representation of the constant.
func (b BrowserProtocols) String() string {
	var s string
	switch b {
	case SOCKS5:
		s = "socks5"
	case SOCKS4A:
		s = "socks4a"
	case HTTP:
		s = "http"
	case HTTPS:
		s = "https"
	}
	return s
}
