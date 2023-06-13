package lokinet

// #include <lokinet.h>
import "C"
import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/net/proxy"
	"io"
	"net"
	"os"
	"sync"
	"unsafe"
)

type logHolder struct {
	writer io.Writer
	buffer bytes.Buffer
	access sync.Mutex
}

// EmbeddedContext is an embedded lokinet client that can do stuff over the network.
type EmbeddedContext struct {
	// void pointer of the underlying lokinet embedded context
	ctx unsafe.Pointer
	// we write lokinet's logs to here and then it gets flushed to be printed.
	log logHolder
	// guards access during runtime
	access sync.Mutex
	// make setup idempotent
	setup sync.Once
	// makes teardown idempotent
	teardown sync.Once
}

// ErrDeadlock is returned when it underlying managed lokinet context has an internal data race based deadlock.
var ErrDeadlock = errors.New("lokinet is deadlocked")

// Setup will Idempotently setup internal state of the managed lokient context.
func (ctx EmbeddedContext) Setup(opts network.Opts) (err error) {
	ctx.setup.Do(func() {
		// acquire this so that any other go routine will wait for us to finish setup if we haven't already.
		ctx.access.Lock()
		defer ctx.access.Unlock()
		// setup logger
		ctx.log.writer = opts.LogWriter
		C.lokinet_set_syncing_logger(logWrite, logSync, unsafe.Pointer(&ctx.log))
		// make lokinet context
		ptr := C.lokinet_context_new()
		// start it up and wait for it to work
		ret := C.lokinet_context_start(ptr)
		if ret != 0 {
			err = fmt.Errorf("failed to start lokinet: %s", C.GoString(C.strerror(ret)))
			return
		}
		l.ctx = unsafe.Pointer(ptr)
		for st := C.lokinet_wait_for_ready(100, l.ctx); st != 0; {
			if st == -2 {
				err = ErrDeadlock
				return
			}
		}
	})
	return
}

// Close implements io.Closer
func (l *EmbeddedContext) Close() error {

	l.teardown.Do(func() {
		l.access.Lock()
		defer l.access.Unlock()
		if ep.ctx == nil {
			return
		}
		ctx := ep.ctx
		ep.ctx = nil
		C.lokinet_context_free(ctx)
	})
	return nil
}

// Addr is a string representation of a local or remote endpoint inside lokinet.
type Addr = string

func (ctx *EmbeddedContext) NewPeristentLocalEndpoint(keyfile string) *NetworkInterface {
}

func (ctx *EmbeddedContext) NewEphemeralLocalEndpoint() *NetworkInterface {

}

func logWrite(msg *C.Char, ptr uintptr) {
	// queues logs into our log buffer so go can print logs.
	log := (*logHolder)(unsafe.Pointer(ptr))
	log.access.Lock()
	defer log.access.Unlock()

	if log.writer == nil {
		// dont queue anything if we have no log writer
		return
	}
	err := io.WriteString(log.buffer, C.GoString(msg))
	// TODO: do we want to actually do that?
	if err != nil {
		panic(err)
	}
}

func logSync(ptr uintptr) {
	// writes all the logs into our io.Writer
	log := (*logHolder)(unsafe.Pointer(ptr))
	log.access.Lock()
	defer log.access.Unlock()
	if log.write != nil {
		io.Copy(log.writer, &log.buffer)
	}
	// we now will reset our internal log buffer
	log.buffer.Reset()
}

//export consumePktIO
func consumePktIO(place uintptr, pkt_ptr *C.uint8, pkt_sz uint16) {
	// this function will consume an ip packet coming from lokinet itself
}

//export completedPktIO
func completedPktIO(place uintptr) {

}
