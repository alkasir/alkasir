package ptc

import (
	"log"
	"path/filepath"
	"strings"
)

// Client
type Client struct {
	Config
	Proxy   string   `json:"proxy"`
	Methods []string `json:"methods"`
}

// Start starts the client process, returns error if start up failes
func (s *Client) Start() error {
	s.Config.env = s.encodeEnv()
	log.Println(s.Config.env)
	err := s.start()
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Client) encodeEnv() []string {
	var env []string
	env = append(env,
		"TOR_PT_MANAGED_TRANSPORT_VER=1",
		"TOR_PT_STATE_LOCATION="+filepath.Join(s.StateLocation, "ptc"),
		"TOR_PT_CLIENT_TRANSPORTS="+strings.Join(s.Methods, ","),
	)
	if s.Proxy != "" {
		env = append(env, "TOR_PT_PROXY="+s.Proxy)
	}
	return env
}
