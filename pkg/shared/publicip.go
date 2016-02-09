package shared

import (
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/thomasf/lg"
)

// shuffle is a Fisher-Yates shuffle
func shuffleStrings(slc []string) {
	N := len(slc)
	for i := 0; i < N; i++ {
		// choose index uniformly in [i, N-1]
		r := i + rand.Intn(N-i)
		slc[r], slc[i] = slc[i], slc[r]
	}
}

var wanipservices = []string{
	"https://ip.appspot.com",
	"http://utils.admin-linux.fr/ip.php",
	"http://bot.whatismyipaddress.com",
	"http://ip-api.com/line/?fields=query",
	"http://ipinfo.io/ip",
	"http://icanhazip.com",
	"http://www.trackip.net/ip",
	"http://myexternalip.com/raw",
	"https://whatismyip.herokuapp.com/",
}

var publicIP struct {
	sync.Mutex
	ip *net.IP

	init  sync.Once
	hasIP sync.WaitGroup
}

func GetPublicIPAddr() net.IP {
	publicIP.init.Do(func() {
		lg.Infoln("starting public ip address updater")

		var gotIP sync.Once
		publicIP.hasIP.Add(1)

		go func() {
			timeout := time.Duration(10 * time.Second)
			client := http.Client{
				Timeout:   timeout,
				Transport: v4Transport,
			}
			var services []string = make([]string, 0)
			services = append(services, wanipservices...)
			shuffleStrings(services)
			serviceIdx := 0
			refreshTicker := time.Tick(time.Minute * 10)
		loop:
			for {
				serviceIdx = (serviceIdx + 1) % (len(services))
				if serviceIdx == 0 {
					gotIP.Do(func() { publicIP.hasIP.Done() })
				}
				URL := services[serviceIdx]
				resp, err := client.Get(URL)

				if err != nil {
					lg.Warningf("Could not read response from %s: %v", URL, err)
					<-time.After(time.Second * 2)
					continue loop

				}
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					lg.Warningf("Could not read response from %s: %v", URL, err)
					resp.Body.Close()
					<-time.After(time.Second * 2)
					continue loop
				}
				resp.Body.Close()
				str := strings.TrimSpace(string(data))
				ip := net.ParseIP(str)
				if ip == nil {
					lg.Warningf("error parsing ip from %s from: %.50s", URL, str)
					<-time.After(time.Second * 2)
					continue loop
				}
				publicIP.Lock()
				if publicIP.ip == nil || !publicIP.ip.Equal(ip) {
					lg.V(5).Infof("Public ip address change: %s -> %s via %s", publicIP.ip, ip, URL)
				}
				publicIP.ip = &ip
				publicIP.Unlock()
				gotIP.Do(func() { publicIP.hasIP.Done() })
				select {
				case <-refreshTicker:
					lg.V(30).Infoln("refreshing IP address")
				}
			}
		}()
	})
	publicIP.hasIP.Wait()
	publicIP.Lock()
	ip := publicIP.ip
	publicIP.Unlock()
	if ip != nil {
		return ip.To4()
	}
	return nil
}

// v4Dail is an ipv4 only tcp dialer.
type v4Dial struct {
	*net.Dialer
}

func (d *v4Dial) Dial(network, address string) (net.Conn, error) {
	return d.Dialer.Dial("tcp4", address)
}

var v4Transport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&v4Dial{
		&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}).Dial,
	TLSHandshakeTimeout: 10 * time.Second,
}
