package bgp

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net"
	"reflect"
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
	myasn            uint32
	routerID         net.IP // May be nil, meaning "derive from context"
	myNode           string
	raddr            string
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
func New(raddr string, routerID net.IP, peerASN uint32, myasn uint32, holdTime time.Duration, myNode string) (*Session, error) {
	s := &Session{
		raddr:       raddr, // remote addr ip:port
		myasn:       myasn,
		routerID:    routerID.To4(),
		myNode:      myNode,
		peerASN:     peerASN,
		holdTime:    holdTime,
		newHoldTime: make(chan bool, 1),
		advertised:  map[string]*Advertisement{},
	}
	s.cond = sync.NewCond(&s.mu)

	go s.sendKeepalives()
	go s.run()

	return s, nil
}

// run tries to stay connected to the peer, and pumps route updates to it.
func (s *Session) run() {
	for {
		if err := s.connect(); err != nil { // try again if connect fail
			if err == errClosed {
				return
			}
			klog.Infof(fmt.Sprintf("failed to connect to peer: %v", err))
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

// sendKeepalives sends BGP KEEPALIVE packets at the negotiated rate whenever the session is connected
// INFO: 这里心跳时间必须小于 holdTime，一般取 1/3 (https://github.com/osrg/gobgp/blob/v2.33.0/pkg/server/fsm.go#L1336-L1340)
//  这里非常重要，否则router server 会报错 holdTime expired，state 由 establish->active
//  查看接收的路由：gobgp -p 50063 -d neighbor 127.0.0.1 adj-in
//  router server 接收到心跳keepalive后，会在fsm establish state中重置holdTimer，变为 10s 之后才会过期:
//  @see https://github.com/osrg/gobgp/blob/v2.33.0/pkg/server/fsm.go#L1080-L1090
//  @see https://github.com/osrg/gobgp/blob/v2.33.0/pkg/server/fsm.go#L1826-L1850
func (s *Session) sendKeepalives() {
	var ch <-chan time.Time
	for {
		select {
		case <-s.newHoldTime:
			ch = time.NewTicker(s.actualHoldTime / 3).C
			klog.Infof(fmt.Sprintf("[sendKeepalives]keepalive interval %s", s.actualHoldTime.String()))
		case <-ch:
			if err := sendKeepalive(s.conn); err != nil {
				klog.Error(err)
			}
		}
	}
}

// sendUpdates waits for changes to desired advertisements, and pushes them out to the peer.
//  INFO: 这里使用了 Cond 来实现锁
func (s *Session) sendUpdates() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return false
	}
	if s.conn == nil {
		return true
	}

	ibgp := s.myasn == s.peerASN
	fbasn := s.peerFBASNSupport
	if s.new != nil {
		s.advertised, s.new = s.new, nil
	}
	for _, adv := range s.advertised {
		if err := sendUpdate(s.conn, s.myasn, ibgp, fbasn, s.defaultNextHop, adv); err != nil {
			//s.abort()
			klog.Infof("failed to send BGP update")
			return true
		}
	}

	for {
		for s.new == nil && s.conn != nil {
			s.cond.Wait() // INFO: 加锁等待 AddPath() 新的路由
		}

		if s.closed {
			return false
		}
		if s.conn == nil {
			return true
		}
		if s.new == nil {
			continue
		}

		//s.defaultNextHop = net.ParseIP("10.20.30.40")
		for c, adv := range s.new {
			if adv2, ok := s.advertised[c]; ok && adv.Equal(adv2) {
				// Peer already has correct state for this advertisement, nothing to do.
				continue
			}
			if err := sendUpdate(s.conn, s.myasn, ibgp, fbasn, s.defaultNextHop, adv); err != nil {
				//s.abort()
				klog.Infof("failed to send BGP update")
				return true
			}
		}

		// INFO: 如果 s.new 是 empty map，则删除已经宣告的 s.advertised 路由
		wdr := []*net.IPNet{}
		for c, adv := range s.advertised {
			if s.new[c] == nil {
				wdr = append(wdr, adv.Prefix)
			}
		}
		if len(wdr) > 0 {
			if err := sendWithdraw(s.conn, wdr); err != nil {
				//s.abort()
				for _, pfx := range wdr {
					klog.Infof(fmt.Sprintf("failed to send BGP update prefix %s", pfx.String()))
				}
				return true
			}
		}

		s.advertised, s.new = s.new, nil
	}
}

// AddPath INFO: 宣告路由给 router server
func (s *Session) AddPath(advs ...*Advertisement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newAdvs := map[string]*Advertisement{}
	for _, adv := range advs {
		if adv.Prefix.IP.To4() == nil {
			return fmt.Errorf("cannot advertise non-v4 prefix %q", adv.Prefix)
		}

		if adv.NextHop != nil && adv.NextHop.To4() == nil {
			return fmt.Errorf("next-hop must be IPv4, got %q", adv.NextHop)
		}
		if len(adv.Communities) > 63 {
			return fmt.Errorf("max supported communities is 63, got %d", len(adv.Communities))
		}
		newAdvs[adv.Prefix.String()] = adv
	}

	s.new = newAdvs
	s.cond.Broadcast() // INFO: 解锁 sendUpdates()
	return nil
}

func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	//s.abort()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.cond.Broadcast()
	return nil
}

// connect establishes the BGP session with the peer.
// sets TCP_MD5 sockopt if password is !="",
func (s *Session) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errClosed
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// INFO: fsm active
	conn, err := s.active(ctx)
	if err != nil {
		return err
	}
	laddr, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		conn.Close()
		return fmt.Errorf("getting local addr for default nexthop to %q: %s", s.raddr, err)
	}
	s.defaultNextHop = laddr.IP
	//s.defaultNextHop = net.ParseIP("10.20.30.40")
	routerID := s.routerID
	if routerID == nil {
		routerID = s.defaultNextHop // ipv4
	}
	s.conn = conn
	s.routerID = routerID

	// INFO: fsm opensent
	/*msg, err := s.opensent(ctx)
	if err != nil {
		s.conn.Close()
		return err
	}*/
	if err = sendOpen(conn, s.myasn, routerID, s.holdTime); err != nil {
		conn.Close()
		return fmt.Errorf("send OPEN to %q: %s", s.raddr, err)
	}
	msg, err := readOpen(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("read OPEN from %q: %s", s.raddr, err)
	}
	klog.Infof(fmt.Sprintf("[fsm opensent]%+v", msg))
	if msg.asn != s.peerASN {
		conn.Close()
		return fmt.Errorf("unexpected peer ASN %d, want %d", msg.asn, s.peerASN)
	}
	s.peerFBASNSupport = msg.fbasn
	// BGP session is established, clear the connect timeout deadline.
	if err := conn.SetDeadline(time.Time{}); err != nil {
		conn.Close()
		return fmt.Errorf("clearing deadline on conn to %q: %s", s.raddr, err)
	}

	// INFO: fsm openconfirm
	// Consume BGP messages until the connection closes.
	go s.consumeBGP(conn)

	err = sendKeepalive(s.conn)
	if err != nil {
		s.conn.Close()
		return err
	}

	klog.Infof(fmt.Sprintf("[fsm opensent]holdtime %s", msg.holdTime.String()))
	s.actualHoldTime = s.holdTime
	if msg.holdTime < s.actualHoldTime {
		s.actualHoldTime = msg.holdTime
	}
	select {
	case s.newHoldTime <- true:
	default:
	}
	return nil
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
		conn, err := d.DialContext(ctx, "tcp", s.raddr) // ip:port
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

func (s *Session) opensent(ctx context.Context) (*openResult, error) {
	err := sendOpen(s.conn, s.myasn, s.routerID, s.holdTime)
	if err != nil {
		return nil, err
	}

	msg := make(chan *openResult)
	go func() {
		op, err := readOpen(s.conn)
		if err != nil {
			klog.Error(err)
			return
		}

		msg <- op
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout for fsm active->opensent")
		case op := <-msg:
			return op, nil
		}
	}
}

func (s *Session) openconfirm(ctx context.Context) error {
	ticker := time.NewTicker(time.Second) // 每秒一次心跳

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout for fsm active->opensent")
		case <-ticker.C:
			err := sendKeepalive(s.conn)
			if err != nil {
				klog.Error(err)
				return err
			}

		}
	}
}

// consumeBGP receives BGP messages from the peer, and ignores
// them. It does minimal checks for the well-formedness of messages,
// and terminates the connection if something looks wrong.
func (s *Session) consumeBGP(conn io.ReadCloser) {
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.conn == conn {
			//s.abort()
		} else {
			conn.Close()
		}
	}()

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
			klog.Infof("peer sent notification, closing session: %v", err)
			return
		}
		if _, err := io.Copy(ioutil.Discard, io.LimitReader(conn, int64(hdr.Len)-19)); err != nil {
			// TODO: propagate
			return
		}
	}
}
