// The alkasir client binary
package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/facebookgo/flagenv"

	"github.com/alkasir/alkasir/pkg/obfs4proxy"
	_ "github.com/thomasf/lg"

	"github.com/alkasir/alkasir/pkg/client"
	"github.com/alkasir/alkasir/pkg/client/ui"
	"github.com/alkasir/alkasir/pkg/client/ui/cliui"

	"github.com/alkasir/alkasir/pkg/service/server"
	"github.com/alkasir/alkasir/pkg/transport/shadowsocks"
	"github.com/alkasir/alkasir/pkg/transport/socks5"
	"github.com/alkasir/alkasir/pkg/transport/torpt"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var errors []error
	errors = append(errors,
		flag.Set("logtomemory", "true"),
	)
	// set more flag defaults if debug is not enabled
	if os.Getenv("ALKASIR_DEBUG") == "" {
		errors = append(errors,
			flag.Set("logtofile", "false"),
			flag.Set("v", "19"),
		)
	}
	for _, err := range errors {
		if err != nil {
			panic(err)
		}
	}

	flag.Parse()
	flagenv.Prefix = "ALKASIR_"
	flagenv.Parse()
	server.AddChecker(socks5.Check)
	server.AddChecker(torpt.Check)
	server.AddChecker(shadowsocks.Check)

	if server.CheckServiceEnv() {
		torpt.SetConfigDir(client.ConfigPath("torpt"))
		server.RunService()
	} else if obfs4proxy.Check() {
		obfs4proxy.Run(obfs4proxy.Config{})
	} else {
		ui.Set(cliui.New())
		client.SetUpgradeArtifact("alkasir-client")
		client.Run()
	}
}
