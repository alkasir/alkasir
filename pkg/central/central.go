package central

import (
	"crypto/rand"
	"encoding/base64"
	"expvar"
	"flag" // register expvar in default servermux
	"net/http"
	_ "net/http/pprof" // register pprof
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/garyburd/redigo/redis"
	maxminddb "github.com/oschwald/maxminddb-golang"
	"github.com/thomasf/internet"
	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/central/db"
	"github.com/alkasir/alkasir/pkg/shared"
)

var (
	validCountryCodesMu sync.RWMutex
	validCountryCodes   map[string]bool
	redisPool           *redis.Pool
	redisServer         = flag.String("redisServer", ":39550", "")
	redisPassword       = flag.String("redisPassword", "", "")
	apiBindAddr         = flag.String("APIAddr", ":8080", "port do bind api server to")
	exportApiBindAddr   = flag.String("exportAPIAddr", ":8082", "port to bind export api server to")
	exportApiSecretKey  = flag.String("exportAPISecretKey", "", "Secret key for export api auth")
	monitorBindAddr     = flag.String("monitorAddr", "localhost:8081", "port to bind monitor server to")
	datadirFlag         = flag.String("datadir", "", "directory to store data")
	datadir             string
)

var mmCountryDB, mmCityDB *maxminddb.Reader

// Init initializes the server.
func Init() error {
	emptyRev := shared.InitialRevision()
	lg.SetSrcHighlight("alkasir/cmd", "alkasir/pkg")
	lg.CopyStandardLogTo("INFO")
	lg.V(1).Info("Log v-level:", lg.Verbosity())
	lg.V(1).Info("Active country codes:", shared.CountryCodes)
	lg.V(1).Info("Empty hash is:", emptyRev.Hash)
	lg.Flush()

	if *datadirFlag == "" {
		u, err := user.Current()
		if err != nil {
			lg.Fatal(err)
		}
		datadir = filepath.Join(u.HomeDir, ".alkasir-central")
	} else {
		datadir = *datadirFlag
	}

	validCountryCodes = make(map[string]bool, len(shared.CountryCodes))
	validCountryCodesMu.Lock()
	for _, cc := range shared.CountryCodes {
		validCountryCodes[cc] = true
	}
	validCountryCodesMu.Unlock()

	err := InitDB()
	if err != nil {
		lg.Fatalln(err)
		return err
	}
	redisPool = newRedisPool(*redisServer, *redisPassword)

	internet.SetDataDir(filepath.Join(datadir, "internet"))

	countryFile := filepath.Join(datadir, "internet", "GeoLite2-Country.mmdb")
	if _, err := os.Stat(countryFile); os.IsNotExist(err) {
		// http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz
		lg.Fatalf("cannot enable IP2CountryCode lookups, %s is missing", countryFile)
	} else {
		var err error
		mmCountryDB, err = maxminddb.Open(countryFile)
		if err != nil {
			lg.Fatal(err)
		}
	}

	cityFile := filepath.Join(datadir, "internet", "GeoLite2-City.mmdb")
	if _, err := os.Stat(cityFile); os.IsNotExist(err) {
		// http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz
		lg.Warningf("cannot enable IP2CityGeoNameID lookups, %s is missing", cityFile)
	} else {
		mmCityDB, err = maxminddb.Open(cityFile)
		if err != nil {
			lg.Fatal(err)
		}
		// defer mmCityDB.Close()
	}

	return nil
}

// Run runs the initialized server.
func Run() {
	var wg sync.WaitGroup

	// start expvar monitor server
	go startMonitoring(*monitorBindAddr)

	// start the getpublic ip updater.
	go func() {
		_ = shared.GetPublicIPAddr()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lg.V(2).Infoln("Loading recent sessions from postgres...")
		recents, err := sqlDB.RecentSuggestionSessions(20000)
		if err != nil {
			lg.Fatal(err)
		}
		db.SessionTokens.Reset(recents)
		lg.V(2).Infof("Loaded %d sessions from postgres...", len(recents))
		lg.Flush()

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		conn := redisPool.Get()
		defer conn.Close()
		lg.V(2).Infoln("BGPDump refresh started...")
		n, err := internet.RefreshBGPDump(conn)
		lg.V(2).Infof("BGPDump refresh ended, %d items added.", n)
		lg.Flush()
		if err != nil {
			if *offline {
				lg.Infoln("offline", err)
			} else {
				lg.Fatal(err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		conn := redisPool.Get()
		defer conn.Close()
		lg.V(2).Infoln("CIDRReport refresh started...")
		n, err := internet.RefreshCIDRReport(conn)
		lg.V(2).Infof("CIDRReport refresh ended, %d items added", n)
		if err != nil {
			if *offline {
				lg.Infoln("offline", err)
			} else {
				lg.Fatal(err)
			}
		}
	}()
	wg.Wait()

	// start signal handling
	wg.Add(1)
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT)
		lg.Infoln(<-ch)
		wg.Done()
	}()

	internetClient := db.NewInternetClient(redisPool)
	maxmindClient := db.NewMaxmindClient(mmCountryDB, mmCityDB)

	clients := db.Clients{
		DB:       sqlDB,
		Internet: internetClient,
		Maxmind:  maxmindClient,
	}

	// start http json api server
	go func(addr string, dba db.Clients) {
		mux, err := apiMux(dba)
		lg.Info("Starting http server", addr)
		err = http.ListenAndServe(addr, mux)
		if err != nil {
			lg.Fatal(err)
		}
	}(*apiBindAddr, clients)

	// start http export api server
	go func(addr string, dba db.Clients) {
		if *exportApiSecretKey == "" {
			lg.Warningln("exportApiSecretKey flag/env not set, will not start export api server")
			b := make([]byte, 32)
			_, err := rand.Read(b)
			if err != nil {
				lg.Fatalf("random generator not functioning...")
				return
			}
			suggestedkey := base64.StdEncoding.EncodeToString(b)
			lg.Infoln("suggested export key:", suggestedkey)
			return
		}
		key, err := base64.StdEncoding.DecodeString(*exportApiSecretKey)
		if err != nil {
			lg.Fatalf("could not decode export api secret key: %s", *exportApiSecretKey)
		}

		mux, err := apiMuxExport(dba, key)
		lg.Info("Starting export api server", addr)
		err = http.ListenAndServe(addr, mux)
		if err != nil {
			lg.Fatal(err)
		}
	}(*exportApiBindAddr, clients)

	startAnalysis(clients)
	startMeasurer(clients)

	wg.Wait()
}

var (
	sqlDB                   *db.DB
	offline                 = flag.Bool("offline", false, "don't require an internet connection (dev mode)")
	pgConnString            = flag.String("pgconn", "user=alkasir_central password=alkasir_central dbname=alkasir_central port=39558 sslmode=disable", "postgresql connection string")
	pgMaxIdleConnections    = flag.Int("pg_max_idle_conn", 15, "Maximum idle postgres connections")
	pgMaxOpenConnections    = flag.Int("pg_max_open_conn", 100, "Maximum active postgres connections")
	redisMaxIdleConnections = flag.Int("redis_max_idle_conn", 50, "Maximum idle redis connections")
	redisMaxConnections     = flag.Int("redis_max_open_conn", 10000, "Maximum active redis connections")
)

// InitDB opens a connection to the database.
func InitDB() error {
	var err error
	sqlDB, err = db.Open(*pgConnString)
	if err != nil {
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		sqlDB = nil
		return err
	}
	sqlDB.SetMaxIdleConns(*pgMaxIdleConnections)
	sqlDB.SetMaxOpenConns(*pgMaxOpenConnections)
	lg.Infoln("Successfully connected to the database")
	return nil
}

var startTime = time.Now().UTC()

func goroutines() interface{} {
	return runtime.NumGoroutine()
}

// uptime is an expvar.Func compliant wrapper for uptime info.
func uptime() interface{} {
	uptime := time.Since(startTime)
	return int64(uptime)
}

func startMonitoring(addr string) {

	expvar.Publish("Goroutines", expvar.Func(goroutines))
	expvar.Publish("Uptime", expvar.Func(uptime))

	redisActiveConn := expvar.NewInt("redis_pool_conn_active")
	redisMaxConn := expvar.NewInt("redis_pool_conn_max")
	redisMaxConn.Set(int64(redisPool.MaxActive))
	go func() {
		tick := time.NewTicker(time.Duration(1 * time.Second))
		for range tick.C {
			if redisPool == nil {
				redisActiveConn.Set(0)
			} else {
				redisActiveConn.Set(int64(redisPool.ActiveCount()))
			}
		}
	}()
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		lg.Fatal(err)
	}
}

func newRedisPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     *redisMaxIdleConnections,
		MaxActive:   *redisMaxConnections,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}

			// if _, err := c.Do("AUTH", password); err != nil {
			// c.Close()
			// return nil, err
			// }
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

}
