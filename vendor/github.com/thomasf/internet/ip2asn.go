package internet

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// NewIP2ASNClient returns a new client for IP address to AS Number queries.
// The caller is reposible for closing the conn when the client is not used
// anymore.
func NewIP2ASNClient(conn redis.Conn) *IP2ASNClient {
	ipasn := &IP2ASNClient{
		conn: conn,
	}
	return ipasn
}

// ASNResult contains the lookup results from resolving an IP address to a ASN.
type ASNResult struct {
	Mask net.IPNet
	ASN  int
	Date time.Time // Recorded time for when
}

// IP2ASNClient is the query client.
type IP2ASNClient struct {
	conn redis.Conn
}

// Current returns the latest known result for an IP2ASN lookup.
func (i *IP2ASNClient) Current(IP string) (*ASNResult, error) {
	ip, err := i.parseIP(IP)
	if err != nil {
		return &ASNResult{}, err
	}
	var current string
	allDates, err := i.importedDates()
	if err != nil {
		return &ASNResult{}, err
	}

	if len(allDates) < 0 {
		return &ASNResult{}, err
	}
	current = allDates[len(allDates)-1]

	results, err := i.dates(ip, []string{current})
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		return &results[0], nil
	}
	return nil, nil
}

// AllHistory returns the full history for the given IP address.
func (i *IP2ASNClient) AllHistory(IP string) ([]ASNResult, error) {
	ip, err := i.parseIP(IP)
	if err != nil {
		return []ASNResult{}, err
	}
	dates, err := i.importedDates()
	if err != nil {
		return []ASNResult{}, err
	}
	return i.dates(ip, dates)
}

func (i *IP2ASNClient) parseIP(IP string) (net.IP, error) {
	I := net.ParseIP(IP)
	if I == nil {
		return nil, net.InvalidAddrError(IP)
	}
	return I, nil
}

func (i *IP2ASNClient) keys(IP net.IP) []net.IPNet {
	var keys []net.IPNet
	for _, n := range netmasks {
		ipn := net.IPNet{
			IP:   IP.Mask(n),
			Mask: n,
		}
		keys = append(keys, ipn)
	}
	return keys
}

// Stringer.
func (a *ASNResult) String() string {
	return fmt.Sprintf("%s %d %s",
		a.Date.Format("2006-01-02"),
		a.ASN,
		a.Mask.String())
}

// importedDates fetches all imported dates from redis
func (i *IP2ASNClient) importedDates() ([]string, error) {
	return redis.Strings(i.conn.Do("SMEMBERS", "i2a:imported_dates"))
}

// dates resolves IP2ASN for all date entries, if available.
func (i *IP2ASNClient) dates(IP net.IP, dates []string) ([]ASNResult, error) {
	keys := i.keys(IP)
	for _, d := range dates {
		for _, k := range keys {
			i.conn.Send("HGET", "i2a:"+k.String(), d)
		}
	}
	i.conn.Flush()
	var results []ASNResult
	for _, date := range dates {

		var found bool

		for idx := 0; idx < len(keys); idx++ {
			if !found {
				r, err := redis.String(i.conn.Receive())
				if err != nil {
					if err == redis.ErrNil {
						continue
					}
					return []ASNResult{}, err
				}

				timedate, err := time.Parse("20060102", date)
				asn, err := strconv.Atoi(r)
				if err != nil {
					// redis data error
					return []ASNResult{}, err
				}
				results = append(results, ASNResult{
					Mask: keys[idx],
					ASN:  asn,
					Date: timedate,
				})
				found = true
			} else {
				_, _ = i.conn.Receive()
			}
		}

	}
	return results, nil
}

// RIPE-NCC-RIS BGP IPv6 Anchor Prefix @RRC00
// RIPE-NCC-RIS BGP Anchor Prefix @ rrc00 - RIPE NCC
var (
	asn12654blocks = map[string]bool{
		"2001:7fb:ff00::/48": true,
		"84.205.80.0/24":     true,
		"2001:7fb:fe00::/48": true,
		"84.205.64.0/24":     true,
	}
	netmasks []net.IPMask
)

func init() {
	for i := 0; i < 8*net.IPv4len; i++ {
		netmasks = append(netmasks, net.CIDRMask(8*net.IPv4len-i, 8*net.IPv4len))
	}
}
