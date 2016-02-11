package clientconfig

import (
	"crypto/md5"
	"encoding/hex"
	"errors"

	"github.com/alkasir/alkasir/pkg/browsercode"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/thomasf/lg"
)

// modifyConnections .
type modifyConnections struct {
	Connections []shared.Connection
	Remove      []string // values of shared.Connection.Hash()
	Add         []string // values of shared.Connection.Encode()
	Protect     []string // valiues of shared.Connection.ID
}

func (m *modifyConnections) Update() []shared.Connection {

	lg.Infoln("updating connections..")

	if lg.V(19) {
		lg.Infof("pre upgrade state:")
		for _, v := range m.Connections {
			lg.Infoln(v)

		}
	}

	// create map for id lookups
	conns := make(map[string]shared.Connection, 0)
	for _, connection := range m.Connections {
		conns[connection.ID] = connection
	}

	// remove old old connections
	for _, ID := range m.Remove {
		if _, ok := conns[ID]; ok {
			lg.V(19).Infof("remove connection: %s", ID)
			delete(conns, ID)
		}
	}

	// add new connections
	for _, v := range m.Add {
		conn, err := shared.DecodeConnection(v)
		if err != nil {
			lg.Fatal(err)
		}
		ID := conn.ID
		if _, ok := conns[ID]; !ok {
			lg.V(19).Infof("add connection: %s", ID)
			conns[ID] = conn
		}
	}

	// protect connections
	for _, ID := range m.Protect {
		if _, ok := conns[ID]; ok {
			c := conns[ID]
			c.Protected = true
			conns[ID] = c
			lg.V(19).Infof("protected connection: %s", ID)
		}
	}

	var result []shared.Connection
	for _, v := range conns {
		result = append(result, v)
	}
	if lg.V(19) {
		lg.Infof("upgraded connections result:")
		for _, v := range result {
			lg.Infoln(v)

		}
	}
	return result
}

// the default central server url, the value value is overridden on release builds
var centralAddr string

// UpgradeConfig updates the config, if needed.
func UpgradeConfig() (bool, error) {
	currentConrfigMu.Lock()

	if centralAddr == "" {
		centralAddr = "https://localhost:8080"
		lg.Warningln("WARNING centralAddr not set, defaulting to dev value: ", centralAddr)
	}

	defer currentConrfigMu.Unlock()
	if !currentConfig.configRead {
		return false, errors.New("config not read")
	}
	prevVer := currentConfig.Settings.Version

	switch currentConfig.Settings.Version {
	case 0: // when config file was created from template
		currentConfig.Settings.Version = 1
		currentConfig.Settings.Local.CentralAddr = centralAddr
		fallthrough
	case 1:
		lg.Infoln("updating configuration to v1")
		m := &modifyConnections{
			Connections: currentConfig.Settings.Connections,
			Add: []string{
				"aieyJ0Ijoib2JmczQiLCJzIjoiXCJjZXJ0PUdzVFAxVmNwcjBJeE9aUkNnUHZ0Z1JsZVJWQzFTRmpYWVhxSDVvSEhYVFJ4M25QVnZXN2xHK2RkREhKWmw4YjBOVFZ1VGc7aWF0LW1vZGU9MFwiIiwiYSI6IjEzOS4xNjIuMjIxLjEyMzo0NDMifQ==",
				"aieyJ0Ijoib2JmczQiLCJzIjoiXCJjZXJ0PTMzdXNkSUVDemFyRUpsWFpQM0w0Y2x0bi9vNXdhVUlOcHRUU0JSYk5BQVpVcVlsajhiZm5MWGYyb3BFNHE2c0NlYzY3Ync7aWF0LW1vZGU9MFwiIiwiYSI6IjQ2LjIxLjEwMi4xMDk6NDQzIn0=",
			},
			Remove: []string{
				// NOTE: old hash function used, not a current format ID
				"xFa9T1i6bJMJIvK6kxFA1xvQGfW58BY3OLkrPXbpvAY=",
			},
		}
		currentConfig.Settings.Connections = m.Update()
		currentConfig.Settings.Version = 2
		fallthrough
	case 2:
		m := &modifyConnections{
			Connections: currentConfig.Settings.Connections,
			Protect: []string{
				"z0VZS-Kx9tMfoBEyX6br19GaJgKDe0IK0i6JKyhKp2s",
				"ipJ2oW8xr9TFDvfU92qGvDaPwZttf_GSjGZ4KW7inBI",
			},
			Remove: []string{
				"Ul3D2G1dI3_Z4sLXQ9IUQdIFH4pSDyNjTwf_auy93Os",
			},
		}
		currentConfig.Settings.Connections = m.Update()
		currentConfig.Settings.Version = 3
		fallthrough
	case 3:
		key := browsercode.NewKey()
		currentConfig.Settings.Local.ClientAuthKey = key
		currentConfig.Settings.Local.ClientBindAddr = "127.0.0.1:8899"
		currentConfig.Settings.Version = 4
		fallthrough
	case 4:
		// If the beta url to central server was set, replace it with the default value
		s := md5.New()
		s.Write([]byte(currentConfig.Settings.Local.CentralAddr))
		hsum := hex.EncodeToString(s.Sum(nil))
		if hsum == "796547bd38f0d8722c2e0802de34b53f" {
			currentConfig.Settings.Local.CentralAddr = centralAddr
		}
		currentConfig.Settings.Version = 5
		fallthrough
	case 5:
		lg.Infoln("Settings version", currentConfig.Settings.Version)
	default:
		lg.Errorln("Future configuration version!", currentConfig.Settings.Version)
	}
	return currentConfig.Settings.Version != prevVer, nil
}
