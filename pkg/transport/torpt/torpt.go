package torpt

import (
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/alkasir/alkasir/pkg/service/server"

	"github.com/alkasir/ptc"
)

type TorPTTransport struct {
	Name       string
	Transport  string
	RemoteAddr string
	Args       ptc.Args
}

var transportOption = server.NewOption("transport")

var transportNames = map[string]bool{
	"obfs4": true,
	"obfs3": true,
}

var (
	configdir = "."
)

func SetConfigDir(path string) {
	configdir = path
}

func Check(serviceName *server.Option) (server.Parser, error) {
	if !serviceName.Is("transport") {
		return nil, server.ServiceNotFound()
	}
	transportName := transportOption.Get()
	if !transportNames[transportName] {
		return nil, server.TransportNotFound()
	}

	st := TorPTTransport{
		Name: transportName,
	}
	return st, nil
}

func (s TorPTTransport) Parse() (starter server.Starter, err error) {

	handler := server.NewHandler(s.Name)

	remoteAddr := server.NewOption("remoteaddr")
	err = remoteAddr.Required()
	if err != nil {
		handler.PrintError(err.Error())
		return
	}
	s.RemoteAddr = remoteAddr.Get()

	secret := server.NewOption("secret")
	err = secret.Required()
	if err != nil {
		handler.PrintError(err.Error())
		return
	}

	err = s.Args.UnmarshalJSON([]byte(secret.Get()))
	if err != nil {
		handler.PrintError(err.Error())
		return
	}
	handler.PrintVersion()
	return s, nil

}

func (s TorPTTransport) Start() error {
	handler := server.NewHandler(s.Name)

	client := ptc.Client{
		Config: ptc.Config{
			ExitOnStdinClose: true,
			StateLocation:    configdir,
			Command:          os.Args[0],
		},
		Methods: []string{s.Name},
	}

	err := client.Start()
	if err != nil {
		_ = handler.PrintError(err.Error())
		return err
	}
	ptmethod, err := client.Method(s.Name)
	if err != nil {
		handler.PrintError(err.Error())
		return err
	}

	dialconf := ptc.PTDialMethodConfig{
		Args:       s.Args,
		ClientAddr: ptmethod.Addr.String(),
		ServerAddr: s.RemoteAddr,
	}

	dialer, err := dialconf.ClientDialer()
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		log.Fatalf("Failed to setup listener: %v", err)
	}

	handler.PrintExpose("socks5", listener.Addr().String())
	handler.PrintDone()

	forward := func(conn net.Conn) {
		conn = handler.MonitorConn(conn)
		client, err := dialer.Dial("tcp", s.RemoteAddr)
		if err != nil {
			log.Println(err)
		}

		client = handler.MonitorConn(client)
		defer client.Close()
		defer conn.Close()
		if err != nil {
			log.Fatalf("Dial failed: %v", err)
		}
		// log.Printf("Connected to localhost %v\n", conn)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := io.Copy(client, conn)
			if err != nil {
				log.Println("error", err)
			}

		}()
		go func() {
			defer wg.Done()
			_, err := io.Copy(conn, client)
			if err != nil {
				log.Println("error", err)
			}
		}()
		wg.Wait()
	}

	listener = handler.MonitorListener(listener)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf("ERROR: failed to accept listener: %v", err)
			}
			// log.Printf("Accepted connection %v\n", conn)

			go forward(conn)
		}
	}()
	handler.Wait()
	err = client.Kill()
	if err != nil {
		log.Println(err)
	}
	return nil
}
