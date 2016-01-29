package debugexport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/shared"
)

type DebugResponse struct {
	Config       interface{} `json:"config"` // maybe add back later
	Header       DebugHeader `json:"header"`
	Log          []string    `json:"log"`
	Heap         []string    `json:"heap"`
	GoRoutines   []string    `json:"goroutines"`
	Block        []string    `json:"block"`
	ThreadCreate []string    `json:"thread_create"`
}

type DebugHeader struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	Cmd       string    `json:"cmd"`
	GoVersion string    `json:"go_ver"`
}

func (d *DebugResponse) filename(paths ...[]string) string {
	var newpath []string
	newpath = append(newpath,
		"alkasir-debug-reports",

		fmt.Sprintf("%s-%s-%s-%s-%s",
			d.Header.CreatedAt.Format("2006-01-02--15-04"),
			d.Header.OS,
			d.Header.Arch,
			d.Header.ID,
			d.Header.Version,
		),
	)
	return filepath.Join(newpath...)

}

func (d *DebugResponse) WriteToDisk() error {
	dir := d.filename()
	if err := os.MkdirAll(dir, 0775); err != nil {
		panic(err)
	}

	writeTextFile := func(data []string, basename string) error {
		filename := filepath.Join(dir, basename+".txt")
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0665)
		if err != nil {
			return err
		}
		defer f.Close()
		for _, v := range data {
			_, err := f.WriteString(v + "\n")
			if err != nil {
				return err
			}
		}
		return nil
	}

	writeJsonFile := func(data interface{}, basename string) error {
		bytes, err := json.MarshalIndent(&data, "", "  ")
		if err != nil {
			panic(err)
		}
		filename := filepath.Join(dir, basename+".json")
		err = ioutil.WriteFile(filename, bytes, 0665)
		if err != nil {
			return err
		}
		return nil
	}

	failed := false

	for _, v := range []error{
		writeJsonFile(d.Header, "header"),
		writeTextFile(d.Log, "log"),
		writeTextFile(d.Heap, "heap"),
		writeTextFile(d.GoRoutines, "goroutines"),
		writeTextFile(d.Block, "block"),
		writeTextFile(d.ThreadCreate, "threadcreate"),
		writeJsonFile(d.Config, "config"),
	} {
		if v != nil {
			failed = true
			lg.Error(v)
		}
	}

	if failed {
		return fmt.Errorf("errors writing out report %s", dir)
	}
	lg.Infof("wrote report for %s", dir)

	return nil
}

func NewDebugResposne(version string, config interface{}) *DebugResponse {
	ID, err := shared.SecureRandomString(12)
	if err != nil {
		panic("could not generate random number")
	}
	response := &DebugResponse{
		Header: DebugHeader{
			Cmd:       filepath.Base(os.Args[0]),
			ID:        ID,
			Version:   version,
			CreatedAt: time.Now(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			GoVersion: runtime.Version(),
		},
		Config: config,
	}

	getProfile := func(name string) []string {
		buf := bytes.NewBuffer(nil)
		err := pprof.Lookup(name).WriteTo(buf, 2)
		if err != nil {
			lg.Errorln(err)
		} else {
			return strings.Split(
				buf.String(),
				"\n")
		}
		return []string{}
	}

	response.Heap = getProfile("heap")
	response.GoRoutines = getProfile("goroutine")
	response.ThreadCreate = getProfile("threadcreate")
	response.Block = getProfile("block")
	// memlog should be last so that it can catch errors up to the point of
	// collection.
	response.Log = lg.Memlog()

	return response

}
