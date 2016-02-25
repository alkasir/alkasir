package shadowsocks

import (
	"errors"
	"net"
	"strconv"

	"github.com/alkasir/alkasir/pkg/service/server"
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
)

type ShadowsocksTransport struct {
	Name       string
	Transport  string
	BindPort   int
	BindAddr   string
	RemoteAddr string
	RemoteHost string
	RemotePort int
	Secret     string
	Verbose    bool
}

const (
	transportName = "shadowsocks-client"
)

var transportOption = server.NewOption("transport")

func Check(serviceName *server.Option) (server.Parser, error) {
	if !serviceName.Is("transport") {
		return nil, server.ServiceNotFound()
	}
	if !transportOption.Is(transportName) {
		return nil, server.TransportNotFound()
	}
	st := ShadowsocksTransport{
		Name:      serviceName.Get(),
		Transport: transportOption.Get(),
	}
	return st, nil
}

func (s ShadowsocksTransport) Parse() (starter server.Starter, err error) {

	handler := server.NewHandler(s.Name)
	handler.PrintVersion()

	bindAddr := server.NewOption("bindaddr")
	bindAddr.Default("127.0.0.1:0")

	_, bindPortStr, err := net.SplitHostPort(bindAddr.BindAddr())
	if err != nil {
		handler.PrintError(err.Error())
		return
	}
	s.BindAddr = bindAddr.Get()
	bindPort, err := strconv.Atoi(bindPortStr)
	if err != nil {
		handler.PrintError(err.Error())
		return
	}
	s.BindPort = bindPort

	remoteAddr := server.NewOption("remoteaddr")
	err = remoteAddr.Required()
	if err != nil {
		handler.PrintError(err.Error())
		return
	}
	s.RemoteAddr = remoteAddr.Get()

	remoteHost, remotePortStr, err := net.SplitHostPort(remoteAddr.Get())
	if err != nil {
		handler.PrintError(err.Error())
		return
	}
	s.RemoteHost = remoteHost

	remotePort, err := strconv.Atoi(remotePortStr)
	if err != nil {
		handler.PrintError(err.Error())
		return
	}
	s.RemotePort = remotePort

	secret := server.NewOption("secret")
	err = secret.Required()
	if err != nil {
		handler.PrintError(err.Error())
		return
	}
	s.Secret = secret.Get()
	return s, nil
}

func (s ShadowsocksTransport) Start() error {
	handler := server.NewHandler(s.Name)

	listener, err := net.Listen("tcp", s.BindAddr)
	if err != nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		handler.PrintError("CANNOT BIND TO ANY PORT")
		return errors.New("Cannot bind to any port")
	}

	listener = handler.MonitorListener(listener)
	handler.PrintExpose("socks5", listener.Addr().String())

	config := &ss.Config{
		Server:     s.RemoteHost,
		ServerPort: s.RemotePort,
		Password:   s.Secret,
		LocalPort:  s.BindPort,
		Method:     "aes-256-cfb",
	}

	parseServerConfig(config)

	handler.PrintDone()

	verbose := server.NewOption("verbose")

	if verbose.Has() {
		s.Verbose = true
		ss.SetDebug(true)
		debug = true
	}

	go func() {
		for {
			defer listener.Close()
			conn, err := listener.Accept()
			if err != nil {
				debug.Println("accept:", err)
				continue
			}
			go func(conn net.Conn) {
				conn = handler.MonitorConn(conn)
				if debug {
					debug.Printf("socks connect from %s\n", conn.RemoteAddr().String())
				}
				closed := false
				defer func() {
					if !closed {
						conn.Close()
					}
				}()

				var err error = nil
				if err = handShake(conn); err != nil {
					debug.Printf("socks handshake: %s", err)
					return
				}
				rawaddr, addr, err := getRequest(conn)
				if err != nil {
					debug.Println("error getting request:", err)
					return
				}
				// Sending connection established message immediately to client.
				// This some round trip time for creating socks connection with the client.
				// But if connection failed, the client will get connection reset error.
				_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
				if err != nil {
					debug.Println("send connection confirmation:", err)
					return
				}

				remote, err := createServerConn(rawaddr, addr)
				if err != nil {
					if len(servers.srvCipher) > 1 {
						debug.Println("Failed connect to all avaiable shadowsocks server")
					}
					return
				}
				handler.TrackOpenConn(addr)
				defer func() {
					if !closed {
						remote.Close()
					}
				}()

				go ss.PipeThenClose(conn, remote)
				ss.PipeThenClose(remote, conn)
				closed = true
				handler.TrackCloseConn(addr)

			}(conn)
		}
	}()
	handler.Wait()
	return nil
}
