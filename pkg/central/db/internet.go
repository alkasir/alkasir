package db

import (
	"net"

	"github.com/garyburd/redigo/redis"
	"github.com/thomasf/internet"
)

// Internet .
type Internet struct {
	*redis.Pool
}

type InternetClient interface {
	IP2ASN(IP net.IP) (*internet.ASNResult, error)
}

func NewInternetClient(pool *redis.Pool) *Internet {
	return &Internet{pool}
}

func (i *Internet) IP2ASN(IP net.IP) (*internet.ASNResult, error) {
	rconn := i.Get()
	r, e := internet.NewIP2ASNClient(rconn).Current(IP.String())
	rconn.Close()
	return r, e
}
