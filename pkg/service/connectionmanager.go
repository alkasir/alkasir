package service

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/thomasf/lg"
	"h12.me/socks"

	"github.com/alkasir/alkasir/pkg/client/ui"
	"github.com/alkasir/alkasir/pkg/shared"
)

// ConnectionEvent describes the current state of a transport connection.
type ConnectionEvent struct {
	State      ConnectionState   // Current state
	Time       time.Time         // Timestamp of the event
	Connection shared.Connection // Associated saved connection
	ServiceID  string            // Service instance id
}

// ConnectionState describes the current known state of the default transport.
type ConnectionState int

//go:generate stringer -type=ConnectionState
// TODO: Split STATES and ACTIONS
const (
	Init ConnectionState = iota // Initial state
	// Service stages
	ServiceInit  //  Service initializing
	ServiceStart // The transport service is starting
	Test         // Waiting for respone from test during start up

	// everything is fine
	Up // Connected and tested

	// Problems
	Backoff // Backoff because of multiple failed attepmts at using this connection

	// Errors
	WrongProtocol // The expected protocol was not found
	Failed        // One of the previous steps failed
	NotConfigured // No transport has been set up
	TestFailed    // A network test failed
	Ended         // Nothing more can happen
)

func (c ConnectionEvent) newState(state ConnectionState) ConnectionEvent {
	switch c.State {
	case Ended:
		panic("Connection already ended")
	}
	event := ConnectionEvent{
		State:      state,
		Connection: c.Connection,
		Time:       time.Now(),
		ServiceID:  c.ServiceID,
	}
	connectionEvents <- event
	return event
}

// newConnectionEventhistory initializes a history for a transportconnection
func newConnectionEventhistory(connection shared.Connection) ConnectionEvent {
	event := ConnectionEvent{
		State:      Init,
		Time:       time.Now(),
		Connection: connection,
	}

	connectionEvents <- event
	return event
}

func AddListener(c chan ConnectionHistory) {
	addNetworkStateListener <- c
}

// Channel to listen for new event listeners
var addNetworkStateListener = make(chan chan ConnectionHistory)

// Channel for all events
var connectionEvents = make(chan ConnectionEvent)

type ConnectionHistory struct {
	History []ConnectionEvent // All previous connectionevents from this connection
}

// Current returns the current state from the connection history
func (c *ConnectionHistory) Current() ConnectionEvent {
	return c.History[len(c.History)-1]
}

// Down returns true if the connection is known to be not up.
func (c ConnectionHistory) IsUp() bool {
	switch c.Current().State {
	case Up:
		return true
	default:
		return false
	}
}

var stopCh = make(chan bool)
var reconnectCh = make(chan bool)
var connectionTestedCh = make(chan bool)

func StartConnectionManager(authKey string) {

	listeners := make([]chan ConnectionHistory, 0)

	// TODO: Test on irregular intervals
	reverifyTicker := time.NewTicker(10 * time.Second)

	// the key is Connection.UUID
	histories := make(map[string][]ConnectionEvent)
	currents := make(map[string]*ConnectionEvent)

	currentConnectionID := ""
	currentConnectionsMu.Lock()
	currentConnIdx := 0
	if len(currents) < 1 {
		go connect(currentConnections[currentConnIdx], authKey)
		currentConnectionID = currentConnections[currentConnIdx].ID
	}
	currentConnectionsMu.Unlock()

	firstUpNoProblems := true // no need to spam the user with popups

	var reconnectTimer *time.Timer
	var connectionTestRunning bool
loop:

	for {
	s:
		select {
		case <-stopCh:
			lg.Infoln("connection manager shut down")
			break loop

		case <-reconnectCh:
			currentConnectionsMu.Lock()
			if len(currentConnections) < 1 {
				currentConnectionsMu.Unlock()
				lg.Warningln("No connections enabled")
				reconnectTimer = time.AfterFunc(4*time.Second, func() {
					reconnectCh <- true
				})
				break s
			}
			currentConnIdx = (currentConnIdx + 1) % (len(currentConnections))
			c := currentConnections[currentConnIdx]
			lg.V(10).Infof("reconnecting to transport %v", c)
			go connect(currentConnections[currentConnIdx], authKey)
			currentConnectionID = c.ID
			currentConnectionsMu.Unlock()

		case listener := <-addNetworkStateListener:
			listeners = append(listeners, listener)

		case event := <-connectionEvents:
			if _, v := histories[event.Connection.ID]; !v {
				histories[event.Connection.ID] = make([]ConnectionEvent, 0)
			} else if len(histories[event.Connection.ID]) > 20 {
				lg.V(5).Infoln("trimming connection history")
				histories[event.Connection.ID] = histories[event.Connection.ID][:20]
			}
			histories[event.Connection.ID] = append(histories[event.Connection.ID], event)
			currents[event.Connection.ID] = &event
			emitEvent := ConnectionHistory{
				History: histories[event.Connection.ID],
			}

			if lg.V(3) {
				switch event.State {
				case Failed, TestFailed:
					lg.Warningln("event  ", event.Connection.ID, ": ", event.State)
				default:
					lg.Infoln("event  ", event.Connection.ID, ": ", event.State)
				}
			}
			switch event.State {
			case Up:
				if firstUpNoProblems {
					firstUpNoProblems = false
				} else {
					ui.Notify("transport_connected_message")
				}
			case Failed:
				firstUpNoProblems = false
				ui.Notify("transport_error_message")
			case TestFailed:
				firstUpNoProblems = false
				ui.Notify("transport_retry")
			case Ended:
				delete(currents, event.Connection.ID)
				lg.V(15).Infoln("waiting 4 seconds before sending reconnect")
				if reconnectTimer != nil {
					reconnectTimer.Stop()
				}
				reconnectTimer = time.AfterFunc(4*time.Second, func() {
					reconnectCh <- true
				})
			}
			lg.V(7).Infoln("Forwarding connection event to listeners", emitEvent.Current())
			for _, l := range listeners {
				l <- emitEvent
			}

		case <-reverifyTicker.C:
			if !connectionTestRunning {
				conn, ok := currents[currentConnectionID]
				if ok && conn.State == Up {
					connectionTestRunning = true
					go func() {
						err := testConn(conn)
						if err != nil {
							lg.Warningln(err)
							connectionTestedCh <- false
						}
						connectionTestedCh <- true
					}()
				}
			}

		case <-connectionTestedCh:
			connectionTestRunning = false
		}
	}
}

func StopConnectionManager() {
	stopCh <- true
}

func testSocks5Internet(addr string) (err error) {
	dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, addr)
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial:                  dialSocksProxy,
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: time.Duration(time.Second * 10),
		},
	}
	resp, err := httpClient.Get("https://alkasir.com/ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return
}

func testConn(event *ConnectionEvent) error {
	defaultTransportM.Lock()
	defer defaultTransportM.Unlock()
	if defaultTransport == nil {
		transportOk = false
		event.newState(TestFailed)
		event.newState(NotConfigured)
		event.newState(Ended)
		return errors.New("No active transport")
	}
	err := testSocks5Internet(defaultTransport.Service.Response["bindaddr"])
	if err != nil {
		transportOk = false
		event.newState(TestFailed)
		event.newState(Failed)
		event.newState(Ended)
		return errors.New("Http get test failed")
	} else {
		if event.State != Up && lg.V(4) {
			lg.Infof("event: tested %s -> %s (%s)", event.State, Up, event.ServiceID)
		}
		transportOk = true
		if event.State != Up {
			event.newState(Up)
		}
	}
	transportOk = true
	return nil
}

var (
	currentConnectionsMu sync.Mutex
	currentConnections   = make([]shared.Connection, 0)
)

// shuffle is a Fisher-Yates shuffle
func shuffleConnections(slc []shared.Connection) {
	N := len(slc)
	for i := 0; i < N; i++ {
		// choose index uniformly in [i, N-1]
		r := i + rand.Intn(N-i)
		slc[r], slc[i] = slc[i], slc[r]
	}
}

func UpdateConnections(connections []shared.Connection) {
	currentConnectionsMu.Lock()
	defer currentConnectionsMu.Unlock()
	currentConnections = make([]shared.Connection, 0)
	for _, v := range connections {
		if !v.Disabled {
			currentConnections = append(currentConnections, v)
		}
	}
	// Connections are shuffeled so that the internal connection order of
	// retires can be sequencial while a random connection is chosen from the
	// setting file.
	shuffleConnections(currentConnections)

}

var DefaultProxyBindAddr = "127.0.0.1:0"

func connect(connection shared.Connection, authKey string) {

	defaultTransportM.Lock()
	if defaultTransport != nil {
		err := defaultTransport.Remove()
		if err != nil {
			lg.Warningln(err)
		}
	}
	defaultTransportM.Unlock()

	event := newConnectionEventhistory(connection)

	event.newState(ServiceInit)
	ts, err := NewTransportService(connection)
	ts.authSecret = authKey
	if lg.V(6) {
		ts.SetVerbose()
	}
	ts.SetBindaddr(DefaultProxyBindAddr)

	if err != nil {
		event.newState(Failed)
		event.newState(Ended)
		return
	}
	event.newState(ServiceStart)
	event.ServiceID = ts.Service.ID
	err = ts.Start()
	if err != nil {
		event.newState(Failed)
		event.newState(Ended)
		return
	}
	response := ts.Service.Response
	if response["protocol"] != "socks5" {
		event.newState(WrongProtocol)
		event.newState(Failed)
	}

	event.newState(Test)
	go testConn(&event)

	defaultTransportM.Lock()
	defaultTransport = ts
	defaultTransportM.Unlock()

}

var (
	defaultTransportM sync.Mutex
	defaultTransport  *TransportService
	transportOk       = false
)

func TransportOk() bool {
	defaultTransportM.Lock()
	defer defaultTransportM.Unlock()
	return transportOk

}

// NewTransportHTTPClient returns a http client of the default transport
func NewTransportHTTPClient() (*http.Client, error) {
	defaultTransportM.Lock()
	defer defaultTransportM.Unlock()
	if defaultTransport == nil {
		return nil, errors.New("transport not connected)")
	}
	client := defaultTransport.HTTPClient()
	return client, nil
}

type Dialer func(network, addr string) (net.Conn, error)

// Dial returns a
func (t *TransportService) Dial() Dialer {
	response := t.Service.Response
	return socks.DialSocksProxy(socks.SOCKS5, response["bindaddr"])
}

func (t *TransportService) HTTPClient() *http.Client {
	dial := t.Dial()
	tr := &http.Transport{Dial: dial}
	httpClient := &http.Client{Transport: tr}
	return httpClient
}
