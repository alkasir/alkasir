// Entrypoint for running as a service.
package server

import (
	"errors"
	"log"
	"os"
)

// checks if it is a good idea to start as a service process
func CheckServiceEnv() bool {
	alkasirService := os.Getenv("ALKASIR_SERVICE")
	return alkasirService != ""
}

var checkers = make([]Check, 0)

func AddChecker(check Check) {
	checkers = append(checkers, check)
}

func findParser() (parser Parser, err error) {

	serviceOpt := NewOption("service")
	if !serviceOpt.Has() {
		return nil, errors.New("no service specified")
	}
	for _, c := range checkers {
		parser, err = c(serviceOpt)
		if err == nil {
			return parser, nil
		}
	}
	return
}

func tryRunService() error {
	server, err := findParser()
	if err != nil {
		return err
	}
	parser, err := server.Parse()
	if err != nil {
		return err
	}
	err = parser.Start()
	return err
}

// start serving t
func RunService() {
	log.SetFlags(log.Lshortfile)
	err := tryRunService()
	if err != nil {
		log.Printf("err: %+v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
