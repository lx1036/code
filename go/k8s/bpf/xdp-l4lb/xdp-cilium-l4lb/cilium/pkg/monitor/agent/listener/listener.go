package listener

import (
    "golang.org/x/sys/unix"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/payload"
    "net"
    "os"
)

// Version is the version of a node-monitor listener client. There are
// two API versions:
// - 1.0 which encodes the gob type information with each payload sent, and
//   adds a meta object before it.
// - 1.2 which maintains a gob session per listener, thus only encoding the
//   type information on the first payload sent. It does NOT prepend the a meta
//   object.
type Version string

const (
    // VersionUnsupported is here for use in error returns etc.
    VersionUnsupported = Version("unsupported")

    // Version1_2 is the API 1.0 version of the protocol (see above).
    Version1_2 = Version("1.2")
)

// MonitorListener is a generic consumer of monitor events. Implementers are
// expected to handle errors as needed, including exiting.
type MonitorListener interface {
    // Enqueue adds this payload to the send queue. Any errors should be logged
    // and handled appropriately.
    Enqueue(pl *payload.Payload)

    // Version returns the API version of this listener
    Version() Version

    // Close closes the listener.
    Close()
}

// IsDisconnected is a convenience function that wraps the absurdly long set of
// checks for a disconnect.
// 判断 unix socket 是否 disconnected
func IsDisconnected(err error) bool {
    if err == nil {
        return false
    }

    op, ok := err.(*net.OpError)
    if !ok {
        return false
    }

    syscerr, ok := op.Err.(*os.SyscallError)
    if !ok {
        return false
    }

    errn := syscerr.Err.(unix.Errno)
    return errn == unix.EPIPE
}
