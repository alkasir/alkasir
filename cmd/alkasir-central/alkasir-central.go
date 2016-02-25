package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/alkasir/alkasir/pkg/central"
	"github.com/facebookgo/flagenv"
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
