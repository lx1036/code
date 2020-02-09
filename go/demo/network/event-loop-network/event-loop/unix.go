package event_loop

import (
	"github.com/libp2p/go-reuseport"
	"k8s-lx1036//demo/network/event-loop-network/internal"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type server struct {
	events Events         // user events
	loops  []*loop        // all the loops
	lns    []*listener    // all the listeners
	wg     sync.WaitGroup // loop close waitgroup
	cond   *sync.Cond     // shutdown signaler
	//balance  LoadBalance        // load balancing method
	accepted uintptr            // accept counter
	tch      chan time.Duration // ticker channel

	//ticktm   time.Time      // next tick time
}

// waitForShutdown waits for a signal to shutdown
func (s *server) waitForShutdown() {
	s.cond.L.Lock()
	s.cond.Wait()
	s.cond.L.Unlock()
}

type loop struct {
	idx     int            // loop index in the server loops list
	poll    *internal.Poll // epoll or kqueue
	packet  []byte         // read packet buffer
	fdconns map[int]*conn  // loop connections fd -> conn
	count   int32          // connection count
}

// LoadBalance sets the load balancing method.
type LoadBalance int

const (
	// Random requests that connections are randomly distributed.
	Random LoadBalance = iota
	// RoundRobin requests that connections are distributed to a loop in a
	// round-robin fashion.
	RoundRobin
	// LeastConnections assigns the next accepted connection to the loop with
	// the least number of active connections.
	LeastConnections
)

func reuseportListenPacket(proto, addr string) (l net.PacketConn, err error) {
	return reuseport.ListenPacket(proto, addr)
}

func reuseportListen(proto, addr string) (l net.Listener, err error) {
	return reuseport.Listen(proto, addr)
}

const (
	// None indicates that no action should occur following an event.
	None Action = iota
	// Detach detaches a connection. Not available for UDP connections.
	Detach
	// Close closes the connection.
	Close
	// Shutdown shutdowns the server.
	Shutdown
)

type conn struct {
	fd         int              // file descriptor
	lnidx      int              // listener index in the server lns list
	out        []byte           // write buffer
	sa         syscall.Sockaddr // remote socket address
	reuse      bool             // should reuse input buffer
	opened     bool             // connection opened event fired
	action     Action           // next user action
	ctx        interface{}      // user-defined context
	addrIndex  int              // index of listening address
	localAddr  net.Addr         // local addre
	remoteAddr net.Addr         // remote addr
	loop       *loop            // connected loop
}

func (c conn) Context() interface{} {
	panic("implement me")
}

func (c conn) SetContext(interface{}) {
	panic("implement me")
}

func (c conn) AddrIndex() int {
	panic("implement me")
}

func (c conn) LocalAddr() net.Addr {
	panic("implement me")
}

func (c conn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (c conn) Wake() {
	panic("implement me")
}

func serve(events Events, listeners []*listener) error {
	// figure out the correct number of loops/goroutines to use.
	numLoops := events.NumLoops
	if numLoops <= 0 {
		if numLoops == 0 {
			numLoops = 1
		} else {
			numLoops = runtime.NumCPU()
		}
	}

	server := &server{}
	server.events = events
	server.lns = listeners
	server.cond = sync.NewCond(&sync.Mutex{})
	//server.balance = events.LoadBalance
	server.tch = make(chan time.Duration)

	if server.events.Serving != nil {
		var svr Server
		svr.NumLoops = numLoops
		svr.Addrs = make([]net.Addr, len(listeners))
		for i, ln := range listeners {
			svr.Addrs[i] = ln.lnaddr
		}
		action := server.events.Serving(svr)
		switch action {
		case None:
		case Shutdown:
			return nil
		}
	}

	defer func() {
		// wait on a signal for shutdown
		server.waitForShutdown()

		// notify all loops to close by closing all listeners
		for _, loop := range server.loops {
			loop.poll.Trigger(errClosing)
		}

		// wait on all loops to complete reading events
		server.wg.Wait()

		// close loops and all outstanding connections
		for _, loop := range server.loops {
			for _, c := range loop.fdconns {
				_ = loopCloseConn(server, loop, c, nil)
			}
			loop.poll.Close()
		}
		//println("-- server stopped")
	}()

	// create loops locally and bind the listeners.
	for i := 0; i < numLoops; i++ {
		l := &loop{
			idx:     i,
			poll:    internal.OpenPoll(),
			packet:  make([]byte, 0xFFFF),
			fdconns: make(map[int]*conn),
		}
		for _, ln := range listeners {
			l.poll.AddRead(ln.fd)
		}
		server.loops = append(server.loops, l)
	}

	// start loops in background
	server.wg.Add(len(server.loops))
	for _, l := range server.loops {
		go loopRun(server, l)
	}

	return nil
}

func loopRun(s *server, l *loop) {

}

func loopCloseConn(s *server, l *loop, c *conn, err error) error {
	atomic.AddInt32(&l.count, -1)
	delete(l.fdconns, c.fd)
	_ = syscall.Close(c.fd)
	if s.events.Closed != nil {
		switch s.events.Closed(c, err) {
		case None:
		case Shutdown:
			return errClosing
		}
	}
	return nil
}
