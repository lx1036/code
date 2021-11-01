package bgp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type session struct {
	mu sync.Mutex

	newHoldTime chan bool
	routerID    net.IP // May be nil, meaning "derive from context"

	addr     string
	srcAddr  net.IP
	localASN uint32
	peerASN  uint32

	holdTime time.Duration

	defaultNextHop net.IP

	closed bool
}

func NewSession(addr string, srcAddr net.IP, localASN, peerASN uint32, routerID net.IP, holdTime time.Duration) *session {

	s := &session{
		newHoldTime: make(chan bool, 1),

		addr:     addr,
		srcAddr:  srcAddr,
		localASN: localASN,
		peerASN:  peerASN,
		routerID: routerID.To4(),

		holdTime: holdTime,
	}

	go s.sendKeepalives()
	go s.run()

	return s
}

// sendKeepalives sends BGP KEEPALIVE packets at the negotiated rate whenever the session is connected.
func (s *session) sendKeepalives() {

}

func (s *session) run() {
	for {

		s.connect()

	}
}

// connect establishes the BGP session with the peer.
// Sets TCP_MD5 sockopt if password is !="".
func (s *session) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("session closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	deadline, _ := ctx.Deadline()
	conn, err := dialMD5(ctx, s.addr, s.srcAddr)
	if err != nil {
		return fmt.Errorf("dial %q: %s", s.addr, err)
	}
	if err = conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return fmt.Errorf("setting deadline on conn to %q: %s", s.addr, err)
	}

	addr, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		conn.Close()
		return fmt.Errorf("getting local addr for default nexthop to %q: %s", s.addr, err)
	}
	s.defaultNextHop = addr.IP

	routerID := s.routerID
	if err = sendOpen(conn, s.localASN, routerID, s.holdTime); err != nil {
		conn.Close()
		return fmt.Errorf("send OPEN to %q: %s", s.addr, err)
	}

	op, err := readOpen(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("read OPEN from %q: %s", s.addr, err)
	}
	if op.peerASN != s.peerASN {
		conn.Close()
		return fmt.Errorf("unexpected peer ASN %d, want %d", op.peerASN, s.peerASN)
	}
	s.peerFBASNSupport = op.fbasn
	if s.localASN > 65536 && !s.peerFBASNSupport {
		conn.Close()
		return fmt.Errorf("peer does not support 4-byte ASNs")
	}

	// BGP session is established, clear the connect timeout deadline.
	if err := conn.SetDeadline(time.Time{}); err != nil {
		conn.Close()
		return fmt.Errorf("clearing deadline on conn to %q: %s", s.addr, err)
	}

	// Consume BGP messages until the connection closes.
	go s.consumeBGP(conn)

	// Send one keepalive to say that yes, we accept the OPEN.
	if err := sendKeepalive(conn); err != nil {
		conn.Close()
		return fmt.Errorf("accepting peer OPEN from %q: %s", s.addr, err)
	}

	// Set up regular keepalives from now on.
	s.actualHoldTime = s.holdTime
	if op.holdTime < s.actualHoldTime {
		s.actualHoldTime = op.holdTime
	}
	select {
	case s.newHoldTime <- true:
	default:
	}

	s.conn = conn
	return nil
}
