package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/facebookgo/flagenv"
	"github.com/thomasf/lg"
)

func main() {
	var bindaddr = flag.String("bindaddr", "0.0.0.0:7245", "bind address")
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	flagenv.Prefix = "ALKASIR_WANIP_SERVER_"
	flagenv.Parse()
	lg.CopyStandardLogTo("INFO")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ipAddress, _, _ := net.SplitHostPort(r.RemoteAddr)
		fmt.Fprintf(w, "%s", ipAddress)
		if lg.V(5) {
			lg.Infof("returning %s", ipAddress)
		}
	})
	lg.Infof("Listening to http://%s", *bindaddr)
	err := http.ListenAndServe(*bindaddr, nil)
	if err != nil {
		lg.Fatal(err)
	}
}
