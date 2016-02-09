package client

import (
	"net"
	"net/url"
	"strings"

	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/central/client"
	"github.com/alkasir/alkasir/pkg/client/internal/config"
	"github.com/alkasir/alkasir/pkg/service"
	"github.com/alkasir/alkasir/pkg/shared"
)

// deprecated, function moved to shared package
func getPublicIPAddr() net.IP {
	return shared.GetPublicIPAddr()
}

// NewRestClient returns an central server client using the current default
// transport if the central server is not runing locally.
func NewRestClient() (*client.Client, error) {
	conf := clientconfig.Get()
	apiurl := conf.Settings.Local.CentralAddr
	u, err := url.Parse(apiurl)
	if err != nil {
		return nil, err
	}
	host := u.Host
	if strings.Contains(host, ":") {
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return nil, err
		}
	}

	if host == "localhost" || host == "127.0.0.1" {
		lg.V(19).Infoln("Opening restclient to localhost central api without transport")
		return client.NewClient(apiurl, nil), nil
	}

	httpclient, err := service.NewTransportHTTPClient()
	if err != nil {
		return nil, err
	}
	lg.V(19).Infoln("Opening restclient thru transport")
	return client.NewClient(apiurl, httpclient), nil
}

// ConfigPath - TODO: deprecate or something...
func ConfigPath(file ...string) string {
	return clientconfig.ConfigPath(file...)
}
