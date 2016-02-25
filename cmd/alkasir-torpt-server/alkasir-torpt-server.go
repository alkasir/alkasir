package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/alkasir/alkasir/pkg/obfs4proxy"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/alkasir/ptc"
	"github.com/armon/go-socks5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thomasf/lg"
)

var (
	allowedConnects prometheus.Counter
	deniedConnects  prometheus.Counter
)

type PermitCommand struct{}

func (p PermitCommand) Allow(req *socks5.Request) bool {
	r := shared.AcceptedIP(req.DestAddr.IP) && shared.AcceptedPort(req.DestAddr.Port)
	if r {
		allowedConnects.Inc()
	} else {
		deniedConnects.Inc()
	}
	return r
}

func socks5server(bindaddr string) {
	conf := &socks5.Config{
		Rules: PermitCommand{},
	}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	if err := server.ListenAndServe("tcp", bindaddr); err != nil {
		panic(err)
	}
}

func startMonitoring(addr string) {
	allowedConnects = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "connect_allow_total",
		Help: "Total started connections",
	})
	deniedConnects = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "connect_deny_total",
		Help: "Total denied connections",
	})
	prometheus.MustRegister(allowedConnects)
	prometheus.MustRegister(deniedConnects)

	http.Handle("/metrics", prometheus.Handler())

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
		_ = err
	}
}

func main() {

	lg.SetSrcHighlight("alkasir/cmd", "alkasir/pkg")
	lg.CopyStandardLogTo("INFO")
	lg.V(1).Info("Log v-level:", lg.Verbosity())
	lg.V(1).Info("Active country codes:", shared.CountryCodes)
	lg.Flush()

	if obfs4proxy.Check() {
		monitorBindAddr := flag.String("monitorAddr", "localhost:8037", "port to bind monitor server to")
		flag.Parse()
		go startMonitoring(*monitorBindAddr)

		obfs4proxy.Run(obfs4proxy.Config{
			LogLevel:      "INFO",
			EnableLogging: true,
			UnsafeLogging: false,
		})
		os.Exit(0)
	}

	monitorBindAddr := flag.String("monitorAddr", "localhost:8036", "port to bind monitor server to")
	flag.Parse()
	go startMonitoring(*monitorBindAddr)

	// server := ptc.Server{
	// 	Config: ptc.Config{
	// 		StateLocation: ".",
	// 		Command:       os.Args[0],
	// 	},
	// 	ORPort: "127.0.0.1:23450",
	// 	Methods: []ptc.SMethod{ptc.SMethod{
	// 		MethodName: "obfs4",
	// 		BindAddr:   "127.0.0.1:4423",
	// 		Args: ptc.Args{
	// 			"node-id":     []string{"4d3e4561149907025571827a1277661b5f9fca46"},
	// 			"private-key": []string{"7886017adfd178cd139e91250dddee0b2af8ab6cf93f1fe9a7a469d3a13a3067"},
	// 			"public-key":  []string{"4393d7641042620a60881a72ebb47ef8c5a5840acb4f858e5c0026cb2e75fd6f"},
	// 			"drbg-seed":   []string{"c23e876ddc408cc392317e017a6796a96161f76d8dd90522"},
	// 			"iat-mode":    []string{"0"},
	// 		},
	// 	},
	// 	},
	// }

	configdir := os.Getenv("ALKASIR_TORPT_CONFIGDIR")
	configfile := filepath.Join(configdir, "alkasir-torpt-server.json")

	data, err := ioutil.ReadFile(configfile)
	if err != nil {
		panic(err)
	}

	server := ptc.Server{}
	err = json.Unmarshal(data, &server)
	if err != nil {
		panic(err)
	}
	server.Command = os.Args[0]

	go socks5server(server.ORPort)

	err = server.Start()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.Wait()
		if err != nil {
			log.Println(err)
		}
	}()
	wg.Wait()

}
