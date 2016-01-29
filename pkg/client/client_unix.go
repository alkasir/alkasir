// +build !windows

package client

import (
	"os"
	"os/signal"
	"syscall"
)

var sigIntC chan os.Signal

func init() {
	sigIntC = make(chan os.Signal)
	signal.Notify(sigIntC, syscall.SIGINT)
}
