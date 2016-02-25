package service_test

// todo: implement this

import (
	"os"
	"os/exec"
	"testing"

	"github.com/alkasir/alkasir/pkg/service"
	"github.com/alkasir/alkasir/pkg/service/server"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/thomasf/lg"
)

var (
	testTransports = map[string]shared.Transport{
		"socks5": {
			Name:    "socks5",
			Bundled: true,
		},
	}
	testConnections = []shared.Connection{}
	transportName   = "socks5"
	failToStart     = false
	transportOption = server.NewOption("transport")
	testSettings    = `
[local]
ControllerBindAddr = "127.0.0.1:8899"
HttpProxyBindAddr =  "localhost:9124"
SocksProxyBindAddr = "localhost:9125"
Language = "en"
[[connection]]
transport = "socks5"

[transport.socks5]
bundled = true
`
)

// MockTransport to support testing.
type MockTransport struct {
	Name      string
	Transport string
	BindAddr  string
}

// MockTransport server.Check function.
func MockTransportCheck(serviceName *server.Option) (server.Parser, error) {
	if !serviceName.Is("transport") {
		return nil, server.ServiceNotFound()
	}
	if !transportOption.Is(transportName) {
		return nil, server.TransportNotFound()
	}
	st := MockTransport{
		Name:      serviceName.Get(),
		Transport: transportOption.Get(),
	}
	return st, nil
}

func (m MockTransport) Parse() (starter server.Starter, err error) {
	handler := server.NewHandler(m.Name)
	handler.PrintVersion()
	bindAddr := server.NewOption("bindaddr")
	m.BindAddr = bindAddr.Get()
	return m, nil
}

func (m MockTransport) Start() error {
	handler := server.NewHandler(m.Name)
	if failToStart {
		return handler.PrintError("no fun")
	} else {
		handler.PrintExpose("socks5", "127.0.0.1:1")
		handler.PrintDone()
	}
	return nil
}

func TestServerFail(t *testing.T) {
	if os.Getenv("__TEST_SUBCMD") == "1" {
		server.AddChecker(MockTransportCheck)
		failToStart = true
		if !server.CheckServiceEnv() {
			lg.Infoln("error no support")
		}
		server.RunService()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestServerFail")
	env := append(os.Environ(), "__TEST_SUBCMD=1")
	cmd.Env = env
	// cmd.Stdout = os.Stdout
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestServerSuccess(t *testing.T) {
	if os.Getenv("__TEST_SUBCMD") == "1" {
		failToStart = false
		server.AddChecker(MockTransportCheck)
		if !server.CheckServiceEnv() {
			lg.Infoln("error no support")
		}
		server.RunService()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestServerSuccess")
	env := append(os.Environ(), "__TEST_SUBCMD=1")
	env = append(env, "ALKASIR_SERVICE=transport")
	env = append(env, "ALKASIR_TRANSPORT=socks5")
	cmd.Env = env
	// cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}

func TestServiceClientSuccess(t *testing.T) {
	if os.Getenv("ALKASIR_SERVICE") == "transport" {
		failToStart = false
		server.AddChecker(MockTransportCheck)
		if !server.CheckServiceEnv() {
			lg.Infoln("error no support")
		}
		server.RunService()
		return
	}
	service.UpdateConnections(testConnections)
	service.UpdateTransports(testTransports)
	service.Arg = "-test.run=TestServiceClientSuccess"

	connection := shared.Connection{
		Transport: "socks5",
	}
	connection1Proxy, err := service.NewTransportService(connection)
	if err != nil {
		t.Fatalf("could not start %v", err)
	}
	err = connection1Proxy.Start()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}

func TestServiceClientFail(t *testing.T) {
	if os.Getenv("ALKASIR_SERVICE") == "transport" {
		failToStart = false
		server.AddChecker(MockTransportCheck)
		if !server.CheckServiceEnv() {
			lg.Infoln("error no support")
		}
		server.RunService()
		return
	}
	service.UpdateConnections(testConnections)
	service.UpdateTransports(testTransports)
	service.Arg = "-test.run=TestServiceClientFail"

	connection := shared.Connection{
		Transport: "socks5",
	}
	connection1Proxy, err := service.NewTransportService(connection)
	if err != nil {
		t.Fatalf("could not start %v", err)
	}

	err = connection1Proxy.Start()
	if err != nil {
		panic(err)
	}

	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}
