package ptc


import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Config struct {
	Command          string `json:"command"`
	StateLocation    string `json:"statelocation"`
	ExitOnStdinClose bool   `json:"exitonstdinclose"`
	env              []string
	cmd              *exec.Cmd
	done             sync.WaitGroup
	methods          []MethodState
	started          bool
	stdin            io.WriteCloser
}

// // A combination of a method name and an address, as extracted from
type MethodState struct {
	Name    string
	Addr    *net.TCPAddr
	Options Args
}

// Wait waits until the transport process has shut down.
func (c *Config) Wait() error {
	if !c.started {
		return errors.New("not running")
	}
	c.done.Wait()
	return nil
}

// Kill terminates the transport process
func (c *Config) Kill() error {
	_ = c.stdin.Close()
	return c.cmd.Process.Kill()
}

// Methods returns all currently started methods
func (c *Config) Methods() []MethodState {
	return c.methods
}

// Method returns one method of requested name, error if no method was found.
func (c *Config) Method(methodName string) (MethodState, error) {
	methods := c.methods
	for _, m := range methods {
		if methodName == m.Name {
			return m, nil
		}
	}
	return MethodState{}, fmt.Errorf("Method %s not found", methodName)
}

// clenv returns a clean version os os.Environ().
//
// os.Env() is copied, then all varables matching ALKASIR_* are removed.
// This is the only Alkasir specific code in this package.
func cleanEnv() []string {
	osenv := os.Environ()
	var env []string
	for _, v := range osenv {
		if v != "ALKASIR_DEBUG" && !strings.HasPrefix(v, "ALKASIR_") {
			env = append(env, v)
		}
	}
	return env
}

func (c *Config) start() error {
	if c.started {
		return errors.New("already started")
	}
	env := cleanEnv()
	env = append(env, c.env...)
	if c.ExitOnStdinClose {
		env = append(env, "TOR_PT_EXIT_ON_STDIN_CLOSE=1")
	}

	cmd := exec.Command(c.Command)
	cmd.Env = env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	c.stdin = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdoutc := make(chan string)
	go func() {
		defer close(stdoutc)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if scanner.Err() != nil {
				log.Println(err)
				break
			}
			line := scanner.Text()
			stdoutc <- line
		}
	}()

	stderrc := make(chan string)
	go func() {
		defer close(stderrc)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if scanner.Err() != nil {
				log.Println(err)
				return
			}
			line := scanner.Text()
			stderrc <- line
		}
	}()

	go func() {
		for line := range stderrc {
			log.Println("stderr:", line)
		}
	}()
	parseerr := make(chan error, 0)
	parsed := make(chan []MethodState, 0)
	defer close(parsed)
	defer close(parseerr)

	go func() {
		methods, err := parseStdout(stdoutc)
		if err != nil {
			parseerr <- err
			return
		}
		parsed <- methods
		for line := range stdoutc {
			log.Println("stdout:", line)
		}
	}()

	c.done.Add(1)
	go func() {
		defer c.done.Done()
		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}
	}()

	select {
	case err := <-parseerr:
		return err
	case methods := <-parsed:
		c.methods = methods
	}

	c.cmd = cmd
	c.started = true
	log.Println("success start")
	return nil
}

func parseStdout(linec chan string) ([]MethodState, error) {
	kwds := []string{"VERSION", "ENV-ERROR", "VERSION-ERROR", "SMETHOD",
		"SMETHOD-ERROR", "SMETHODS", "CMETHODS", "CMETHOD", "CMETHOD-ERROR", "PROXY", "PROXY_DONE"}

	kwdmap := make(map[string]interface{}, 0)
	for _, v := range kwds {
		kwdmap[v] = true
	}
	var methods []MethodState
	var (
		methodsDone = false
		versionDone = false
	)
loop:
	for line := range linec {
		log.Printf("Received from sub-transport: %q.", line)

		fields := strings.Fields(strings.TrimRight(line, "\n"))
		if len(fields) < 1 {
			continue loop
		}
		kw := fields[0]
		args := fields[1:]

		if _, ok := kwdmap[kw]; !ok {
			return methods, fmt.Errorf("unknown keyword: %s in %v", kw, line)
		}

		if kw == "ENV-ERROR" || kw == "VERSION-ERROR" {
			return methods, fmt.Errorf("Error: %v", line)
		}

		if !versionDone {
			if kw != "VERSION" {
				return methods, fmt.Errorf("expected VERSION as first line: %v", line)
			}
			if len(args) < 1 {
				return methods, fmt.Errorf("expected VERSION arg: %v", line)
			}

			versionDone = true
			log.Println("version done")
			continue loop
		}

		if !methodsDone {
			if (kw == "SMETHODS" || kw == "CMETHODS") && len(args) == 1 && args[0] == "DONE" {
				methodsDone = true
				log.Println("methods done")
				break loop
			}
			if kw == "SMETHOD" {
				if len(args) < 2 {
					return methods, fmt.Errorf("expected SMETHOD args: %v", line)
				}

				addr, err := net.ResolveTCPAddr("", args[1])
				if err != nil {
					return methods, err
				}

				m := MethodState{
					Name: args[0],
					Addr: addr,
				}
				methods = append(methods, m)
				log.Println("registered smethod", m)
				continue loop
			} else if kw == "CMETHOD" {
				if len(args) < 3 {
					return methods, fmt.Errorf("expected CMETHOD args: %v", line)
				}

				addr, err := net.ResolveTCPAddr("", args[2])
				if err != nil {
					return methods, err
				}

				m := MethodState{
					Name: args[0],
					Addr: addr,
				}
				methods = append(methods, m)
				log.Println("registered cmethod", m)
				continue loop
			} else if kw == "SMETHOD-ERROR" || kw == "CMETHOD-ERROR" {
				log.Printf("warning: method failed, %s", line)
			}

		}
	}
	if len(methods) < 1 {
		return nil, errors.New("No requested methods found")
	}

	log.Println("returning", methods)
	return methods, nil
}
