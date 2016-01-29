// The alkasir client binary
package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/facebookgo/flagenv"

	"github.com/alkasir/alkasir/pkg/obfs4proxy"

	"github.com/alkasir/alkasir/pkg/client"
	"github.com/alkasir/alkasir/pkg/client/ui"
	"github.com/alkasir/alkasir/pkg/client/ui/wm"

	"os"

	"github.com/alkasir/alkasir/pkg/service/server"
	"github.com/alkasir/alkasir/pkg/transport/shadowsocks"
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

	server.AddChecker(torpt.Check)
	server.AddChecker(shadowsocks.Check)

	if server.CheckServiceEnv() {
		torpt.SetConfigDir(client.ConfigPath("torpt"))
		server.RunService()
	} else if obfs4proxy.Check() {
		obfs4proxy.Run(obfs4proxy.Config{})
	} else {
		ui.Set(wm.New())
		client.SetUpgradeArtifact("alkasir-gui")
		client.Run()
	}
}
