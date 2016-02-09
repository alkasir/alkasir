// Provides pluggble network transports
// this is the run-as part of the service specification..
package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getlantern/bytecounting"
	"github.com/alkasir/alkasir/pkg/shared"
)

// Check if service is supported
type Checker interface {
	Check(serviceName *Option) (Parser, error)
}

type Check func(serviceName *Option) (Parser, error)

// Parse the environment and emit errors if misconfiguration is detected
type Parser interface {
	Parse() (Starter, error)
}

// Start the service
type Starter interface {
	Start() error // should probably return a channel or similar
}

type Service struct {
	Name string
}

// Handler for a service method
type MethodHandler struct {
	method  string // transport method
	version string // api version
	nread   uint64 // number of bytes read
	nwrite  uint64 // number of bytes written

	amu   sync.RWMutex      // mutex belonging to addrs
	addrs map[string]uint64 // [addr]numOpenConnections

}

// Create a new Handler
func NewHandler(method string) *MethodHandler {
	mh := &MethodHandler{
		method:  method,
		version: "1.0",
		addrs:   make(map[string]uint64, 0),
	}
	return mh
}

//  TrackOpenedConn is to be called when the the transport is used to open a specific address
func (h *MethodHandler) TrackOpenConn(addr string) {
	h.amu.Lock()
	if _, ok := h.addrs[addr]; ok {
		h.addrs[addr] = h.addrs[addr] + 1
	} else {
		h.addrs[addr] = 1
	}
	h.amu.Unlock()
}

// TrackCloseConn is to be called after a connection opened and registered with TrackOpenConn is closed again.
func (h *MethodHandler) TrackCloseConn(addr string) {

	h.amu.Lock()
	if _, ok := h.addrs[addr]; ok {
		h.addrs[addr] = h.addrs[addr] - 1
	} else {
		log.Println("warning: should not arrive here", h.addrs[addr])
		h.addrs[addr] = 0
	}
	h.amu.Unlock()
}

// MonitorConn waps a net.Conn for transport usage statistics
func (h *MethodHandler) MonitorConn(conn net.Conn) net.Conn {
	c := bytecounting.Conn{
		Orig: conn,
		OnRead: func(bytes int64) {
			atomic.AddUint64(&h.nread, uint64(bytes))
		},
		OnWrite: func(bytes int64) {
			atomic.AddUint64(&h.nwrite, uint64(bytes))
		},
	}
	return &c
}

// MonitorListener wraps a net.Listener for transport usage statistics
func (h *MethodHandler) MonitorListener(listener net.Listener) net.Listener {
	l := bytecounting.Listener{
		Orig: listener,
		OnRead: func(bytes int64) {
			atomic.AddUint64(&h.nread, uint64(bytes))
		},
		OnWrite: func(bytes int64) {
			atomic.AddUint64(&h.nwrite, uint64(bytes))
		},
	}
	return &l
}

func (h *MethodHandler) Logln(values ...interface{}) {
	log.Println(values...)
}

// Wait waits until stdin is closed which is when the transport should terminate
func (h *MethodHandler) Wait() {
	_, err := io.Copy(ioutil.Discard, os.Stdin)
	if err != nil {
		log.Println(err)
	}
}

// Fatal logs an error and exits immediately
func (h *MethodHandler) Fatal(err error) {
	h.Logln("fatal error %s", err.Error())
	os.Exit(1)
}

// Handles returns true if it's Handler handles a service
func (h *MethodHandler) Handles() (handles bool) {
	service := Option{name: "service"}
	transport := Option{name: "transport"}
	if service.Get() == "transport" && transport.Get() == h.method {
		handles = true
	}
	return
}

// Start the server initiation
func (h *MethodHandler) PrintVersion() error {
	line("VERSION", h.version)
	err := NewOption("service").Required()
	if err != nil {
		return h.PrintError(err.Error())
	}
	return nil
}

// Echo DONE to client.
//
// This means initiation was successful.
func (h *MethodHandler) PrintDone() {
	line("DONE")
	go h.monitor()

}

// Echo ERROR to the client.
//
// This means the initiation process should abort.
// func (h *MethodHandler) PrintError(text string) error {
// 	return doError("ERROR", h.method, text)
// }
func (h *MethodHandler) PrintError(text string) error {
	return doError("ERROR", h.method, text)
}

// Echo PARENT to the client.
//
// This means the server has successfully configured
// itself for connecting thru a parent proxy.
func (h *MethodHandler) PrintParent(proto, addr string) {
	line("PARENT", h.method, proto, addr)

}

//Echo EXPOSE to the client.
//
// This means that the service has exposed
// a Method requested by the configuration.
func (h *MethodHandler) PrintExpose(protocol, addr string) {
	line("EXPOSE", h.method, protocol, addr)
}

// Option represents an environment variable and is used for parsing and
// validating the server input.
type Option struct {
	name         string
	callback     func() error
	defaultValue string
}

func NewOption(name string) *Option {
	return &Option{name: name}
}

// Returns the option's environment variable name.
func (o *Option) EnvName() string {
	return "ALKASIR_" + strings.ToUpper(o.name)
}

// Check if the current environment has this option set.
func (o *Option) Has() (exists bool) {
	return o.defaultValue != "" || os.Getenv(o.EnvName()) != ""
}

// Check if the current environment has this option set.
func (o *Option) Is(value string) (is bool) {
	if o.Has() && o.Get() == value {
		return true
	}
	return false
}

// Get the value or default value if a default is set.
func (o *Option) Get() (value string) {
	value = os.Getenv(o.EnvName())
	if value == "" {
		value = o.defaultValue
	}
	return
}

// Set default value to be returned by Option.Get() if there is no actual value
// in the environment variables.
func (o *Option) Default(value string) {
	o.defaultValue = value
}

// Returns an error if the the variable is not set in the environment.
func (o *Option) Required() (err error) {
	if os.Getenv(o.EnvName()) == "" {
		err = errors.New(fmt.Sprintf("required variable not set: %s", o.EnvName()))
	}
	return
}

// Return argument or replaced by an default netaddr if not set
func (o *Option) BindAddr() (addr string) {
	if o.Has() {
		addr = o.Get()
	} else {
		addr = "127.0.0.1:"
	}
	return
}

// ! CLEAN EVERYTHING BELOW UP !

// protocol
//
// - print version
// - env vars are read
// - verify ALKASIR_TRANSPORT=[name] is supported by current module
// - print back ROLE
//

func TransportNotFound() error {
	return getError("METHOD-ERROR", os.Getenv("ALKASIR_TRANSPORT"), "TRANSPORT NOT FOUND")
}

func ServiceNotFound() error {
	return getError("METHOD-ERROR", os.Getenv("ALKASIR_SERVICE"), "SERVICE NOT FOUND")
}

// Writer to which pluggable transports negotiation messages are written. It
// defaults to a Writer that writes to os.Stdout and calls Sync after each
// write.
var Stdout io.Writer = shared.SyncWriter{File: os.Stdout}

// Represents an error that can happen during negotiation, for example
// ENV-ERROR. When an error occurs, we print it to stdout and also pass it up
// the return chain.
type ptErr struct {
	Keyword string
	Args    []string
}

// Implements the error interface.
func (err *ptErr) Error() string {
	return formatline(err.Keyword, err.Args...)
}

var keywordIsValid = regexp.MustCompile("^[[:alnum:]_-]+$")

// Returns true if keyword contains only bytes allowed in a servie output line
func keywordIsSafe(keyword string) bool {
	return keywordIsValid.MatchString(keyword)
}

// Returns true iff arg contains only bytes allowed in a PT→Tor output line arg.
// <ArgChar> ::= <any US-ASCII character but NUL or NL>
func argIsSafe(arg string) bool {
	for _, b := range []byte(arg) {
		if b >= '\x80' || b == '\x00' || b == '\n' {
			return false
		}
	}
	return true
}

func formatline(keyword string, v ...string) string {
	var buf bytes.Buffer
	if !keywordIsSafe(keyword) {
		panic(fmt.Sprintf("keyword %q contains forbidden bytes", keyword))
	}
	buf.WriteString(keyword)
	for _, x := range v {
		if !argIsSafe(x) {
			panic(fmt.Sprintf("arg %q contains forbidden bytes", x))
		}
		buf.WriteString(" " + x)
	}
	return buf.String()
}

// Print a service protocol line to Stdout.
func line(keyword string, v ...string) {
	fmt.Fprintln(Stdout, formatline(keyword, v...))
}

// return the given error as a ptErr.
func getError(keyword string, v ...string) *ptErr {
	return &ptErr{keyword, v}
}

// Emit and return the given error as a ptErr.
func doError(keyword string, v ...string) *ptErr {
	line(keyword, v...)
	return &ptErr{keyword, v}
}

// TODO: this needs to be formalized into the transport api
func (h *MethodHandler) postActivity(form *shared.TransportTraffic) {
	authOpt := Option{name: "sauth"}
	addrOpt := Option{name: "saddr"}

	if !authOpt.Has() || !addrOpt.Has() {
		log.Printf("will not log, sauth or saddr are not set\n")
		return
	}

	body, err := json.Marshal(&form)

	poster, err := http.NewRequest("POST", addrOpt.Get(), nil)
	if err != nil {
		log.Println(err)
		return
	}

	poster.Header.Add("Content-Type", "application/json")
	poster.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authOpt.Get()))
	poster.Body = byteBufferCloser{bytes.NewBuffer(body)}

	resp, err := http.DefaultClient.Do(poster)

	if err != nil {
		log.Println(err)
		log.Printf("Could not post activity to %s", addrOpt.Get())
		return
	}
	defer resp.Body.Close()

}

func (h *MethodHandler) monitor() {
	start := time.Now()
	lasttime := start
	var lastval uint64
	for {
		<-time.After(time.Second)

		nread := atomic.LoadUint64(&h.nread)
		nwrite := atomic.LoadUint64(&h.nwrite)
		ntotal := nread + nwrite
		since := ntotal - lastval
		now := time.Now()
		dur := now.Sub(lasttime)
		rate := float64(since) / dur.Seconds()
		lasttime = now
		lastval = ntotal

		opened := []string{}
		h.amu.RLock()
		for k, v := range h.addrs {
			if v > 0 {
				opened = append(opened, k)
			}
		}
		h.amu.RUnlock()

		stat := shared.TransportTraffic{
			Opened:     opened,
			ReadTotal:  nread,
			WriteTotal: nwrite,
			Throughput: rate,
		}
		// log.Printf("%f kb/s", rate/1024.0)
		go h.postActivity(&stat)

	}
}

// byteBufferCloser
type byteBufferCloser struct {
	*bytes.Buffer
}

func (c byteBufferCloser) Close() error {
	return nil
}
