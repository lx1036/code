package bgp

import (
	"context"
	"fmt"
	"net"
	"os"
	"syscall"

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
