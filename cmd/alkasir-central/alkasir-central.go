package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/facebookgo/flagenv"
	"github.com/alkasir/alkasir/pkg/central"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	flagenv.Prefix = "ALKASIR_"
	flagenv.Parse()
	err := central.Init()
	if err == nil {
		central.Run()
	}

}
