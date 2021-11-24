package bgp

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/klog/v2"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"
)

const BGP_PORT = 179
const LocalAddress = "127.0.0.1"

// Advertisement represents one network path and its BGP attributes.
type Advertisement struct {
	// The prefix being advertised to the peer.
	Prefix *net.IPNet
	// The address of the router to which the peer should forward traffic.
	NextHop net.IP
	// The local preference of this route. Only propagated to IBGP
	// peers (i.e. where the peer ASN matches the local ASN).
	LocalPref uint32
	// BGP communities to attach to the path.
	Communities []uint32
}

// Equal returns true if a and b are equivalent advertisements.
func (a *Advertisement) Equal(b *Advertisement) bool {
	if a.Prefix.String() != b.Prefix.String() {
		return false
	}
	if !a.NextHop.Equal(b.NextHop) {
		return false
	}
	if a.LocalPref != b.LocalPref {
		return false
	}
	return reflect.DeepEqual(a.Communities, b.Communities)
}

var errClosed = errors.New("session closed")

// Session represents one BGP session to an external router.
type Session struct {
	asn              uint32
	routerID         net.IP // May be nil, meaning "derive from context"
	myNode           string
	addr             string
	peerASN          uint32
	peerFBASNSupport bool
	holdTime         time.Duration
	password         string

	newHoldTime chan bool
	backoff     backoff

	mu             sync.Mutex
	cond           *sync.Cond
	closed         bool
	conn           net.Conn
	actualHoldTime time.Duration
	defaultNextHop net.IP
	advertised     map[string]*Advertisement
	new            map[string]*Advertisement
}

// New INFO: 会立即与 router server 建立 BGP session 连接
func New(addr string, asn uint32, routerID net.IP, peerASN uint32, holdTime time.Duration, myNode string) (*Session, error) {
	s := &Session{
		addr:        addr, // ip:port
		asn:         asn,
		routerID:    routerID.To4(),
		myNode:      myNode,
		peerASN:     peerASN,
		holdTime:    holdTime,
		newHoldTime: make(chan bool, 1),
		advertised:  map[string]*Advertisement{},
	}
	s.cond = sync.NewCond(&s.mu)

	go s.run()
	go s.sendKeepalives()

	return s, nil
}

// run tries to stay connected to the peer, and pumps route updates to it.
func (s *Session) run() {
	for {
		if err := s.connect(); err != nil { // try again if connect fail
			if err == errClosed {
				return
			}
			klog.Infof(fmt.Sprintf("failed to connect to peer"))
			backoff := s.backoff.Duration()
			time.Sleep(backoff)
			continue
		}
		s.backoff.Reset()

		klog.Infof(fmt.Sprintf("BGP session established"))

		if !s.sendUpdates() {
			return
		}
		klog.Infof(fmt.Sprintf("BGP session down"))
	}
}

// connect establishes the BGP session with the peer.
// sets TCP_MD5 sockopt if password is !="",
func (s *Session) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errClosed
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := s.active(ctx)
	if err != nil {
		return err
	}

	laddr, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		conn.Close()
		return fmt.Errorf("getting local addr for default nexthop to %q: %s", s.addr, err)
	}
	s.defaultNextHop = laddr.IP

	// INFO: fsm opensent
	if err = sendOpen(conn, s.asn, routerID, s.holdTime); err != nil {
		conn.Close()
		return fmt.Errorf("send OPEN to %q: %s", s.addr, err)
	}
	op, err := readOpen(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("read OPEN from %q: %s", s.addr, err)
	}

}

func (s *Session) active(ctx context.Context) (net.Conn, error) {
	connCh := make(chan net.Conn)
	errCh := make(chan error)
	go func() {
		laddr, _ := net.ResolveTCPAddr("tcp", net.JoinHostPort(LocalAddress, "0"))
		d := net.Dialer{
			LocalAddr: laddr,
			Timeout:   5 * time.Second,
		}
		conn, err := d.DialContext(ctx, "tcp", s.addr) // ip:port
		if err != nil {
			errCh <- err
		}
		connCh <- conn
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout for fsm idle->active")
		case err := <-errCh:
			return nil, err
		case conn := <-connCh:
			return conn, nil
		}
	}
}
