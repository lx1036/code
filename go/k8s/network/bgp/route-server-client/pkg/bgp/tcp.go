package bgp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"syscall"
	"time"
	
	"golang.org/x/sys/unix"
)

// DialTCP does the part of creating a connection manually,  including setting the
// proper TCP MD5 options when the password is not empty. Works by manupulating
// the low level FD's, skipping the net.Conn API as it has not hooks to set
// the neccessary sockopts for TCP MD5.
func dialMD5(ctx context.Context, addr string, srcAddr net.IP) (net.Conn, error) {
	src := fmt.Sprintf("[%s]", srcAddr.String())
	localAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", src))
	if err != nil {
		return nil, fmt.Errorf("Error resolving local address: %s ", err)
	}
	peerAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid remote address: %s ", err)
	}

	var family int
	var peerSocketAddr, localSocketAddr unix.Sockaddr
	if peerAddr.IP.To4() != nil {
		family = unix.AF_INET
		rsockaddr := &unix.SockaddrInet4{Port: peerAddr.Port}
		copy(rsockaddr.Addr[:], peerAddr.IP.To4())
		peerSocketAddr = rsockaddr
		lsockaddr := &unix.SockaddrInet4{}
		copy(lsockaddr.Addr[:], localAddr.IP.To4())
		localSocketAddr = lsockaddr
	}

	//sockType := unix.SOCK_STREAM | unix.SOCK_CLOEXEC | unix.SOCK_NONBLOCK
	sockType := unix.SOCK_STREAM
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
	if err = unix.Bind(fd, localSocketAddr); err != nil {
		return nil, os.NewSyscallError("bind", err)
	}
	err = unix.Connect(fd, peerSocketAddr)
	switch err {
	case syscall.EINPROGRESS, syscall.EALREADY, syscall.EINTR:
	case nil:
		return net.FileConn(fi)
	default:
		return nil, os.NewSyscallError("connect", err)
	}

	// TODO:

	return nil, nil
}

type openResult struct {
	asn      uint32
	holdTime time.Duration
	mp4      bool
	mp6      bool
	// Four-byte ASN supported
	fbasn bool
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
	fmt.Printf("%#v\n", hdr)
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
		c := struct {
			Code uint8
			Len  uint8
		}{}
		if err := binary.Read(r, binary.BigEndian, &c); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		lr := io.LimitedReader{
			R: r,
			N: int64(c.Len),
		}
		switch c.Code {
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
			return fmt.Errorf("%d leftover bytes after decoding capability %d", lr.N, c.Code)
		}
	}
}

func ipnet(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return n
}
