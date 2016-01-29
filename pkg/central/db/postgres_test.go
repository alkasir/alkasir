// +build databases

package db

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/influxdb/influxdb/client"
	"github.com/thomasf/lg"
)

var testclient *client.Client
var sqlDB *DB
var pgConnString = flag.String("pgconn", "user=alkasir_central password=alkasir_central dbname=alkasir_central port=39558 sslmode=disable", "postgresql connection string")

// InitDB opens a connection to the database.
func InitDB() error {
	var err error
	sqlDB, err = Open(*pgConnString)
	if err != nil {
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		sqlDB = nil
		return err
	}
	sqlDB.SetMaxIdleConns(100)
	lg.Infoln("Successfully connected to the database")
	return nil
}

func TestMain(m *testing.M) {
	err := InitDB()
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestFeature(t *testing.T) {

	res, err := sqlDB.RecentSuggestionSessions(100)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(res)

}
