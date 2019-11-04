package net

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

type listener struct {
	ln      net.Listener
	lnaddr  net.Addr
	pconn   net.PacketConn
	f       *os.File
	fd      int
	network string
	addr    string
}

func (ln *listener) close() {
	if ln.f != nil {
		sniffError(ln.f.Close())
	}
	if ln.ln != nil {
		sniffError(ln.ln.Close())
	}
	if ln.pconn != nil {
		sniffError(ln.pconn.Close())
	}
	if ln.network == "unix" {
		sniffError(os.RemoveAll(ln.addr))
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
	return unix.SetNonblock(ln.fd, true)
}

type server struct {
	ln               *listener          // all the listeners
	wg               sync.WaitGroup     // loop close waitgroup
	tch              chan time.Duration // ticker channel
	opts             *Options           // options with server
	once             sync.Once          // make sure only signalShutdown once
	cond             *sync.Cond         // shutdown signaler
	mainLoop         *loop              // main loop for accepting connections
	eventHandler     EventHandler       // user eventHandler
	subLoopGroup     IEventLoopGroup    // loops for handling events
	subLoopGroupSize int                // number of loops
}

func (svr *server) start(numCPU int) error {
	if svr.opts.ReusePort || svr.ln.pconn != nil {
		return svr.activateLoops(numCPU)
	}
	return svr.activateReactors(numCPU)
}

func (svr *server) activateLoops(numLoops int) error {

}
func (svr *server) activateReactors(numLoops int) error {

}

func (svr *server) closeLoops() {
	svr.subLoopGroup.iterate(func(i int, lp *loop) bool {
		_ = lp.poller.Close()
		return true
	})
}

func (svr *server) stop() {

}


const (
	// None indicates that no action should occur following an event.
	None Action = iota
	// DataRead indicates data in buffer has been read.
	DataRead
	// Close closes the connection.
	Close
	// Shutdown shutdowns the server.
	Shutdown
)

func serve(eventHandler EventHandler, listener *listener, options *Options) error {
	// Figure out the correct number of loops/goroutines to use.
	var numCPU int
	if options.Multicore {
		numCPU = runtime.NumCPU()
	} else {
		numCPU = 1
	}

	svr := new(server)
	svr.eventHandler = eventHandler
	svr.ln = listener
	svr.subLoopGroup = new(eventLoopGroup)
	svr.cond = sync.NewCond(&sync.Mutex{})
	svr.tch = make(chan time.Duration)
	svr.opts = options

	server := Server{options.Multicore, listener.lnaddr, numCPU}
	action := svr.eventHandler.OnInitComplete(server)
	switch action {
	case None:
	case Shutdown:
		return nil
	}
	if err := svr.start(numCPU); err != nil {
		svr.closeLoops()
		log.Printf("gnet server is stoping with error: %v\n", err)
		return err
	}
	defer svr.stop()

	return nil
}

