package socks5

import (
	"net"

	"github.com/armon/go-socks5"
	"github.com/alkasir/alkasir/pkg/service/server"
)

type Socks5Transport struct {
	Name      string
	Transport string
	BindAddr  string
}

const (
	transportName = "socks5"
)

var transportOption = server.NewOption("transport")

func Check(serviceName *server.Option) (server.Parser, error) {
	if !serviceName.Is("transport") {
		return nil, server.ServiceNotFound()
	}
	if !transportOption.Is(transportName) {
		return nil, server.TransportNotFound()
	}
	st := Socks5Transport{
		Name:      serviceName.Get(),
		Transport: transportOption.Get(),
	}
	return st, nil
}

func (s Socks5Transport) Parse() (starter server.Starter, err error) {
	handler := server.NewHandler(s.Name)
	err = handler.PrintVersion()
	if err != nil {
		return nil, err
	}
	// handler := server.NewHandler(s.Name)
	bindAddr := server.NewOption("bindaddr")
	s.BindAddr = bindAddr.Get()
	return s, nil
}

func (s Socks5Transport) Start() error {
	handler := server.NewHandler(s.Name)

	socksConf := &socks5.Config{}
	server, err := socks5.New(socksConf)
	if err != nil {
		_ = handler.PrintError(err.Error())
		return err
	}

	listener, err := net.Listen("tcp", s.BindAddr)
	if err != nil {
		_ = handler.PrintError(err.Error())
		return err
	}
	listener = handler.MonitorListener(listener)
	defer func() {
		err := listener.Close()
		if err != nil {
			handler.Logln(err)
		}

	}()

	handler.PrintExpose("socks5", listener.Addr().String())
	handler.PrintDone()

	serve := func(l net.Listener) error {
		for {
			conn, err := l.Accept()
			if err != nil {
				return err
			}
			go func(conn net.Conn) {
				conn = handler.MonitorConn(conn)
				defer handler.TrackCloseConn(conn.RemoteAddr().String())
				handler.TrackOpenConn(conn.RemoteAddr().String())
				err := server.ServeConn(conn)
				if err != nil {
					handler.Logln(err)
				}

			}(conn)
		}
	}

	go func() {
		err = serve(listener)
		if err != nil {
			handler.Fatal(err)
		}
	}()
	handler.Wait()
	return nil
}
