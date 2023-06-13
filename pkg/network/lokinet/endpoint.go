package lokinet

import (
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// how big ip packets pretend to be inside lokinet
const lokinetL3MTU = 1500

// a bullshit value for now
const lokinetOverhead = 128

// local ethernet address for our end
const l2LocalAddr = tcpip.LinkAddress("69:69:69:69:69:69")

// remote ethernet address for the other end
const l2RemoteAddr = tcpip.LinkAddress("42:42:42:42:42:42")

// NetworkInterface implements a stack.NetworkInterface that goes over lokinet with a distinct .loki address.
type NetworkInterface struct {
	// this channel is closed to indicate we should tear down
	quit chan struct{}
	// woken up when we got a frames from lokinet
	read chan struct{}
	// woken up when netstack wants to write frames to lokinet
	write chan stack.PacketBufferList
	// this channel is closed to indicate that we have completed teardown
	done chan struct{}
}

// MTU implements stack.NetworkInterface
func (*NetworkInterface) MTU() uint32 {
	return lokinetL3MTU
}

// MaxHeaderLength implements stack.NetworkInterface
func (*NetworkInterface) MaxHeaderLength() uint16 {
	return lokinetL3MTU + lokinetOverhead
}

// WritePackets implements stack.NetworkInterface
func (link *NetworkInterface) WritePackets(pkts stack.PacketBufferList) (n int, err tcpip.Error) {
	return
}

// teardown will initialize teardown of our loop
func (link *NetworkInterface) teardown() {
	link.done <- struct{}{}
}

// readone does one io cycle for reading frames from lokinet
func (link *NetworkInterface) readone() {

}

// readone does one io cycle for writing frames to lokinet
func (link *NetworkInterface) writeone() {
	link
}

// loop is our busy loop for io
func (link *NetworkInterface) loop() {
	for {
		select {
		case <-link.done:
			link.ctx.Close()
			return
		case <-link.read:
			link.readone()
		case <-link.write:
			link.writeone()
		}

	}
}

func (link *NetworkInterface) Wait() {
	if link.done != nil {
		<-link.done
		close(link.done)
		close(link.read)
		close(link.write)
		link.done = nil
	}
}

func (link *NetworkInterface) startup() {
	link.done = make(chan struct{})
	link.read = make(chan struct{})
	link.write = make(chan struct{})
	go link.loop()
}

// IsAttached implements stack.NetworkInterface
func (link *NetworkInterface) IsAttached() bool {
	return link.dispatcher != nil
}

// Attach implements stack.NetworkInterface
func (link *NetworkInterface) Attach(dispatcher stack.NetworkDispatcher) {
	link.dispatcher = dispatcher
	if link.IsAttached() {
		link.startup()
	} else {
		link.teardown()
	}
}

// DialContext implements proxy.DialContext
func (link *NetworkInterface) DialContext(ctx *context.Context, network, address string) (net.Conn, error) {

}
