package ptc

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

// Represents a server transport plugin configuration like:
// 	ServerTransportPlugin MethodName exec Command
//  ServerTransportPlugin <transports> exec <path> [<options>]
type SMethod struct {
	MethodName string `json:"name"`
	BindAddr   string `json:"bindaddr"`
	Args       Args   `json:"args"`
}

// ServerConfig .
type Server struct {
	Config
	Methods      []SMethod `json:"methods"`
	ExtendedPort string    `json:"extport"`
	ORPort       string    `json:"orport"`
}

// Start starts the pluggable transport server process, returns error if start up fails.
func (s *Server) Start() error {
	s.Config.env = s.encodeEnv()
	log.Println(s.Config.env)
	err := s.start()
	if err != nil {
		log.Println(err)

	}
	return err
}

func (s *Server) encodeEnv() []string {
	methods := s.Methods
	var transports, transportOptions, bindaddr []string

	escape := func(s string) string {
		repl := strings.NewReplacer(":", "\\:", ";", "\\;", "=", "\\=", "\\", "\\\\")
		return repl.Replace(s)
	}

	for _, method := range methods {
		transports = append(transports, method.MethodName)
		bindaddr = append(bindaddr, fmt.Sprintf("%s-%s", method.MethodName, method.BindAddr))
		for key := range method.Args {
			for _, value := range method.Args[key] {
				transportOptions = append(transportOptions,
					escape(method.MethodName)+":"+escape(key)+"="+escape(value))
			}
		}
	}

	var env []string
	env = append(env,
		"TOR_PT_MANAGED_TRANSPORT_VER=1",
		"TOR_PT_STATE_LOCATION="+filepath.Join(s.StateLocation, "ptc"),
		"TOR_PT_EXTENDED_SERVER_PORT="+s.ExtendedPort,
		"TOR_PT_ORPORT="+s.ORPort,
		"TOR_PT_SERVER_BINDADDR="+strings.Join(bindaddr, ","),
		"TOR_PT_SERVER_TRANSPORTS="+strings.Join(transports, ","),
	)

	joinedOpts := strings.Join(transportOptions, ";")
	if joinedOpts != "" {
		env = append(env,
			"TOR_PT_SERVER_TRANSPORT_OPTIONS="+strings.Join(transportOptions, ";"),
		)
	}
	return env
}
