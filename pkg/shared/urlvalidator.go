package shared

import (
	"net"
	"net/url"
	"strings"

	"github.com/thomasf/lg"
)

var (
	disallowedNets  []*net.IPNet
	disallowedPorts map[int]bool
	disallowedHosts = map[string]bool{
		"localhost": true,
	}
	allowedProtocols = map[string]bool{
		"http":  true,
		"https": true,
	}
)

// AcceptedURL returns true if the supplied url is allowed to be added to alkasir.
func AcceptedURL(u *url.URL) bool {
	if _, ok := allowedProtocols[u.Scheme]; !ok {
		if lg.V(50) {
			lg.Warningf("url scheme %s is not allowed", u.Scheme)
		}
		return false
	}
	return AcceptedHost(u.Host)

}

// AcceptedHost return true if the supplied host:port or host is allowed to be added to alkasir.
func AcceptedHost(host string) bool {
	if strings.TrimSpace(host) == "" {
		if lg.V(50) {
			lg.Warningf("empty url host is not allowed")
		}
		return false
	}
	if strings.Contains(host, ":") {
		onlyhost, _, err := net.SplitHostPort(host)
		if err == nil {
			host = onlyhost
		}
	}
	if _, ok := disallowedHosts[host]; ok {
		if lg.V(50) {
			lg.Warningf("url host %s is not allowed", host)
		}
		return false
	}
	IP := net.ParseIP(host)
	if IP != nil {
		return AcceptedIP(IP)
	}
	return true
}

// AcceptedURL returns true if the supplied IP is allowed to be added to alkasir.
func AcceptedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.Equal(net.IPv4zero) || ip.Equal(net.IPv6zero) {
		if lg.V(50) {
			lg.Warningf("ip address %s is not allowed because loopback or zeoro address", ip.String())
		}
		return false
	}
	for _, v := range disallowedNets {
		if v.Contains(ip) {
			if lg.V(50) {
				lg.Warningf("ip %s is not allowed because network %s is not allowed", ip.String(), v.String())
			}
			return false
		}
	}
	return true
}

// AcceptedPort
func AcceptedPort(port int) bool {
	return !disallowedPorts[port]
}

func init() {
	for _, v := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	} {
		_, ipnet, err := net.ParseCIDR(v)
		if err != nil {
			panic(err)
		}
		disallowedNets = append(disallowedNets, ipnet)
	}

	badPorts := []int{
		1,    // tcpmux
		7,    // echo
		9,    // discard
		11,   // systat
		13,   // daytime
		15,   // netstat
		17,   // qotd
		19,   // chargen
		20,   // ftp data
		21,   // ftp access
		22,   // ssh
		23,   // telnet
		25,   // smtp
		37,   // time
		42,   // name
		43,   // nicname
		53,   // domain
		77,   // priv-rjs
		79,   // finger
		87,   // ttylink
		95,   // supdup
		101,  // hostriame
		102,  // iso-tsap
		103,  // gppitnp
		104,  // acr-nema
		109,  // pop2
		110,  // pop3
		111,  // sunrpc
		113,  // auth
		115,  // sftp
		117,  // uucp-path
		119,  // nntp
		123,  // NTP
		135,  // loc-srv /epmap
		139,  // netbios
		143,  // imap2
		179,  // BGP
		389,  // ldap
		465,  // smtp+ssl
		512,  // print / exec
		513,  // login
		514,  // shell
		515,  // printer
		526,  // tempo
		530,  // courier
		531,  // chat
		532,  // netnews
		540,  // uucp
		556,  // remotefs
		563,  // nntp+ssl
		587,  // stmp?
		601,  // ??
		636,  // ldap+ssl
		993,  // ldap+ssl
		995,  // pop3+ssl
		2049, // nfs
		3659, // apple-sasl / PasswordServer
		4045, // lockd
		6000, // X11
		6665, // Alternate IRC [Apple addition]
		6666, // Alternate IRC [Apple addition]
		6667, // Standard IRC [Apple addition]
		6668, // Alternate IRC [Apple addition]
		6669, // Alternate IRC [Apple addition]
	}
	disallowedPorts = make(map[int]bool, 0)
	for _, v := range badPorts {
		disallowedPorts[v] = true
	}

}
