// Package debugexport contains functions for exporting client state to
// investigate user problems.
package debugexport

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/mdp/sodiumbox"
	"github.com/thomasf/lg"
)

// PublicKey is set during build time for automatically encrypted debug logs
var PublicKey string

// DebugResponse can be sent to Alkasir for debugging purposes
type DebugResponse struct {
	Header       DebugHeader `json:"header,omitempty"`
	Config       interface{} `json:"config,omitempty"`
	Log          []string    `json:"log,omitempty"`
	Heap         []string    `json:"heap,omitempty"`
	GoRoutines   []string    `json:"goroutines,omitempty"`
	Block        []string    `json:"block,omitempty"`
	ThreadCreate []string    `json:"thread_create,omitempty"`
	Encrypted    string      `json:"encrypted,omitempty"`
}

// DebugHeader contains very general build and runtime information
type DebugHeader struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	Cmd       string    `json:"cmd"`
	GoVersion string    `json:"go_ver"`
}

// NewDebugResponse creates a filled DebugResponse struct
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

	if PublicKey != "" {
		response.Encrypt(PublicKey)
	} else {
		lg.Warningln("PublicKey not set, exporting unencrypted debug log")
	}

	return response

}

// Encrypt encrypts the response using the provided hex encoded public key
func (d *DebugResponse) Encrypt(pubKey string) error {
	if pubKey == "" {
		return fmt.Errorf("pubKey must be supplied")
	}

	pk := new([32]byte)
	dpk, err := hex.DecodeString(pubKey)
	if err != nil {
		lg.Fatalln("Could not decode debug public key")
	}
	copy(pk[:], dpk[:32])
	data, err := json.Marshal(&d)
	if err != nil {
		lg.Errorln("could not marshal debug response", err)
		return nil
	}

	encrypted, err := sodiumbox.Seal(data, pk)
	if err != nil {
		lg.Errorln("could not encrypt debug response", err)
		return nil
	}

	*d = DebugResponse{
		Header:    d.Header,
		Encrypted: hex.EncodeToString(encrypted.Box),
	}

	return nil

}

// Decrypt decrypts the response using the provided hex encoded public and secure keys.
func (d *DebugResponse) Decrypt(pubKey, secKey string) error {
	if pubKey == "" {
		return fmt.Errorf("public key must be supplied")

	}

	if secKey == "" {
		return fmt.Errorf("secret key must be supplied")

	}

	if d.Encrypted == "" {
		return fmt.Errorf("encrypted field is empty")
	}

	pk := new([32]byte)
	dpk, err := hex.DecodeString(pubKey)
	if err != nil {
		return fmt.Errorf("Could not decode debug public key")
	}
	if len(dpk) != 32 {
		return fmt.Errorf("Invalid public key length")
	}
	copy(pk[:], dpk[:32])

	sk := new([32]byte)
	dsk, err := hex.DecodeString(secKey)
	if err != nil {
		return fmt.Errorf("Could not decode debug secret key")
	}
	if len(dsk) != 32 {
		return fmt.Errorf("Invalid public key length")
	}
	copy(sk[:], dsk[:32])

	data, err := hex.DecodeString(d.Encrypted)
	if err != nil {
		return fmt.Errorf("could not decode encrypted info: %v", err)
	}

	decrypted, err := sodiumbox.SealOpen(data, pk, sk)
	if err != nil {
		return fmt.Errorf("could not decrypt debug info: %v", err)
	}

	var dr DebugResponse
	err = json.Unmarshal(decrypted.Content, &dr)
	if err != nil {
		return fmt.Errorf("could not decode decrypted content: %v", err.Error())
	}
	*d = dr

	return nil
}

// WriteToDisk writes the contets of the DebugResponse as individual files in a directory structure.
func (d *DebugResponse) WriteToDisk() error {
	if d.Encrypted != "" {
		return fmt.Errorf("will not write encrypted file to disk")
	}
	dir := d.filename()
	if err := os.MkdirAll(dir, 0775); err != nil {
		panic(err)
	}

	writeTextFile := func(data []string, basename string) error {
		filename := filepath.Join(dir, basename+".txt")
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0660)
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

	writeJSONFile := func(data interface{}, basename string) error {
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
		writeJSONFile(d.Header, "header"),
		writeTextFile(d.Log, "log"),
		writeTextFile(d.Heap, "heap"),
		writeTextFile(d.GoRoutines, "goroutines"),
		writeTextFile(d.Block, "block"),
		writeTextFile(d.ThreadCreate, "threadcreate"),
		writeJSONFile(d.Config, "config"),
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
