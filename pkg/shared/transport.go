package shared

// Transport holds the basic runtime configuration for a transport service.
type Transport struct {
	Name    string
	Bundled bool // bundled in distribution binary
	Command string
	TorPT   bool // if the transport is an tor pluggable transport
}

// TransportTraffic is sent from a transport to the client so that the client can know that the transport is active.
type TransportTraffic struct {
	Opened     []string `json:"opened"`     // Target of all currently open connections
	ReadTotal  uint64   `json:"readTotal"`  // bytes
	WriteTotal uint64   `json:"writeTotal"` // bytes
	Throughput float64  `json:"throughput"` // bytes/second

}
