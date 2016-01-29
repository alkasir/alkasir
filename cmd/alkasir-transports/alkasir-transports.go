// Alkasir transport bundle
package main

import (
	"flag"

	_ "github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/service/server"
	"github.com/alkasir/alkasir/pkg/transport/shadowsocks"
	"github.com/alkasir/alkasir/pkg/transport/socks5"
	"github.com/alkasir/alkasir/pkg/transport/torpt"
)

func main() {
	flag.Parse()
	server.AddChecker(socks5.Check)
	server.AddChecker(shadowsocks.Check)
	server.AddChecker(torpt.Check)
	server.RunService()
}
