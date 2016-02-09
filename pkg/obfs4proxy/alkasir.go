package obfs4proxy

import "os"

// Config .
type Config struct {
	ShowVer       bool
	LogLevel      string
	EnableLogging bool
	UnsafeLogging bool
}

func Check() bool {
	ptserv := os.Getenv("TOR_PT_SERVER_TRANSPORTS")
	ptcli := os.Getenv("TOR_PT_CLIENT_TRANSPORTS")
	return ptserv != "" || ptcli != ""

}
