package main

import (
    "errors"
    "fmt"
    "github.com/cilium/ebpf"
    "golang.org/x/sys/unix"
    "strings"
    "syscall"
)

/**
架构设计, destination 设计的意义，主要是 labels 是 [255]byte 不固定长度：

Labels are convenient for humans but they are of variable length. Dealing with variable length data in BPF
is cumbersome and slow, so the BPF program never references labels at all.

Instead, the user space code allocates fixed length numeric IDs, which are then used in the BPF.
Each ID represents a (label, domain, protocol) tuple, internally called destination.
*/

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

// DestinationID is a numeric identifier for a destination.
type DestinationID uint32

type Domain uint8

const (
    AF_INET  Domain = unix.AF_INET
    AF_INET6 Domain = unix.AF_INET6
)

func (d Domain) String() string {
    switch d {
    case AF_INET:
        return "ipv4"
    case AF_INET6:
        return "ipv6"
    default:
        return fmt.Sprintf("unknown(%d)", uint8(d))
    }
}

type Protocol uint8

// Valid protocols.
const (
    TCP Protocol = unix.IPPROTO_TCP
    UDP Protocol = unix.IPPROTO_UDP
)

func (p Protocol) String() string {
    switch p {
    case TCP:
        return "tcp"
    case UDP:
        return "udp"
    default:
        return fmt.Sprintf("unknown(%d)", uint8(p))
    }
}

func ConvertProtocol(protocol string) Protocol {
    switch strings.ToLower(protocol) {
    case "tcp":
        return TCP
    case "udp":
        return UDP
    default:
        return TCP
    }
}

type Destination struct {
    Label    string   // foo
    Domain   Domain   // ipv4/ipv6
    Protocol Protocol // tcp/udp
}

func newDestinationFromBinding(bind *Binding) *Destination {
    domain := AF_INET
    if bind.Prefix.IP().Is6() {
        domain = AF_INET6
    }

    return &Destination{bind.Label, domain, bind.Protocol}
}

func (dest *Destination) String() string {
    return fmt.Sprintf("%s:%s:%s", dest.Domain, dest.Protocol, dest.Label)
}

// INFO: 包含对 fd 的 socket_fd TCP/UDP 检查
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
    listening := acceptConn == 1 // 这里为何认为 ==1 是 listening

    // INFO: 来判断当前 socket_fd 是不是 unconnected for UDP
    unconnected := false
    if _, err = unix.Getpeername(int(fd)); err != nil { // 这里为何认为 err != nil 是 unconnected, 因为 get peer 有 error
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
    maxID        DestinationID
}

func NewDestinations(maps dispatcherMaps) *Destinations {
    return &Destinations{
        destinations: maps.Destinations,
        sockets:      maps.Sockets,
        metrics:      maps.DestinationMetrics,
        maxID:        DestinationID(maps.Sockets.MaxEntries()),
    }
}

func (destinations *Destinations) Acquire(dest *Destination) (DestinationID, error) {
    destKey, err := newDestinationKey(dest)
    if err != nil {
        return 0, err
    }

    destValue, err := destinations.getAllocation(destKey)
    if err != nil {
        return 0, fmt.Errorf("get allocation for %v: %s", destKey, err)
    }

    destValue.Count++
    if destValue.Count == 0 {
        return 0, fmt.Errorf("acquire binding %v: counter overflow", destKey)
    }

    // ?? 为何不是 UpdateAny
    if err = destinations.destinations.Update(destKey, destValue, ebpf.UpdateExist); err != nil {
        return 0, fmt.Errorf("acquire binding %v: %s", destKey, err)
    }

    return destValue.ID, nil
}

func (destinations *Destinations) AddSocket(dest *Destination, conn syscall.Conn) (created bool, err error) {
    key, err := newDestinationKey(dest)
    if err != nil {
        return false, err
    }

    destValue, err := destinations.getAllocation(key)
    if err != nil {
        return false, err
    }

    err = Control(conn, func(fd int) error {
        err := destinations.sockets.Update(destValue.ID, uint64(fd), ebpf.UpdateExist)
        if errors.Is(err, ebpf.ErrKeyNotExist) {
            created = true
            err = destinations.sockets.Update(destValue.ID, uint64(fd), ebpf.UpdateNoExist)
        }
        return err
    })
    if err != nil {
        return false, fmt.Errorf("update socket map: %s", err)
    }

    return
}

type destinationKey struct {
    Label    label // 注意 [255]byte
    Domain   Domain
    Protocol Protocol
}

func newDestinationKey(dest *Destination) (*destinationKey, error) {
    key := &destinationKey{
        Domain:   dest.Domain,
        Protocol: dest.Protocol,
    }

    if dest.Label == "" {
        return nil, fmt.Errorf("label is empty")
    }
    if strings.ContainsRune(dest.Label, 0) {
        return nil, fmt.Errorf("label contains null byte")
    }
    if max := len(key.Label); len(dest.Label) > max {
        return nil, fmt.Errorf("label exceeds maximum length of %d bytes", max)
    }

    copy(key.Label[:], dest.Label)
    return key, nil
}

type destinationValue struct {
    ID    DestinationID
    Count uint32
}

func (destinations *Destinations) getAllocation(key *destinationKey) (*destinationValue, error) {
    alloc := new(destinationValue)
    if err := destinations.destinations.Lookup(key, alloc); err == nil {
        return alloc, nil
    }

    alloc = &destinationValue{ID: id}
    // This may replace an unused-but-not-deleted allocation.
    if err := destinations.destinations.Update(key, alloc, ebpf.UpdateAny); err != nil {
        return nil, fmt.Errorf("allocate destination: %s", err)
    }

    return alloc, nil
}

// ReleaseByID releases a reference on a destination by its ID.
//
// This function is linear to the number of destinations and should be avoided
// if possible.
func (destinations *Destinations) ReleaseByID(id DestinationID) error {
    var (
        key   destinationKey
        alloc destinationValue
        iter  = destinations.destinations.Iterate()
    )
    for iter.Next(&key, &alloc) {
        if alloc.ID != id {
            continue
        }

        return destinations.releaseAllocation(&key, alloc)
    }
    if err := iter.Err(); err != nil {
        return err
    }

    return fmt.Errorf("release reference: no allocation for id %d", id)
}

// Close INFO: 这里才 close bpf map
func (destinations *Destinations) Close() error {
    if err := destinations.destinations.Close(); err != nil {
        return err
    }
    if err := destinations.metrics.Close(); err != nil {
        return err
    }
    return destinations.sockets.Close()
}
