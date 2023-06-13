package network

import (
	"crypto/ed25519"
	"golang.org/x/net/proxy"
	"io"
	"net"
)

type Opts struct {
	/// LogWriter is an io.Writer which we write out logs of the internal state of the network to.
	LogWriter io.Writer
}

// Network provides a way to talk to the network
type Endpoint interface {
	io.Closer
	proxy.ContextDialer
	// Listen is an analog to net.Listen that only needs the port and listens on our endpoint only
	Listen(port string) (net.Listener, error)
	// ListenPacket is an anolog to net.ListenPacket that only needs the port on our endpoint only
	ListenPacket(port string) (net.PacketConn, error)
}

// Network provides a way to talk to the network
type Network interface {
	io.Closer

	// NewEndpoint will create a new network endpoint with private keys we provide.
	NewEndpoint(privkey ed25519.PrivateKey) *Endpoint

	// DefaultEndpoint will give us our main network we use that is already created.
	DefaultEndpoint() *Endpoint
}
