package ptc

import "golang.org/x/net/proxy"

// PTDialMethodConfig creates proxy dails for using PT client/severs
type PTDialMethodConfig struct {
	Args       Args `json:"args"`
	ClientAddr string
	ServerAddr string
}

func (p *PTDialMethodConfig) auth() *proxy.Auth {
	optsstr := encodeClientArgs(p.Args)
	var username, password string
	username = optsstr[0:1]
	password = optsstr[1:]
	auth := &proxy.Auth{
		User:     username,
		Password: password,
	}
	return auth
}

// ClientDailer returns an socks5 proxy dialer for the pluggable transport client connection
func (p PTDialMethodConfig) ClientDialer() (proxy.Dialer, error) {
	// tcp is hardcoded for the time being.
	network := "tcp4"

	auth := p.auth()
	clidial, err := proxy.SOCKS5(network, p.ClientAddr, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return clidial, err

}

// ClientDailer returns an socks5 proxy dialer relevant if the other end of the pluggable transport server is also an socks5 proxy, in that case this proxy dialer can be used to access the internet.
func (p PTDialMethodConfig) ServerSocksDialer() (proxy.Dialer, error) {
	// tcp is hardcoded for the time being.
	network := "tcp4"

	clientDialer, err := p.ClientDialer()
	if err != nil {
		return nil, err
	}

	socksDailer, err := proxy.SOCKS5(network, p.ServerAddr, nil, clientDialer)
	return socksDailer, err
}
