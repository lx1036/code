// +build linux

package bgp

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"net"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// TODO: 查下 BGP 标准文档
func sendKeepalive(w io.Writer) error {
	msg := struct {
		Marker1, Marker2 uint64
		Len              uint16
		Type             uint8
	}{
		Marker1: 0xffffffffffffffff,
		Marker2: 0xffffffffffffffff,
		Len:     19,
		Type:    4,
	}
	return binary.Write(w, binary.BigEndian, msg)
}

type openResult struct {
	asn      uint32
	holdTime time.Duration
	mp4      bool
	mp6      bool
	// Four-byte ASN supported
	fbasn bool
}

func sendOpen(w io.Writer, asn uint32, routerID net.IP, holdTime time.Duration) error {
	if routerID.To4() == nil {
		panic("non-ipv4 address used as RouterID")
	}

	msg := struct {
		// Header
		Marker1, Marker2 uint64
		Len              uint16
		Type             uint8

		// OPEN
		Version  uint8
		ASN16    uint16
		HoldTime uint16
		RouterID [4]byte

		// Options (we only send one, capabilities)
		OptsLen uint8
		OptType uint8
		OptLen  uint8

		// Capabilities: multiprotocol extension for IPv4+IPv6
		// unicast, and 4-byte ASNs

		MP4Type uint8
		MP4Len  uint8
		AFI4    uint16
		SAFI4   uint16

		MP6Type uint8
		MP6Len  uint8
		AFI6    uint16
		SAFI6   uint16

		CapType uint8
		CapLen  uint8
		ASN32   uint32
	}{
		Marker1: 0xffffffffffffffff,
		Marker2: 0xffffffffffffffff,
		Len:     0, // Filled below
		Type:    1, // OPEN

		Version:  4,
		ASN16:    uint16(asn), // Possibly tweaked below
		HoldTime: uint16(holdTime.Seconds()),
		// RouterID filled below

		OptsLen: 20,
		OptType: 2, // Capabilities
		OptLen:  18,

		MP4Type: 1, // BGP Multi-protocol Extensions
		MP4Len:  4,
		AFI4:    1, // IPv4
		SAFI4:   1, // Unicast

		MP6Type: 1, // BGP Multi-protocol Extensions
		MP6Len:  4,
		AFI6:    2, // IPv6
		SAFI6:   1, // Unicast

		CapType: 65, // 4-byte ASN
		CapLen:  4,
		ASN32:   asn,
	}
	msg.Len = uint16(binary.Size(msg))
	if asn > 65535 {
		msg.ASN16 = 23456
	}
	copy(msg.RouterID[:], routerID.To4())

	return binary.Write(w, binary.BigEndian, msg)
}

func readOpen(r io.Reader) (*openResult, error) {
	hdr := struct {
		// Header
		Marker1, Marker2 uint64
		Len              uint16
		Type             uint8
	}{}
	if err := binary.Read(r, binary.BigEndian, &hdr); err != nil {
		return nil, err
	}
	if hdr.Marker1 != 0xffffffffffffffff || hdr.Marker2 != 0xffffffffffffffff {
		return nil, fmt.Errorf("synchronization error, incorrect header marker")
	}
	if hdr.Type == 3 {
		return nil, readNotification(r)
	}
	if hdr.Type != 1 {
		return nil, fmt.Errorf("message type is not OPEN, got %d, want 1", hdr.Type)
	}
	if hdr.Len < 37 {
		return nil, fmt.Errorf("message length %d too small to be OPEN", hdr.Len)
	}

	lr := &io.LimitedReader{
		R: r,
		N: int64(hdr.Len) - 19,
	}
	open := struct {
		Version  uint8
		ASN16    uint16
		HoldTime uint16
		RouterID uint32
		OptsLen  uint8
	}{}
	if err := binary.Read(lr, binary.BigEndian, &open); err != nil {
		return nil, err
	}
	fmt.Printf("%#v\n", open)
	if open.Version != 4 {
		return nil, fmt.Errorf("wrong BGP version")
	}
	if open.HoldTime != 0 && open.HoldTime < 3 {
		return nil, fmt.Errorf("invalid hold time %q, must be 0 or >=3s", open.HoldTime)
	}

	ret := &openResult{
		asn:      uint32(open.ASN16),
		holdTime: time.Duration(open.HoldTime) * time.Second,
	}

	if err := readOptions(lr, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func readOptions(r io.Reader, ret *openResult) error {
	for {
		hdr := struct {
			Type uint8
			Len  uint8
		}{}
		if err := binary.Read(r, binary.BigEndian, &hdr); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if hdr.Type != 2 {
			return fmt.Errorf("unknown BGP option type %d", hdr.Type)
		}
		lr := &io.LimitedReader{
			R: r,
			N: int64(hdr.Len),
		}
		if err := readCapabilities(lr, ret); err != nil {
			return err
		}
		if lr.N != 0 {
			return fmt.Errorf("%d trailing garbage bytes after capability option", lr.N)
		}
	}
}

func readCapabilities(r io.Reader, ret *openResult) error {
	for {
		capability := struct {
			Code uint8
			Len  uint8
		}{}
		if err := binary.Read(r, binary.BigEndian, &capability); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		lr := io.LimitedReader{
			R: r,
			N: int64(capability.Len),
		}
		switch capability.Code {
		case 65:
			if err := binary.Read(&lr, binary.BigEndian, &ret.asn); err != nil {
				return err
			}
			ret.fbasn = true
		case 1:
			af := struct{ AFI, SAFI uint16 }{}
			if err := binary.Read(&lr, binary.BigEndian, &af); err != nil {
				return err
			}
			switch {
			case af.AFI == 1 && af.SAFI == 1:
				ret.mp4 = true
			case af.AFI == 2 && af.SAFI == 1:
				ret.mp6 = true
			}
		default:
			// TODO: only ignore capabilities that we know are fine to
			// ignore.
			if _, err := io.Copy(ioutil.Discard, &lr); err != nil {
				return err
			}
		}
		if lr.N != 0 {
			return fmt.Errorf("%d leftover bytes after decoding capability %d", lr.N, capability.Code)
		}
	}
}

var notificationCodes = map[uint16]string{
	0x0100: "Message header error (unspecific)",
	0x0101: "Connection not synchronized",
	0x0102: "Bad message length",
	0x0103: "Bad message type",

	0x0200: "OPEN message error (unspecific)",
	0x0201: "Unsupported version number",
	0x0202: "Bad peer AS",
	0x0203: "Bad BGP identifier",
	0x0204: "Unsupported optional parameter",
	0x0206: "Unacceptable hold time",
	0x0207: "Unsupported capability",

	0x0300: "UPDATE message error (unspecific)",
	0x0301: "Malformed Attribute List",
	0x0302: "Unrecognized Well-known Attribute",
	0x0303: "Missing Well-known Attribute",
	0x0304: "Attribute Flags Error",
	0x0305: "Attribute Length Error",
	0x0306: "Invalid ORIGIN Attribute",
	0x0308: "Invalid NEXT_HOP Attribute",
	0x0309: "Optional Attribute Error",
	0x030a: "Invalid Network Field",
	0x030b: "Malformed AS_PATH",

	0x0400: "Hold Timer Expired (unspecific)",

	0x0500: "BGP FSM state error (unspecific)",
	0x0501: "Receive Unexpected Message in OpenSent State",
	0x0502: "Receive Unexpected Message in OpenConfirm State",
	0x0503: "Receive Unexpected Message in Established State",

	0x0601: "Maximum Number of Prefixes Reached",
	0x0602: "Administrative Shutdown",
	0x0603: "Peer De-configured",
	0x0604: "Administrative Reset",
	0x0605: "Connection Rejected",
	0x0606: "Other Configuration Change",
	0x0607: "Connection Collision Resolution",
	0x0608: "Out of Resources",
}

// readNotification reads the body of a notification message (header
// has already been consumed). It must always return an error, because
// receiving a notification is an error.
func readNotification(r io.Reader) error {
	var code uint16
	if err := binary.Read(r, binary.BigEndian, &code); err != nil {
		return err
	}
	v, ok := notificationCodes[code]
	if !ok {
		v = "unknown code"
	}
	return fmt.Errorf("got BGP notification code 0x%04x (%s)", code, v)
}

func sendUpdate(w io.Writer, asn uint32, ibgp, fbasn bool, defaultNextHop net.IP, adv *Advertisement) error {
	var b bytes.Buffer

	hdr := struct {
		M1, M2  uint64
		Len     uint16
		Type    uint8
		WdrLen  uint16
		AttrLen uint16
	}{
		M1:   uint64(0xffffffffffffffff),
		M2:   uint64(0xffffffffffffffff),
		Type: 2,
	}
	if err := binary.Write(&b, binary.BigEndian, hdr); err != nil {
		return err
	}
	l := b.Len()
	if err := encodePathAttrs(&b, asn, ibgp, fbasn, defaultNextHop, adv); err != nil {
		return err
	}
	binary.BigEndian.PutUint16(b.Bytes()[21:23], uint16(b.Len()-l))
	encodePrefixes(&b, []*net.IPNet{adv.Prefix})
	binary.BigEndian.PutUint16(b.Bytes()[16:18], uint16(b.Len()))

	if _, err := io.Copy(w, &b); err != nil {
		return err
	}
	return nil
}

func encodePathAttrs(b *bytes.Buffer, asn uint32, ibgp, fbasn bool, defaultNextHop net.IP, adv *Advertisement) error {
	b.Write([]byte{
		0x40, 1, // mandatory, origin
		1, // len
		2, // incomplete

		0x40, 2, // mandatory, as-path
	})
	if ibgp {
		b.WriteByte(0) // empty AS path
	} else {
		if fbasn {
			b.Write([]byte{
				6, // len (1x 4-byte ASN)
				2, // AS_SEQUENCE
				1, // len (in number of ASes)
			})
			if err := binary.Write(b, binary.BigEndian, asn); err != nil {
				return err
			}
		} else {
			b.Write([]byte{
				4, // len (1x 2-byte ASN)
				2, // AS_SEQUENCE
				1, // len (in number of ASes)
			})
			if err := binary.Write(b, binary.BigEndian, uint16(asn)); err != nil {
				return err
			}
		}
	}
	b.Write([]byte{
		0x40, 3, // mandatory, next-hop
		4, // len
	})
	if adv.NextHop != nil {
		b.Write(adv.NextHop.To4())
	} else {
		b.Write(defaultNextHop)
	}
	if ibgp {
		b.Write([]byte{
			0x40, 5, // well-known, localpref
			4, // len
		})
		if err := binary.Write(b, binary.BigEndian, adv.LocalPref); err != nil {
			return err
		}
	}

	if len(adv.Communities) > 0 {
		b.Write([]byte{
			0xc0, 8, // optional transitive, communities
		})
		if err := binary.Write(b, binary.BigEndian, uint8(len(adv.Communities)*4)); err != nil {
			return err
		}
		for _, c := range adv.Communities {
			if err := binary.Write(b, binary.BigEndian, c); err != nil {
				return err
			}
		}
	}

	return nil
}

func encodePrefixes(b *bytes.Buffer, pfxs []*net.IPNet) {
	for _, pfx := range pfxs {
		o, _ := pfx.Mask.Size()
		b.WriteByte(byte(o))
		b.Write(pfx.IP.To4()[:bytesForBits(o)])
	}
}

func bytesForBits(n int) int {
	// Evil bit hack that rounds n up to the next multiple of 8, then
	// divides by 8. This returns the minimum number of whole bytes
	// required to contain n bits.
	return ((n + 7) &^ 7) / 8
}

func sendWithdraw(w io.Writer, prefixes []*net.IPNet) error {
	var b bytes.Buffer

	hdr := struct {
		M1, M2 uint64
		Len    uint16
		Type   uint8
		WdrLen uint16
	}{
		M1:   uint64(0xffffffffffffffff),
		M2:   uint64(0xffffffffffffffff),
		Type: 2,
	}
	if err := binary.Write(&b, binary.BigEndian, hdr); err != nil {
		return err
	}
	l := b.Len()
	encodePrefixes(&b, prefixes)
	binary.BigEndian.PutUint16(b.Bytes()[19:21], uint16(b.Len()-l))
	if err := binary.Write(&b, binary.BigEndian, uint16(0)); err != nil {
		return err
	}
	binary.BigEndian.PutUint16(b.Bytes()[16:18], uint16(b.Len()))

	if _, err := io.Copy(w, &b); err != nil {
		return err
	}
	return nil
}

const (
	//tcpMD5SIG TCP MD5 Signature (RFC2385)
	tcpMD5SIG = 14
)

// This struct is defined at; linux-kernel: include/uapi/linux/tcp.h,
// It must be kept in sync with that definition, see current version:
// https://github.com/torvalds/linux/blob/v4.16/include/uapi/linux/tcp.h#L253
// nolint[structcheck]
type tcpmd5sig struct {
	ssFamily uint16
	ss       [126]byte
	pad1     uint16
	keylen   uint16
	pad2     uint32
	key      [80]byte
}

// INFO: 这个函数必须在 linux 上运行

// DialTCP does the part of creating a connection manually,  including setting the
// proper TCP MD5 options when the password is not empty. Works by manupulating
// the low level FD's, skipping the net.Conn API as it has not hooks to set
// the neccessary sockopts for TCP MD5.
func dialMD5(ctx context.Context, addr, password string) (net.Conn, error) {
	laddr, err := net.ResolveTCPAddr("tcp", "[::]:0")
	if err != nil {
		return nil, fmt.Errorf("Error resolving local address: %s ", err)
	}

	raddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid remote address: %s ", err)
	}

	var family int
	var ra, la unix.Sockaddr
	if raddr.IP.To4() != nil {
		family = unix.AF_INET
		rsockaddr := &unix.SockaddrInet4{Port: raddr.Port}
		copy(rsockaddr.Addr[:], raddr.IP.To4())
		ra = rsockaddr
		lsockaddr := &unix.SockaddrInet4{}
		copy(lsockaddr.Addr[:], laddr.IP.To4())
		la = lsockaddr
	} else {
		family = unix.AF_INET6
		rsockaddr := &unix.SockaddrInet6{Port: raddr.Port}
		copy(rsockaddr.Addr[:], raddr.IP.To16())
		ra = rsockaddr
		var zone uint32
		if laddr.Zone != "" {
			intf, errs := net.InterfaceByName(laddr.Zone)
			if errs != nil {
				return nil, errs
			}
			zone = uint32(intf.Index)
		}
		lsockaddr := &unix.SockaddrInet6{ZoneId: zone}
		copy(lsockaddr.Addr[:], laddr.IP.To16())
		la = lsockaddr
	}

	sockType := unix.SOCK_STREAM | unix.SOCK_CLOEXEC | unix.SOCK_NONBLOCK
	proto := 0
	fd, err := unix.Socket(family, sockType, proto)
	if err != nil {
		return nil, err
	}

	// A new socket was created so we must close it before this
	// function returns either on failure or success. On success,
	// net.FileConn() in newTCPConn() increases the refcount of
	// the socket so this fi.Close() doesn't destroy the socket.
	// The caller must call Close() with the file later.
	// Note that the above os.NewFile() doesn't play with the
	// refcount.
	fi := os.NewFile(uintptr(fd), "")
	defer fi.Close()

	if password != "" {
		sig := buildTCPMD5Sig(raddr.IP, password)
		b := *(*[unsafe.Sizeof(sig)]byte)(unsafe.Pointer(&sig))
		// Better way may be available in  Go 1.11, see go-review.googlesource.com/c/go/+/72810
		if err = os.NewSyscallError("setsockopt", unix.SetsockoptString(fd, unix.IPPROTO_TCP, tcpMD5SIG, string(b[:]))); err != nil {
			return nil, err
		}
	}

	if err = unix.Bind(fd, la); err != nil {
		return nil, os.NewSyscallError("bind", err)
	}

	err = unix.Connect(fd, ra)

	switch err {
	case syscall.EINPROGRESS, syscall.EALREADY, syscall.EINTR:
	case nil:
		return net.FileConn(fi)
	default:
		return nil, os.NewSyscallError("connect", err)
	}

	// With a non-blocking socket, the connection process is
	// asynchronous, so we need to manually wait with epoll until the
	// connection succeeds. All of the following is doing that, with
	// appropriate use of the deadline in the context.
	epfd, err := unix.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	defer unix.Close(epfd)

	var event unix.EpollEvent
	events := make([]unix.EpollEvent, 1)

	event.Events = syscall.EPOLLIN | syscall.EPOLLOUT | syscall.EPOLLPRI
	event.Fd = int32(fd)
	if err = unix.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &event); err != nil {
		return nil, err
	}

	for {
		timeout := int(-1)
		if deadline, ok := ctx.Deadline(); ok {
			timeout = int(time.Until(deadline).Nanoseconds() / 1000000)
			if timeout <= 0 {
				return nil, fmt.Errorf("timeout")
			}
		}
		nevents, err := unix.EpollWait(epfd, events, timeout)
		if err != nil {
			return nil, err
		}
		if nevents == 0 {
			return nil, fmt.Errorf("timeout")
		}
		if nevents > 1 || events[0].Fd != int32(fd) {
			return nil, fmt.Errorf("unexpected epoll behavior")
		}

		nerr, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil {
			return nil, os.NewSyscallError("getsockopt", err)
		}
		switch err := syscall.Errno(nerr); err {
		case syscall.EINPROGRESS, syscall.EALREADY, syscall.EINTR:
		case syscall.Errno(0), unix.EISCONN:
			return net.FileConn(fi)
		default:
			return nil, os.NewSyscallError("getsockopt", err)
		}
	}
}

func buildTCPMD5Sig(addr net.IP, key string) tcpmd5sig {
	t := tcpmd5sig{}
	if addr.To4() != nil {
		t.ssFamily = unix.AF_INET
		copy(t.ss[2:], addr.To4())
	} else {
		t.ssFamily = unix.AF_INET6
		copy(t.ss[6:], addr.To16())
	}

	t.keylen = uint16(len(key))
	copy(t.key[0:], []byte(key))

	return t
}
