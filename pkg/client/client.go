// Alkasir client
package client

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	_ "net/http/pprof" // imported for registring pprof
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/thomasf/lg"
	clientconfig "github.com/alkasir/alkasir/pkg/client/internal/config"
	"github.com/alkasir/alkasir/pkg/client/ui"
	"github.com/alkasir/alkasir/pkg/pac"
	"github.com/alkasir/alkasir/pkg/service"
	"github.com/alkasir/alkasir/pkg/shared"
)

var (
	wipeData          bool
	debugPort         int
	saveChromeExt     bool
	debugEnabled      bool
	hotEnabled        bool
	upgradeEnabled    bool
	clientAuthKeyFlag string
	centralAddrFlag   string
	bindAddrFlag      string

	upgradeDiffsBaseURL string // this value is overridden on release builds, also the full url will be provided by the server.
)

func init() {
	flag.BoolVar(&wipeData, "wipedata", false, "Wipe all user data")
	flag.BoolVar(&saveChromeExt, "saveChromeExt", false, "Export chrome extension to disk on startup.")
	flag.BoolVar(&debugEnabled, "debug", false, "Enable debugging (development feature)")
	flag.IntVar(&debugPort, "debugPort", 6067, "Port to expose alkasir debug on (developer feaute)")
	flag.BoolVar(&hotEnabled, "hot", false, "Enable hot reloading (development feature)")
	flag.BoolVar(&upgradeEnabled, "upgrade", true, "Enable client binary upgrades")
	flag.StringVar(&clientAuthKeyFlag, "authKey", "", "Override generated client<->browser authentication key")
	flag.StringVar(&centralAddrFlag, "centralAddr", "", "Override the URL to where the central server is expected to exist")
	flag.StringVar(&bindAddrFlag, "bindAddr", "", "Override the configured client bindAddr")
}

var (
	uiRunning sync.WaitGroup
)

// Init does precondition check if the application can/should be started.
// Init will return an error message with reason for exit printed.
func Run() {
	if debugEnabled {
		log.Println("ALKASIR_DEBUG ENABLED!")
		err := os.Setenv("ALKASIR_DEBUG", "1")
		if err != nil {
			log.Fatal(err)
		}
	}
	if hotEnabled {
		log.Println("ALKASIR_HOT ENABLED!")
		err := os.Setenv("ALKASIR_HOT", "1")
		if err != nil {
			log.Fatal(err)
		}
	}

	// the darwin systray does not exit the main loop
	if runtime.GOOS != "darwin" {
		uiRunning.Add(1)
	}

	err := ui.Run(func() {
		Atexit(ui.Done)

		// start the getpublic ip updater.
		go func() {
			_ = shared.GetPublicIPAddr()
		}()

		if debugEnabled {
			go func() {
				err := http.ListenAndServe(
					fmt.Sprintf("localhost:%d", debugPort), nil)
				if err != nil {
					panic(err)
				}
			}()
		}

		// wipe user data
		if wipeData {
			settingsdir := clientconfig.ConfigPath()
			if settingsdir == "" {
				log.Println("[wipe] Configdir not set")
				os.Exit(1)
			}
			settingsfile := clientconfig.ConfigPath("settings.json")
			if _, err := os.Stat(settingsfile); os.IsNotExist(err) {
				log.Println("[wipe] No settings.json in configdir, will NOT wipe data")
				os.Exit(1)
			}
			log.Println("Wiping all user data")
			if err := os.RemoveAll(settingsdir); err != nil {
				log.Println(err)
			}
		}

		// Prepare logging
		logdir := clientconfig.ConfigPath("log")
		err := os.MkdirAll(logdir, 0775)
		if err != nil {
			log.Println("Could not create logging directory")
			os.Exit(1)
		}
		err = flag.Set("log_dir", logdir)
		if err != nil {
			panic(err)
		}
		lg.SetSrcHighlight("alkasir/cmd", "alkasir/pkg")
		lg.CopyStandardLogTo("INFO")

		// Start init
		if VERSION != "" {
			lg.Infoln("Alkasir v" + VERSION)
		} else {
			lg.Warningln("Alkasir dev version (VERSION not set)")
		}

		lg.V(1).Info("Log v-level:", lg.Verbosity())
		_, err = clientconfig.Read()
		if err != nil {
			lg.Infoln("Could not read config")
			exit()
		}
		lg.V(30).Infoln("settings", clientconfig.Get().Settings)

		if saveChromeExt {
			err := saveChromeExtension()
			if err != nil {
				lg.Fatal(err)
			}
		}

		{
			configChanged, err := clientconfig.UpgradeConfig()
			if err != nil {
				lg.Fatalln("Could not upgrade config", err)

			}
			clientconfig.Update(func(conf *clientconfig.Config) error {

				if clientAuthKeyFlag != "" {
					lg.Warningln("Overriding generated authKey with", clientAuthKeyFlag)
					conf.Settings.Local.ClientAuthKey = clientAuthKeyFlag
					configChanged = true
				}

				if bindAddrFlag != "" {
					lg.Warningln("Overriding configured bindAddr with", bindAddrFlag)
					conf.Settings.Local.ClientBindAddr = bindAddrFlag
					configChanged = true
				}

				if centralAddrFlag != "" {
					lg.Warningln("Overriding central server addr with", centralAddrFlag)
					conf.Settings.Local.CentralAddr = centralAddrFlag
					configChanged = true
				}

				return nil
			})

			if configChanged {
				if err := clientconfig.Write(); err != nil {
					lg.Warning(err)
				}
			}

		}
		conf := clientconfig.Get()
		loadTranslations(LanguageOptions...)
		if err := ui.Language(conf.Settings.Local.Language); err != nil {
			lg.Warningln(err)
		}

		go func() {
			select {
			case <-sigIntC:
				exit()
			case <-ui.Actions.Quit:
				exit()
			}
		}()

		for _, e := range []error{
			mime.AddExtensionType(".json", "application/json"),
			mime.AddExtensionType(".js", "application/javascript"),
			mime.AddExtensionType(".css", "text/css"),
			mime.AddExtensionType(".md", "text/plain"),
		} {
			if e != nil {
				lg.Warning(e)
			}
		}

		err = startInternalHTTPServer(conf.Settings.Local.ClientAuthKey)
		if err != nil {
			lg.Fatal("could not start internal http services")
		}

		// Connect the default transport
		service.UpdateConnections(conf.Settings.Connections)
		service.UpdateTransports(conf.Settings.Transports)
		go service.StartConnectionManager(conf.Settings.Local.ClientAuthKey)

		// TODO: async
		pac.UpdateDirectList(conf.DirectHosts.Hosts)
		pac.UpdateBlockedList(conf.BlockedHostsCentral.Hosts, conf.BlockedHosts.Hosts)
		lastBlocklistChange = time.Now()

		go StartBlocklistUpgrader()
		if upgradeDiffsBaseURL != "" {
			lg.V(19).Infoln("upgradeDiffsBaseURL is ", upgradeDiffsBaseURL)
			go StartBinaryUpgradeChecker(upgradeDiffsBaseURL)
		} else {
			lg.Warningln("empty upgradeDiffsBaseURL, disabling upgrade checks")
		}

		lg.V(5).Info("Alkasir has started")

	})
	// the darwin systray does not exit the main loop
	if runtime.GOOS != "darwin" {
		uiRunning.Done()
	}

	lg.Infoln("ui.Run ended")
	if err != nil {
		log.Println("client.Run error:", err)
	}
}

var (
	atexitFuncs []func()
	atexitMu    sync.Mutex
)

// Atexit adds a func to be run when the application is exiting.
func Atexit(f func()) {
	atexitMu.Lock()
	defer atexitMu.Unlock()
	atexitFuncs = append(atexitFuncs, f)
}

// AtexitKillCmd takes care of killing a command on application exit.
//
// TODO: currently this does not clean up references to dead processes, it just
// adds forever.
func AtexitKillCmd(cmd *exec.Cmd) {
	Atexit(func() {
		lg.V(10).Info("Atexit kill ", cmd.Path, cmd.Args)
		err := cmd.Process.Kill()
		if err != nil {
			lg.V(5).Info("kill failed:", err)
		}
		// TODO: possible deadlock?
		if err := cmd.Wait(); err != nil {
			lg.Warningln(err)
		}
	})
}

func exit() {
	lg.Infoln("alkasir is shutting down")
	atexitMu.Lock() // this lock should be kept, one shutdown should be enough for everyone.
	lg.Flush()
	if err := clientconfig.Write(); err != nil {
		lg.Errorf("could not save config file: %s", err.Error())
	}
	lg.V(9).Infoln("running atexit funcs")
	for _, f := range atexitFuncs {
		funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		lg.V(10).Infoln("Running at exit func", funcName)
		f()
		lg.V(10).Infoln("Finished at exit func", funcName)
	}
	atexitFuncs = atexitFuncs[:0]
	lg.V(9).Infoln("atexit funcs done")
	lg.V(9).Infoln("stopping connectionmanager")
	service.StopConnectionManager()
	lg.V(9).Infoln("stopping services")
	service.StopAll()
	lg.V(9).Infoln("services stopped")
	lg.V(9).Infoln("waiting for UI shutdown to finish")
	uiRunning.Wait()
	lg.V(9).Infoln("ui shut down")

	lg.Flush()
	lg.Infoln("alkasir shutdown complete")
	lg.Flush()
	time.Sleep(time.Millisecond * 1)
	os.Exit(0)
}
