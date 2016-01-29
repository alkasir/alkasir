package internet

import (
	"os"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type MockPool struct{}

func (m MockPool) Get() redis.Conn {
	c := redigomock.NewConn()
	return c
}

var pool = MockPool{}

func TestParseDump(t *testing.T) {
	file, err := os.Open("testdata/bview.20150101.sample.txt")
	if err != nil {
		panic(err)
	}
	conn := redigomock.NewConn()
	b := BGPDump{Date: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)}
	n, err := b.parseBGPCSV(file, conn)
	if err != nil {
		panic(err)
	}
	if n != 801 {
		t.Fatalf("expected 801 imported entries, got %d", n)
	}
}

func TestParseBrokenDump(t *testing.T) {
	file, err := os.Open("testdata/bview.20150102.sample.invalid-file.txt")
	if err != nil {
		panic(err)
	}
	conn := redigomock.NewConn()
	b := BGPDump{Date: time.Date(2015, 1, 2, 0, 0, 0, 0, time.UTC)}
	n, err := b.parseBGPCSV(file, conn)
	if err == nil {
		t.Fatalf("expected parse error")
	}
	if n != 5 {
		t.Fatalf("expected parse error on line 5 but %d lines were read", n)
	}
	if _, ok := err.(ParseError); !ok {
		t.Fatalf("expected parse error, got %s", err)
	}
}
