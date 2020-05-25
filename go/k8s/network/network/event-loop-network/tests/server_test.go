package tests

import (
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestServe(test *testing.T) {
	// start a server
	// connect 10 clients
	// each client will pipe random data for 1-3 seconds.
	// the writes to the server will be random sizes. 0KB - 1MB.
	// the server will echo back the data.
	// waits for graceful connection closing.

	test.Run("poll", func(test *testing.T) {
		test.Run("tcp", func(test *testing.T) {
			test.Run("1-loop", func(test *testing.T) {
				testServe("tcp", ":9991", false, 10, 1, event_loop.Random)
			})
		})
	})
}

func testServe(network, address string, unix bool, nclients int, nloops int, balance event_loop.LoadBalance) {
	var started int32
	var connected int32
	var disconnected int32

	var events event_loop.Events
	events.LoadBalance = balance
	events.NumLoops = nloops
	events.Serving = func(server event_loop.Server) (action event_loop.Action) {
		return
	}

	events.Opened = func(c event_loop.Conn) (out []byte, opts net.Options, action event_loop.Action) {
		c.SetContext(c)
		atomic.AddInt32(&connected, 1)
		out = []byte("sweetness\r\n")
		opts.TCPKeepAlive = time.Minute * 5
		if c.LocalAddr() == nil {
			panic("nil local addr")
		}
		if c.RemoteAddr() == nil {
			panic("nil local addr")
		}
		return
	}

	events.Closed = func(c event_loop.Conn, err error) (action event_loop.Action) {
		if c.Context() != c {
			panic("invalid context")
		}
		atomic.AddInt32(&disconnected, 1)
		if atomic.LoadInt32(&connected) == atomic.LoadInt32(&disconnected) &&
			atomic.LoadInt32(&disconnected) == int32(nclients) {
			action = event_loop.Shutdown
		}
		return
	}
	events.Data = func(c event_loop.Conn, in []byte) (out []byte, action event_loop.Action) {
		out = in
		return
	}
	events.Tick = func() (delay time.Duration, action event_loop.Action) {
		if atomic.LoadInt32(&started) == 0 {
			for i := 0; i < nclients; i++ {
				go startClient(network, address, nloops)
			}
			atomic.StoreInt32(&started, 1)
		}
		delay = time.Second / 5
		return
	}

	var err error
	if unix {
		socket := strings.Replace(address, ":", "socket", 1)
		_ = os.RemoveAll(socket)
		defer os.RemoveAll(socket)
		err = event_loop.Serve(events, network+"://"+address, "unix://"+socket)
	} else {
		err = event_loop.Serve(events, network+"://"+address)
	}
	if err != nil {
		panic(err)
	}

}

func startClient(network, addr string, nloops int) {

}
