// pac browser config generator
//
package pac

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/alkasir/alkasir/pkg/service"
	"github.com/thomasf/lg"
)

var pac struct {
	template       *template.Template
	topLevelDomain string
	directList     string
	blockedList    string
	defaultMethod  string
	blockedMethod  string
	dLRWMutex      sync.RWMutex
}

var topLevelDomain = map[string]bool{
	"ac":  true,
	"co":  true,
	"com": true,
	"edu": true,
	"gov": true,
	"net": true,
	"org": true,
	"se":  true,
}

func init() {
	var err error
	pac.template, err = template.New("pac").Parse(pacRawTmpl)
	pac.defaultMethod = "DIRECT"
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	for k := range topLevelDomain {
		buf.WriteString(fmt.Sprintf("\t\"%s\": true,\n", k))
	}
	pac.topLevelDomain = buf.String()[:buf.Len()-2] // remove the final comma

	go func() {
		connectionEventListener := make(chan service.ConnectionHistory)
		service.AddListener(connectionEventListener)
		for {
			select {
			case event := <-connectionEventListener:
				if event.IsUp() {
					current := event.Current()
					s, ok := service.ManagedServices.Service(current.ServiceID)
					if ok {
						response := s.Response
						SetBlockedMethod(response["protocol"], response["bindaddr"])
					} else {
						SetBlockedMethod("DIRECT", "")
					}
				} else if event.Current().State == service.Ended {
					SetBlockedMethod("DIRECT", "")
				}
			}
		}
	}()
}

// GenPAC regenerates the proxy auto configration file file
func GenPAC() []byte {
	buf := new(bytes.Buffer)

	direct := getDirectList()
	blocked := getBlockedList()

	data := struct {
		BlockedMethod  string
		DirectDomains  string
		BlockedDomains string
		TopLevel       string
		DefaultMethod  string
	}{
		pac.blockedMethod,
		direct,
		blocked,
		pac.topLevelDomain,
		pac.defaultMethod,
	}

	if err := pac.template.Execute(buf, data); err != nil {
		lg.Infoln("Error generating pac file:", err)
		panic("Error generating pac file")
	}
	return buf.Bytes()
}

func getDirectList() string {
	pac.dLRWMutex.RLock()
	dl := pac.directList
	pac.dLRWMutex.RUnlock()
	return dl
}

// UpdateBlockedList updates the list of hosts that are not going through any
// proxy.
func UpdateDirectList(hosts []string) {
	var escaped []string
	for _, v := range hosts {
		escaped = append(escaped, template.JSEscapeString(v))
	}
	dl := strings.Join(escaped, "\",\n\"")
	pac.dLRWMutex.Lock()
	pac.directList = dl
	pac.dLRWMutex.Unlock()
}

func getBlockedList() string {
	pac.dLRWMutex.RLock()
	dl := pac.blockedList
	pac.dLRWMutex.RUnlock()
	return dl
}

// UpdateBlockedList updates the list of hosts that are going through the
// blocked proxy configuration.
func UpdateBlockedList(hosts ...[]string) {
	var allHosts []string
	allHosts = append(allHosts, "alkasir.com")
	for _, h := range hosts {
		allHosts = append(allHosts, h...)
	}

	var escaped []string
	for _, v := range allHosts {
		escaped = append(escaped, template.JSEscapeString(v))
	}
	dl := strings.Join(escaped, "\",\n\"")

	pac.dLRWMutex.Lock()
	pac.blockedList = dl
	pac.dLRWMutex.Unlock()
}

func getTransport() string {
	pac.dLRWMutex.RLock()
	dl := pac.blockedMethod
	pac.dLRWMutex.RUnlock()
	return dl
}

// SetBlockedMethod updates the address of the proxy used when blocked
func SetBlockedMethod(protocol, connectstr string) {
	pac.dLRWMutex.Lock()
	defer pac.dLRWMutex.Unlock()
	switch protocol {
	case "socks5":
		pac.blockedMethod = "SOCKS5 " + connectstr
	case "DIRECT":
		pac.blockedMethod = "DIRECT"
	default:
		panic(errors.New(fmt.Sprintf("protocol not supported: %s", protocol)))
	}
}
