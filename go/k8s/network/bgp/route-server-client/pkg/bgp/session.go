package bgp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net"
	"sync"
	"time"
)

type session struct {
	mu   sync.Mutex
	cond *sync.Cond
	conn net.Conn

	routerID net.IP // May be nil, meaning "derive from context"

	addr     string
	srcAddr  net.IP
	localASN uint32
	peerASN  uint32

	newHoldTime    chan bool
	holdTime       time.Duration
	actualHoldTime time.Duration

	defaultNextHop   net.IP
	advertisement    map[string]*Advertisement
	newAdvertisement map[string]*Advertisement

	peerFBASNSupport bool

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

		holdTime:         holdTime,
		advertisement:    map[string]*Advertisement{},
		newAdvertisement: map[string]*Advertisement{},
	}
	s.cond = sync.NewCond(&s.mu)

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
	if op.asn != s.peerASN {
		conn.Close()
		return fmt.Errorf("unexpected peer ASN %d, want %d", op.asn, s.peerASN)
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

// consumeBGP receives BGP messages from the peer, and ignores
// them. It does minimal checks for the well-formedness of messages,
// and terminates the connection if something looks wrong.
func (s *session) consumeBGP(conn io.ReadCloser) {

	for {
		hdr := struct {
			Marker1, Marker2 uint64
			Len              uint16
			Type             uint8
		}{}
		if err := binary.Read(conn, binary.BigEndian, &hdr); err != nil {
			// TODO: log, or propagate the error somehow.
			return
		}
		if hdr.Marker1 != 0xffffffffffffffff || hdr.Marker2 != 0xffffffffffffffff {
			// TODO: propagate
			return
		}
		if hdr.Type == 3 {
			// TODO: propagate better than just logging directly.
			err := readNotification(conn)
			klog.Infof(fmt.Sprintf("event peerNotification msg peer sent notification, closing session error:%v", err))
			return
		}
		if _, err := io.Copy(ioutil.Discard, io.LimitReader(conn, int64(hdr.Len)-19)); err != nil {
			// TODO: propagate
			return
		}
	}
}

// Set updates the set of Advertisements that this session's peer should receive.
//
// Changes are propagated to the peer asynchronously, Set may return
// before the peer learns about the changes.
func (s *session) Set(advs ...*Advertisement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newAdvertisement := map[string]*Advertisement{}
	for _, adv := range advs {
		err := validate(adv)
		if err != nil {
			return err
		}
		newAdvertisement[adv.Prefix.String()] = adv
	}

	s.newAdvertisement = newAdvertisement
	//stats.PendingPrefixes(s.addr, len(s.new))
	s.cond.Broadcast()
	return nil
}

// Close shuts down the BGP session.
func (s *session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.abort()
	return nil
}

// abort closes any existing connection, updates stats, and cleans up
// state ready for another connection attempt.
func (s *session) abort() {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
		//stats.SessionDown(s.addr)
	}
	// Next time we retry the connection, we can just skip straight to
	// the desired end state.
	if s.newAdvertisement != nil {
		s.advertisement, s.newAdvertisement = s.newAdvertisement, nil
		//stats.PendingPrefixes(s.addr, len(s.advertised))
	}
	s.cond.Broadcast()
}

func validate(adv *Advertisement) error {
	if adv.Prefix.IP.To4() == nil {
		return fmt.Errorf("cannot advertise non-v4 prefix %q", adv.Prefix)
	}

	if adv.NextHop != nil && adv.NextHop.To4() == nil {
		return fmt.Errorf("next-hop must be IPv4, got %q", adv.NextHop)
	}
	if len(adv.Communities) > 63 {
		return fmt.Errorf("max supported communities is 63, got %d", len(adv.Communities))
	}
	return nil
}
