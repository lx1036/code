package net

import (
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type EventServer struct {
}

type Option func(opts *Options)

// Options are set when the client opens.
type Options struct {
	// Multicore indicates whether the server will be effectively created with multi-cores, if so,
	// then you must take care with synchonizing memory between all event callbacks, otherwise,
	// it will run the server with single thread. The number of threads in the server will be automatically
	// assigned to the value of runtime.NumCPU().
	Multicore bool
	//ReusePort ..
	ReusePort bool
	// Ticker ...
	Ticker bool
	// TCPKeepAlive (SO_KEEPALIVE) socket option.
	TCPKeepAlive time.Duration
}

// WithMulticore ...
func WithMulticore(multicore bool) Option {
	return func(opts *Options) {
		opts.Multicore = multicore
	}
}

// Server represents a server context which provides information about the
// running server and has control functions for managing state.
type Server struct {
	// Multicore indicates whether the server will be effectively created with multi-cores, if so,
	// then you must take care with synchonizing memory between all event callbacks, otherwise,
	// it will run the server with single thread. The number of threads in the server will be automatically
	// assigned to the value of runtime.NumCPU().
	Multicore bool

	// The addrs parameter is an array of listening addresses that align
	// with the addr strings passed to the Serve function.
	Addr net.Addr

	// NumLoops is the number of loops that the server is using.
	NumLoops int
}

// Conn is an gnet connection.
type Connection interface {
	// Context returns a user-defined context.
	Context() interface{}

	// SetContext sets a user-defined context.
	SetContext(interface{})

	// LocalAddr is the connection's local socket address.
	LocalAddr() net.Addr

	// RemoteAddr is the connection's remote peer address.
	RemoteAddr() net.Addr

	// Wake triggers a React event for this connection.
	//Wake()

	// ReadPair reads all data from ring buffer.
	ReadPair() ([]byte, []byte)

	// ReadBytes reads all data and return a new slice.
	ReadBytes() []byte

	// ResetBuffer resets the ring buffer.
	ResetBuffer()

	// AyncWrite writes data asynchronously.
	AsyncWrite(buf []byte)
}

// Action is an action that occurs after the completion of an event.
type Action int

// EventHandler represents the server events' callbacks for the Serve call.
// Each event has an Action return value that is used manage the state
// of the connection and server.
type EventHandler interface {
	// OnInitComplete fires when the server can accept connections. The server
	// parameter has information and various utilities.
	OnInitComplete(server Server) (action Action)

	// OnOpened fires when a new connection has opened.
	// The info parameter has information about the connection such as
	// it's local and remote address.
	// Use the out return value to write data to the connection.
	// The opts return value is used to set connection options.
	OnOpened(c Connection) (out []byte, action Action)

	// OnClosed fires when a connection has closed.
	// The err parameter is the last known connection error.
	OnClosed(c Connection, err error) (action Action)

	// PreWrite fires just before any data is written to any client socket.
	PreWrite()

	// React fires when a connection sends the server data.
	// The in parameter is the incoming data.
	// Use the out return value to write data to the connection.
	React(c Connection) (out []byte, action Action)

	// Tick fires immediately after the server starts and will fire again
	// following the duration specified by the delay return value.
	Tick() (delay time.Duration, action Action)
}

func sniffError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func initOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}
func parseAddr(addr string) (network, address string) {
	network = "tcp"
	address = addr
	if strings.Contains(address, "://") {
		network = strings.Split(address, "://")[0]
		address = strings.Split(address, "://")[1]
	}
	return
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
func Serve(eventHandler EventHandler, addr string, opts ...Option) error {
	var ln listener
	defer ln.close()

	options := initOptions(opts...)
	ln.network, ln.addr = parseAddr(addr)
	if ln.network == "unix" {
		sniffError(os.RemoveAll(ln.addr))
	}
	var err error
	if ln.network == "udp" {
		if options.ReusePort {
			ln.pconn, err = netpoll.ReusePortListenPacket(ln.network, ln.addr)
		} else {
			ln.pconn, err = net.ListenPacket(ln.network, ln.addr)
		}
	} else {
		if options.ReusePort {
			ln.ln, err = netpoll.ReusePortListen(ln.network, ln.addr)
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
	if err := ln.system(); err != nil {
		return err
	}
	return serve(eventHandler, &ln, options)
}
