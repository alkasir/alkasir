// Service client api
//
// The service manager is responsiblities are knowing (by name) which services
// exists, configure them using environment variables, launch them, read status
// by standard out and then
package service

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/thomasf/lg"
)

// Service is the central data structure for service client and manager.
type Service struct {
	ID           string              // instance id
	Name         string              // service name
	running      bool                // should proably be replaced by a function
	cmd          *exec.Cmd           // service process
	Command      string              // command name to launch serice
	Request      map[string](string) // initial input
	Response     map[string](string) // initial output
	Methods      *Methods            // methods exposed in Response
	authSecret   string              // authentication token for service callback
	stdout       io.ReadCloser       // stdout pipe
	stderr       io.ReadCloser       // stderr pipe
	stdin        io.WriteCloser      // stdin pipe
	quit         chan bool           // quit request channel
	removeOnStop bool                // Remove the service from ManagedServices upon Stop()
	waiter       sync.WaitGroup      // is released after service shutdown

	isCopy bool // set and managed by copy method
}

func (s *Service) copy() *Service {
	return &Service{
		ID:           s.ID,
		Name:         s.Name,
		running:      s.running,
		Command:      s.Command,
		Request:      s.Request,
		Response:     s.Response,
		Methods:      s.Methods,
		authSecret:   s.authSecret,
		quit:         s.quit,
		removeOnStop: s.removeOnStop,
		// ---
		isCopy: true,
	}
}

// Create a new service instance.
// note: name is not yet a decided requirement.
func NewService(name string) (s *Service) {

	s = &Service{
		ID:       serviceIdGen.New(),
		Name:     name,
		Request:  make(map[string]string),
		Response: make(map[string]string),
		Methods: &Methods{
			list: make([]*Method, 0),
		},
	}
	err := ManagedServices.add(s)
	if err != nil {
		lg.Fatal(err)
	}
	return
}

// register an exposed method in initial service protocl
func (s *Service) registerMethod(name, protocol, bindAddr string) {
	m := &Method{
		ID:       methodIdGen.New(),
		Name:     name,
		Service:  s,
		BindAddr: bindAddr,
		Protocol: protocol,
	}
	s.Methods.add(m)
}

// Returns true if the service currently is running.
func (s *Service) Running() bool {
	return s.running
}

// Stop the service
func (s *Service) Stop() {
	if s.running {
		s.quit <- true
	}

}

// Wait blocks until the underlying process is stopped
func (s *Service) wait() {
	if s.isCopy {
		lg.Fatal("wait called on copy of service!")
	}
	if s.cmd != nil {
		lg.V(10).Infof("Waiting for process %s to exit", s.ID)
		err := s.cmd.Wait()
		if err != nil {
			lg.Warningln(err)
		}
		lg.V(10).Infof("%s exited", s.ID)
	}
	s.waiter.Wait()

}

// Remove the service from the list of manages services, if the service is
// running it will be stopped first.
func (s *Service) Remove() error {
	if s.Running() {
		s.removeOnStop = true
		s.Stop()
	} else {
		return ManagedServices.remove(s)
	}
	return nil
}

// Set a variable in the request context.
func (s *Service) SetVar(key, value string) {
	if s.running {
		panic(errors.New("Service already running, cannot set configuration variable"))
	}
	s.Request[key] = value
}

// Sets up the service commands environment.
//
// os.Env() is copied, then all varables matching ALKASIR_* are removed.
// Finally the variables from Request are merged in as new ALKASIR_ variables.
func (s *Service) initEnv() {
	osenv := os.Environ()
	var env []string
	for _, v := range osenv {
		if v != "ALKASIR_DEBUG" && !strings.HasPrefix(v, "ALKASIR_") {
			env = append(env, v)
		}
	}
	for key, value := range s.Request {
		env = append(env, "ALKASIR_"+strings.ToUpper(key)+"="+value)
	}
	env = append(env, "ALKASIR_SAUTH="+s.authSecret)
	env = append(env, "ALKASIR_SADDR="+"http://localhost:8899/api/transports/traffic/")

	s.cmd.Env = env
}

func checkError(err error) {
	if err != nil {
		lg.Fatalf("Error: %s", err)
	}
}

// Arg is currently only used for testing to specify which test is run.
var Arg = ""

// regexps for service init protocol parsing. TODO: proper parsing
var (
	versionM = regexp.MustCompile(`^VERSION (.*)$`)                   // INITIAL LINE, PROTOCOL VERSION
	errorM   = regexp.MustCompile(`^ERROR.*$`)                        // ERROR abort!
	parentM  = regexp.MustCompile(`^PARENT ([^ ]+) ([^ ]+) ([^ ]+)$`) // PARENT method parent-protocol parent-connection
	exposeM  = regexp.MustCompile(`^EXPOSE ([^ ]+) ([^ ]+) ([^ ]+)$`) // EXPOSE method protocol connection
	doneM    = regexp.MustCompile(`^DONE$`)                           // SERVICE STARTED, PROTOCOL FINISHED
)

// Initialize service
func (s *Service) initService() error {
	cmd := s.cmd
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	s.stdout = stdout
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	s.stderr = stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	s.stdin = stdin

	if lg.V(5) {
		alkasirEnv := ""
		for _, v := range s.cmd.Env {
			if strings.HasPrefix(v, "ALKASIR_") {
				alkasirEnv += v + " "
			}
		}
		lg.Infof("Starting service: %s %s", alkasirEnv, cmd.Path)
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	var transportErrMsg string
	transportErr := false
	var line string
	for scanner.Scan() {
		line = scanner.Text()

		lg.V(5).Infoln("DBG: ", line)

		if errorM.MatchString(line) {
			transportErr = true
			transportErrMsg = line
			return errors.New("error: " + transportErrMsg)
		} else if doneM.MatchString(line) {
			break
		} else if exposeM.MatchString(line) {
			match := exposeM.FindStringSubmatch(line)
			s.Response["bindaddr"] = match[3]
			s.Response["protocol"] = match[2]
			s.registerMethod(match[1], match[2], match[3])
		} else if versionM.MatchString(line) {
		} else if parentM.MatchString(line) {
			match := parentM.FindStringSubmatch(line)
			s.Response["parentaddr"] = match[3]
		} else {
			lg.Infoln("not handled line:", line)
			return errors.New("unhandeled line")
		}
	}
	if transportErr {
		err := cmd.Wait()
		if err != nil {
			lg.Warningln(err)
		}
		lg.Fatal(transportErrMsg)
		return errors.New("transport err")
	}
	return err
}

// initDone means handing off the service process output to it's own goroutine.
func (s *Service) initDone() error {

	lg.V(5).Infof("s.request: %+v", s.Request)
	lg.V(5).Infof("s.response: %+v", s.Response)
	s.waiter.Add(1)
	s.running = true
	go func() {
		defer func() {
			s.running = false
		}()

		go func() {
			scanner := bufio.NewScanner(s.stdout)
			for scanner.Scan() {
				if scanner.Err() != nil {
					lg.Warningln("service stdout", scanner.Err())
					break
				}
				line := scanner.Text()
				lg.Infoln(s.ID, "stdout:", line)
			}
			lg.V(20).Infof("service outch closed %s", s.ID)
		}()

		go func() {
			scanner := bufio.NewScanner(s.stderr)
			for scanner.Scan() {
				if scanner.Err() != nil {
					lg.Warningln("service stderr", scanner.Err())
					break
				}
				line := scanner.Text()
				lg.Infoln(s.ID, "stderr:", line)
			}
			lg.V(20).Infof("service stderr closed %s", s.ID)
		}()

		s.quit = make(chan bool)
		defer close(s.quit)

		select {
		case <-s.quit:
			lg.V(6).Infof("stopping service %s", s.ID)

			if err := s.stdin.Close(); err != nil {
				lg.Errorf("could not close stdin for %s: %s", s.ID, err.Error())
			}

			if err := s.stdout.Close(); err != nil {
				lg.Errorf("could not close stdout for %s: %s", s.ID, err.Error())
			}

			if err := s.stderr.Close(); err != nil {
				lg.Errorf("could not close stderr for %s: %s", s.ID, err.Error())
			}

			lg.V(10).Infof("Killing process service %s", s.ID)

		}

		lg.V(10).Infof("stopped service %s", s.ID)
		if s.removeOnStop {
			ManagedServices.remove(s)
		}
		s.waiter.Done()
	}()
	return nil
}

// Start the service.
// Currently, this function does validation as well which
// might need to be reconsidered.
func (s *Service) Start() (err error) {
	s.cmd = exec.Command(s.Command, Arg)
	s.initEnv()
	err = s.initService()
	if err != nil {
		return
	}
	err = s.initDone()
	if err != nil {
		return
	}
	return
}

// Method holds the runtime configuration of one of
type Method struct {
	ID       string   // instance id
	Name     string   // method name
	Service  *Service // the owning service
	BindAddr string   // exposed network address
	Protocol string   // exposed exposed protocol
}

// Methods manages a list of methods
type Methods struct {
	list []*Method
}

// Add an exposed method to the list
func (m *Methods) add(method *Method) {
	m.list = append(m.list, method)
}

// All returns all medthods added to current list
func (m *Methods) All() []*Method {
	return m.list
}

// TransportService adds some features on top of the plain services.
type TransportService struct {
	*Service                      // the related service
	connection *shared.Connection // connection details
}

// SetBindAddr sets the net address which  the transport service should bind to locally.
func (s *TransportService) SetBindaddr(bindaddr string) error {
	if s.Running() {
		return errors.New("Service already running, cannot set bindaddr")
	}
	s.SetVar("bindaddr", bindaddr)
	return nil
}

// set the net address that the transport service should bind to locally
func (s *TransportService) SetVerbose() error {
	if s.Running() {
		return errors.New("Service already running, cannot set bindaddr")
	}
	s.SetVar("verbose", "yes")
	return nil
}

// let the transport
func (s *TransportService) SetParent(parent *TransportService) error {
	if s.Running() {
		return errors.New("Service already running, cannot set parent")
	}
	if !parent.Running() {
		return errors.New("Parent must be running before starting service")
	}
	parentAddr := parent.Response["bindaddr"]
	s.SetVar("parentaddr", parentAddr)
	return nil
}

var currentTransports = make(map[string]shared.Transport)

func UpdateTransports(transports map[string]shared.Transport) {
	currentTransports = transports
}

// create new transport service from connection details data
func NewTransportService(connection shared.Connection) (transportService *TransportService, err error) {
	transport := currentTransports[connection.Transport]

	var command string
	if transport.Bundled {
		command = os.Args[0]
	} else if transport.Command != "" {
		command = transport.Command
	} else {
		return nil, errors.New("no command registered for service")
	}

	service := NewService("transport")
	transportService = &TransportService{
		Service:    service,
		connection: &connection,
	}
	s := transportService.Service
	s.Command = command

	s.SetVar("service", "transport")
	s.SetVar("transport", connection.Transport)
	s.SetVar("remoteaddr", connection.Addr)
	s.SetVar("secret", connection.Secret)

	return
}
