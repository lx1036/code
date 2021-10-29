package server

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"syscall"
)

type tcpListener struct {
	l  *net.TCPListener
	ch chan struct{}
}

func (l *tcpListener) Close() error {
	if err := l.l.Close(); err != nil {
		return err
	}
	<-l.ch
	return nil
}

// avoid mapped IPv6 address
func newTCPListener(address string, port uint32, bindToDev string, ch chan *net.TCPConn) (*tcpListener, error) {
	proto := "tcp4"
	family := syscall.AF_INET
	if ip := net.ParseIP(address); ip == nil {
		return nil, fmt.Errorf("can't listen on %s", address)
	} else if ip.To4() == nil {
		proto = "tcp6"
		family = syscall.AF_INET6
	}

	addr := net.JoinHostPort(address, strconv.Itoa(int(port)))

	var lc net.ListenConfig
	lc.Control = func(network, address string, c syscall.RawConn) error {
		/*if bindToDev != "" {
			err := setBindToDevSockopt(c, bindToDev)
			if err != nil {
				log.WithFields(log.Fields{
					"Topic":     "Peer",
					"Key":       addr,
					"BindToDev": bindToDev,
				}).Warnf("failed to bind Listener to device (%s): %s", bindToDev, err)
				return err
			}
		}*/
		// Note: Set TTL=255 for incoming connection listener in order to accept
		// connection in case for the neighbor has TTL Security settings.
		err := setsockoptIpTtl(c, family, 255)
		if err != nil {
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   addr,
			}).Warnf("cannot set TTL(=%d) for TCPListener: %s", 255, err)
		}
		return nil
	}

	l, err := lc.Listen(context.Background(), proto, addr)
	if err != nil {
		return nil, err
	}
	listener, ok := l.(*net.TCPListener)
	if !ok {
		err = fmt.Errorf("unexpected connection listener (not for TCP)")
		return nil, err
	}

	closeCh := make(chan struct{})
	go func() error {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				close(closeCh)
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Error": err,
				}).Warn("Failed to AcceptTCP")
				return err
			}
			ch <- conn
		}
	}()
	return &tcpListener{
		l:  listener,
		ch: closeCh,
	}, nil
}
