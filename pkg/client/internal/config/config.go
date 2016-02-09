// Configuration file access
package clientconfig

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/osutil"
	"github.com/alkasir/alkasir/pkg/shared"
)

// The current configuration in memory
var currentConfig = &Config{}
var currentConrfigMu sync.RWMutex

// ModifyConfig wraps a function which modifies the current configuration
func Update(f func(conf *Config) error) error {
	currentConrfigMu.Lock()
	defer currentConrfigMu.Unlock()
	return f(currentConfig)
}

// GetConfig returns the current config
func Get() Config {
	currentConrfigMu.RLock()
	defer currentConrfigMu.RUnlock()
	return *currentConfig
}

// ReadConfig read all app settings from available config files and/or defalts.
func Read() (*Config, error) {
	err := readSettings(currentConfig)
	if err != nil {
		return nil, err
	}
	err = readHostFiles(currentConfig)
	if err != nil {
		lg.Infoln("could not read host lists..")
	}
	currentConfig.configRead = true
	return currentConfig, nil
}

// Write delegates write to everything that is persisted below this level.
func Write() error {
	if lg.V(15) {
		lg.InfoDepth(1, "called config write")
	}

	currentConrfigMu.RLock()
	defer currentConrfigMu.RUnlock()

	filename := ConfigPath("settings.json")
	lg.V(5).Infoln("Saving settings file")
	data, err := json.MarshalIndent(&currentConfig.Settings, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil

}

// Config is the central configuration holding data type. Ususally there will
// be one configuration instance per app instance.
type Config struct {
	Settings            Settings  // main config file
	BlockedHostsCentral HostsFile // config file with all host that must be proxied which are received from the central database
	BlockedHosts        HostsFile // config file with all host that must be proxied
	DirectHosts         HostsFile // config file with all hosts that must not be proxied
	configRead          bool      // Are the the configuration files read?
}

func readHostFiles(c *Config) error {
	errStr := ""
	localBlocked := &HostsFile{
		Name:        "local-blocked",
		CountryCode: c.Settings.Local.CountryCode,
	}
	localDirect := &HostsFile{
		Name:        "local-direct",
		CountryCode: c.Settings.Local.CountryCode,
	}
	centralBlocked := &HostsFile{
		Name:        "central-blocked",
		CountryCode: c.Settings.Local.CountryCode,
	}
	err := centralBlocked.Read(ConfigPath())
	if err != nil {
		errStr += "Could not read central-blocked. "
	}

	c.BlockedHosts = *localBlocked
	c.BlockedHostsCentral = *centralBlocked
	c.DirectHosts = *localDirect

	if errStr != "" {
		return errors.New(errStr)
	}
	return nil
}

// SetConfig reads configuration from json byte stream
func parseConfig(config []byte) (*Settings, error) {
	s := &Settings{}
	err := json.Unmarshal(config, &s)
	if err != nil {
		return nil, err
	}

	for i, c := range s.Connections {
		err := c.EnsureID()
		if err != nil {
			lg.Fatal(err)
		}
		lg.V(15).Infof("connection    id: %s", c.ID)
		if lg.V(50) {
			v, _ := c.Encode()
			lg.Infof("connection encoded: %s", v)
			lg.Infof("connection    full: %+v", c)
		}
		s.Connections[i] = c
	}
	return s, nil
}

func readSettings(c *Config) error {
	lg.V(5).Info("Reading settings file")

	isRead := false

	_, err := mkConfigDir()
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(ConfigPath("settings.json"))
	if err != nil {
		lg.Infof("Error loading settings.json %s", err)
	} else {
		settings, err := parseConfig(data)
		if err != nil {
			lg.Warningf("Config file error, deleting and resetting")
			err := os.Remove(ConfigPath("settings.json"))
			if err != nil {
				lg.Warningf("Could not delete old settingsfile (should probably panic here)")
			}
		} else {
			currentConfig.Settings = *settings
			isRead = true
		}
	}

	if !isRead {
		settings, err := parseConfig([]byte(settingsTemplate))
		if err != nil {
			panic("invalid defaultsettings")
		}
		currentConfig.Settings = *settings
	}

	transports := make(map[string]shared.Transport, 0)
	if currentConfig.Settings.Transports != nil {
		for _, v := range currentConfig.Settings.Transports {
			transports[v.Name] = v
		}
	}

	for _, v := range []shared.Transport{
		{Name: "obfs3", Bundled: true, TorPT: true},
		{Name: "obfs4", Bundled: true, TorPT: true},
		{Name: "shadowsocks-client", Bundled: true},
	} {
		transports[v.Name] = v
	}
	currentConfig.Settings.Transports = transports
	return nil
}

// ConfigPath returns a location for a file or path settings directory
func ConfigPath(file ...string) string {
	var paths []string
	paths = append(paths, osutil.HomePath())

	if runtime.GOOS == "windows" {
		paths = append(paths, "alkasir")
	} else {
		paths = append(paths, ".alkasir")
	}
	if len(file) > 0 {
		paths = append(paths, file...)
	}
	return filepath.Join(paths...)
}

// Settings is the in memory representation of the settings file which usually
// is loaded/saved from disk.
type Settings struct {
	Version     int // settings version
	LastID      int // last (week numbr % 3 ) + 1 an id counter was sent.
	Local       localSettings
	Connections []shared.Connection
	Transports  map[string]shared.Transport
}

type localSettings struct {
	ClientBindAddr      string // Address of the client api/web ui
	ClientAuthKey       string // Key for authenticating browser extension to client sessions
	CountryCode         string // Defaults to __ which has the special meaning that the user has not done the setting.
	Language            string // Defaults to en under testing.
	ClientAutoUpdate    bool
	BlocklistAutoUpdate bool
	CentralAddr         string // The base address for alakasir central server
}

// UserSetup returns true if the user has made the basic application setup.
func (l *localSettings) UserSetup() bool {
	return l.CountryCode != "__"
}

// ApplicationUpdateDuration returns ApplicationUpdateInterval as a duration.
func (l *localSettings) ApplicationUpdateDuration() time.Duration {
	if !l.ClientAutoUpdate {
		return time.Duration(0)
	}
	return time.Duration(time.Hour * 2)
}

// BlocklistUpdateDuration returns BlocklistUpdateInterval as a duration.
func (l *localSettings) BlocklistUpdateDuration() time.Duration {
	if !l.BlocklistAutoUpdate {
		return time.Duration(0)
	}
	return time.Duration(time.Hour * 6)
}

// settingsTemplate is default when there is no config file or if it's broken.
const settingsTemplate = `
{
  "Local": {
    "ClientBindAddr": "127.0.0.1:8899",
    "CountryCode": "__",
    "Language": "en",
    "ClientAutoUpdate": true,
    "BlocklistAutoUpdate": true
  }
}
`
