package pkg

import (
	"fmt"
	"net"
	"time"
)

// tcp keep-alive listener with stopC
type listener struct {
	*net.TCPListener
	stopC <-chan struct{}
}

func newListenerWithStopC(addr string, stopC <-chan struct{}) (*listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &listener{
		TCPListener: ln.(*net.TCPListener),
		stopC:       stopC,
	}, nil
}

func (ln *listener) Accept() (net.Conn, error) {
	connC := make(chan *net.TCPConn, 1)
	errC := make(chan error, 1)
	go func() {
		conn, err := ln.AcceptTCP()
		if err != nil {
			errC <- err
			return
		}
		connC <- conn
	}()

	select {
	case <-ln.stopC:
		return nil, fmt.Errorf("tcp server stopped")
	case err := <-errC:
		return nil, err
	case conn := <-connC:
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Minute * 3)
		return conn, nil
	}
}
