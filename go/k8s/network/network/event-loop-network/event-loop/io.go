package event_loop

import (
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

// Events represents the server events for the Serve call.
// Each event has an Action return value that is used manage the state
// of the connection and server.
type Events struct {
	// NumLoops sets the number of loops to use for the server. Setting this
	// to a value greater than 1 will effectively make the server
	// multithreaded for multi-core machines. Which means you must take care
	// with synchonizing memory between all event callbacks. Setting to 0 or 1
	// will run the server single-threaded. Setting to -1 will automatically
	// assign this value equal to runtime.NumProcs().
	NumLoops int
	// LoadBalance sets the load balancing method. Load balancing is always a
	// best effort to attempt to distribute the incoming connections between
	// multiple loops. This option is only works when NumLoops is set.
	LoadBalance LoadBalance
	// Serving fires when the server can accept connections. The server
	// parameter has information and various utilities.
	Serving func(server Server) (action Action)
	// Opened fires when a new connection has opened.
	// The info parameter has information about the connection such as
	// it's local and remote address.
	// Use the out return value to write data to the connection.
	// The opts return value is used to set connection options.
	Opened func(c Conn) (out []byte, opts Options, action Action)
	// Closed fires when a connection has closed.
	// The err parameter is the last known connection error.
	Closed func(c Conn, err error) (action Action)
	// Detached fires when a connection has been previously detached.
	// Once detached it's up to the receiver of this event to manage the
	// state of the connection. The Closed event will not be called for
	// this connection.
	// The conn parameter is a ReadWriteCloser that represents the
	// underlying socket connection. It can be freely used in goroutines
	// and should be closed when it's no longer needed.
	Detached func(c Conn, rwc io.ReadWriteCloser) (action Action)
	// PreWrite fires just before any data is written to any client socket.
	PreWrite func()
	// Data fires when a connection sends the server data.
	// The in parameter is the incoming data.
	// Use the out return value to write data to the connection.
	Data func(c Conn, in []byte) (out []byte, action Action)
	// Tick fires immediately after the server starts and will fire again
	// following the duration specified by the delay return value.
	Tick func() (delay time.Duration, action Action)
}

// Server represents a server context which provides information about the
// running server and has control functions for managing state.
type Server struct {
	// The addrs parameter is an array of listening addresses that align
	// with the addr strings passed to the Serve function.
	Addrs []net.Addr
	// NumLoops is the number of loops that the server is using.
	NumLoops int
}

// Action is an action that occurs after the completion of an event.
type Action int

// Conn is an evio connection.
type Conn interface {
	// Context returns a user-defined context.
	Context() interface{}
	// SetContext sets a user-defined context.
	SetContext(interface{})
	// AddrIndex is the index of server address that was passed to the Serve call.
	AddrIndex() int
	// LocalAddr is the connection's local socket address.
	LocalAddr() net.Addr
	// RemoteAddr is the connection's remote peer address.
	RemoteAddr() net.Addr
	// Wake triggers a Data event for this connection.
	Wake()
}

type listener struct {
	ln      net.Listener
	lnaddr  net.Addr
	pconn   net.PacketConn
	opts    addrOpts
	f       *os.File
	fd      int
	network string
	addr    string
}

func (ln *listener) close() {
	if ln.fd != 0 {
		syscall.Close(ln.fd)
	}
	if ln.f != nil {
		ln.f.Close()
	}
	if ln.ln != nil {
		ln.ln.Close()
	}
	if ln.pconn != nil {
		ln.pconn.Close()
	}
	if ln.network == "unix" {
		os.RemoveAll(ln.addr)
	}
}

// system takes the net listener and detaches it from it's parent
// event loop, grabs the file descriptor, and makes it non-blocking.
func (ln *listener) system() error {
	var err error
	switch netln := ln.ln.(type) {
	case nil:
		switch pconn := ln.pconn.(type) {
		case *net.UDPConn:
			ln.f, err = pconn.File()
		}
	case *net.TCPListener:
		ln.f, err = netln.File()
	case *net.UnixListener:
		ln.f, err = netln.File()
	}
	if err != nil {
		ln.close()
		return err
	}
	ln.fd = int(ln.f.Fd())
	return syscall.SetNonblock(ln.fd, true)
}

type addrOpts struct {
	reusePort bool
}

// Serve starts handling events for the specified addresses.
//
// Addresses should use a scheme prefix and be formatted
// like `tcp://192.168.0.10:9851` or `unix://socket`.
// Valid network schemes:
//
//	tcp   - bind to both IPv4 and IPv6
//	tcp4  - IPv4
//	tcp6  - IPv6
//	udp   - bind to both IPv4 and IPv6
//	udp4  - IPv4
//	udp6  - IPv6
//	unix  - Unix Domain Socket
//
// The "tcp" network scheme is assumed when one is not specified.
func Serve(events Events, addr ...string) error {
	var listeners []*listener
	defer func() {
		for _, listener := range listeners {
			listener.close()
		}
	}()

	var stdlib bool
	for _, addr := range addr {
		var ln listener
		var stdlibt bool
		ln.network, ln.addr, ln.opts, stdlibt = parseAddr(addr)
		if stdlibt {
			stdlib = true
		}
		if ln.network == "unix" {
			_ = os.RemoveAll(ln.addr)
		}
		var err error
		if ln.network == "udp" {
			if ln.opts.reusePort {
				ln.pconn, err = reuseportListenPacket(ln.network, ln.addr)
			} else {
				ln.pconn, err = net.ListenPacket(ln.network, ln.addr)
			}
		} else {
			if ln.opts.reusePort {
				ln.ln, err = reuseportListen(ln.network, ln.addr)
			} else {
				ln.ln, err = net.Listen(ln.network, ln.addr)
			}
		}
		if err != nil {
			return err
		}
		if ln.pconn != nil {
			ln.lnaddr = ln.pconn.LocalAddr()
		} else {
			ln.lnaddr = ln.ln.Addr()
		}
		if !stdlib {
			if err := ln.system(); err != nil {
				return err
			}
		}
		listeners = append(listeners, &ln)
	}

	/*if stdlib { // close stdserve
		return stdserve(events, lns)
	}*/

	return serve(events, listeners)
}

func parseAddr(addr string) (network, address string, opts addrOpts, stdlib bool) {
	network = "tcp"
	address = addr
	opts.reusePort = false
	if strings.Contains(address, "://") {
		network = strings.Split(address, "://")[0]
		address = strings.Split(address, "://")[1]
	}
	if strings.HasSuffix(network, "-net") {
		stdlib = true
		network = network[:len(network)-4]
	}
	q := strings.Index(address, "?")
	if q != -1 {
		for _, part := range strings.Split(address[q+1:], "&") {
			kv := strings.Split(part, "=")
			if len(kv) == 2 {
				switch kv[0] {
				case "reuseport":
					if len(kv[1]) != 0 {
						switch kv[1][0] {
						default:
							opts.reusePort = kv[1][0] >= '1' && kv[1][0] <= '9'
						case 'T', 't', 'Y', 'y':
							opts.reusePort = true
						}
					}
				}
			}
		}
		address = address[:q]
	}
	return
}
