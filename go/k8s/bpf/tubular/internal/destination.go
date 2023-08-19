package internal

import "golang.org/x/sys/unix"

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
