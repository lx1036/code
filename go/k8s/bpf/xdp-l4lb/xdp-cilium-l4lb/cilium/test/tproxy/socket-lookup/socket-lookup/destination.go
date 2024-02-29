package main

import (
    "errors"
    "fmt"
    "github.com/cilium/ebpf"
    "golang.org/x/sys/unix"
    "syscall"
)

var (
    ErrLoaded            = errors.New("dispatcher already loaded")
    ErrNotLoaded         = errors.New("dispatcher not loaded")
    ErrNotSocket         = syscall.ENOTSOCK
    ErrBadSocketDomain   = syscall.EPFNOSUPPORT
    ErrBadSocketType     = syscall.ESOCKTNOSUPPORT
    ErrBadSocketProtocol = syscall.EPROTONOSUPPORT
    ErrBadSocketState    = syscall.EBADFD
)

// systemd supports names of up to 255 bytes, match the limit.
type label [255]byte

// destinationID is a numeric identifier for a destination.
type destinationID uint32

type Domain uint8

const (
    AF_INET  Domain = unix.AF_INET
    AF_INET6 Domain = unix.AF_INET6
)

type Protocol uint8

// Valid protocols.
const (
    TCP Protocol = unix.IPPROTO_TCP
    UDP Protocol = unix.IPPROTO_UDP
)

type destinationKey struct {
    Label    label
    Domain   Domain
    Protocol Protocol
}

type destinationAlloc struct {
    ID    destinationID
    Count uint32
}

type Destination struct {
    Label    string
    Domain   Domain
    Protocol Protocol
}

func newDestinationFromFd(label string, fd uintptr) (*Destination, error) {
    var stat unix.Stat_t
    err := unix.Fstat(int(fd), &stat)
    if err != nil {
        return nil, fmt.Errorf("fstat: %w", err)
    }
    if stat.Mode&unix.S_IFMT != unix.S_IFSOCK {
        return nil, fmt.Errorf("fd is not a socket: %w", ErrNotSocket)
    }

    domain, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_DOMAIN)
    if err != nil {
        return nil, fmt.Errorf("get SO_DOMAIN: %w", err)
    }

    sotype, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_TYPE)
    if err != nil {
        return nil, fmt.Errorf("get SO_TYPE: %w", err)
    }

    proto, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_PROTOCOL)
    if err != nil {
        return nil, fmt.Errorf("get SO_PROTOCOL: %w", err)
    }

    // INFO: 来判断当前 socket_fd 是不是 listening for TCP
    acceptConn, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_ACCEPTCONN)
    if err != nil {
        return nil, fmt.Errorf("get SO_ACCEPTCONN: %w", err)
    }
    listening := acceptConn == 1

    // INFO: 来判断当前 socket_fd 是不是 unconnected for UDP
    unconnected := false
    if _, err = unix.Getpeername(int(fd)); err != nil {
        if !errors.Is(err, unix.ENOTCONN) {
            return nil, fmt.Errorf("getpeername: %w", err)
        }
        unconnected = true
    }

    if domain != unix.AF_INET && domain != unix.AF_INET6 {
        return nil, fmt.Errorf("unsupported socket domain %v: %w", domain, ErrBadSocketDomain)
    }
    if sotype != unix.SOCK_STREAM && sotype != unix.SOCK_DGRAM {
        return nil, fmt.Errorf("unsupported socket type %v: %w", sotype, ErrBadSocketType)
    }
    if sotype == unix.SOCK_STREAM && proto != unix.IPPROTO_TCP {
        return nil, fmt.Errorf("unsupported stream socket protocol %v: %w", proto, ErrBadSocketProtocol)
    }
    if sotype == unix.SOCK_DGRAM && proto != unix.IPPROTO_UDP {
        return nil, fmt.Errorf("unsupported packet socket protocol %v: %w", proto, ErrBadSocketDomain)
    }
    if sotype == unix.SOCK_STREAM && !listening {
        return nil, fmt.Errorf("stream socket not listening: %w", ErrBadSocketState)
    }
    if sotype == unix.SOCK_DGRAM && !unconnected {
        return nil, fmt.Errorf("packet socket is connected: %w", ErrBadSocketState)
    }

    dest := &Destination{
        Label:    label,
        Domain:   Domain(domain),
        Protocol: Protocol(proto),
    }

    return dest, nil
}

func newDestinationFromConn(label string, conn syscall.Conn) (*Destination, error) {
    var dest *Destination
    err := Control(conn, func(fd int) (err error) {
        dest, err = newDestinationFromFd(label, uintptr(fd))
        return
    })
    if err != nil {
        return nil, err
    }

    return dest, nil
}

type Destinations struct {
    destinations *ebpf.Map
    sockets      *ebpf.Map
    metrics      *ebpf.Map
    maxID        destinationID
}

func NewDestinations(maps dispatcherMaps) *Destinations {
    return &Destinations{
        destinations: maps.Destinations,
        sockets:      maps.Sockets,
        metrics:      maps.DestinationMetrics,
        maxID:        destinationID(maps.Sockets.MaxEntries()),
    }
}

func (destinations *Destinations) AddSocket(dest *Destination, conn syscall.Conn) (created bool, err error) {
    key, err := newDestinationKey(dest)
    if err != nil {
        return false, err
    }

    alloc, err := destinations.getAllocation(key)
    if err != nil {
        return false, err
    }

    err = Control(conn, func(fd int) error {
        err := destinations.sockets.Update(alloc.ID, uint64(fd), ebpf.UpdateExist)
        if errors.Is(err, ebpf.ErrKeyNotExist) {
            created = true
            err = destinations.sockets.Update(alloc.ID, uint64(fd), ebpf.UpdateNoExist)
        }
        return err
    })
    if err != nil {
        return false, fmt.Errorf("update socket map: %s", err)
    }

}
