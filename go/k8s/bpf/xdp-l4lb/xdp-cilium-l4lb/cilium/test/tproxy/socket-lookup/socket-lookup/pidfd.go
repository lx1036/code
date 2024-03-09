package main

import (
    "errors"
    "fmt"
    "golang.org/x/sys/unix"
    "inet.af/netaddr"
    "os"
)

// Files 获取 pid 打开的 socket fd，比如 nginx 424308 打开的监听在 127.0.0.1:80 的 socket_fd
func Files(pid int, ps ...Predicate) (files []*os.File, err error) {
    const maxFDGap = 32

    defer func() {
        if err != nil {
            for _, file := range files {
                file.Close()
            }
        }
    }()

    // > Linux 5.10, 获取进程 pid 对应的 fd，即 pid_fd
    // https://man7.org/linux/man-pages/man2/pidfd_open.2.html
    pidfd, err := unix.PidfdOpen(pid, 0)
    if err != nil {
        return nil, err
    }
    defer unix.Close(pidfd)

    for i, gap := 0, 0; i < int(^uint(0)>>1) && gap < maxFDGap; i++ {
        // fmt.Println(int(^uint(0) >> 1)) // 9223372036854775807
        // https://man7.org/linux/man-pages/man2/pidfd_getfd.2.html
        // This new file descriptor is a duplicate of an existing file descriptor, targetfd.
        // 获取 i 的 duplicate of targetfd, targetfd 是一个真正的 fd，然后经过 filter 来判断监听的 127.0.0.1:80 对应的 socket_fd
        target, err := unix.PidfdGetfd(pidfd, i, 0)
        if errors.Is(err, unix.EBADF) {
            gap++
            continue
        }
        if err != nil {
            return nil, fmt.Errorf("target fd %d: %s", i, err)
        }
        gap = 0

        keep, err := FilterFd(target, ps...)
        if err != nil {
            unix.Close(target)
            return nil, fmt.Errorf("target fd %d: %w", i, err)
        } else if keep {
            files = append(files, os.NewFile(uintptr(target), ""))
        } else {
            unix.Close(target)
        }
    }

    return files, nil
}

// Predicate is a condition for keeping or rejecting a file.
type Predicate func(fd int) (keep bool, err error)

func FilterFd(fd int, ps ...Predicate) (keep bool, err error) {
    for _, p := range ps {
        keep, err = p(fd)
        if err != nil || !keep {
            return
        }
    }
    return
}

// FirstReuseport filters out all but the first socket of a reuseport group.
//
// Non-reuseport sockets and non-sockets are ignored.
func FirstReuseport() Predicate {
    type key struct {
        proto int
        ip    netaddr.IP
        port  uint16
    }

    seen := make(map[key]bool)
    return func(fd int) (bool, error) {
        reuseport, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT)
        if err != nil {
            return false, fmt.Errorf("getsockopt(SO_REUSEPORT): %w", err)
        }
        if reuseport != 1 {
            return true, nil
        }

        proto, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_PROTOCOL)
        if err != nil {
            return false, fmt.Errorf("getsockopt(SO_PROTOCOL): %w", err)
        }

        sa, err := unix.Getsockname(fd)
        if err != nil {
            return false, fmt.Errorf("getsockname: %w", err)
        }

        k := key{proto: proto}
        switch addr := sa.(type) {
        case *unix.SockaddrInet4:
            k.ip, _ = netaddr.FromStdIP(addr.Addr[:])
            k.port = uint16(addr.Port)
        case *unix.SockaddrInet6:
            k.ip = netaddr.IPv6Raw(addr.Addr)
            k.port = uint16(addr.Port)
        default:
            return false, fmt.Errorf("unsupported address family: %T", sa)
        }

        if seen[k] {
            return false, nil
        }

        seen[k] = true
        return true, nil
    }
}

// IgnoreENOTSOCK wraps a predicate and returns false instead of unix.ENOTSOCK.
func IgnoreENOTSOCK(p Predicate) Predicate {
    return func(fd int) (bool, error) {
        keep, err := p(fd)
        if errors.Is(err, unix.ENOTSOCK) {
            return false, nil
        }
        return keep, err
    }
}

// LocalAddress filters for sockets with the given address and port.
func LocalAddress(ip netaddr.IP, port int) Predicate {
    return func(fd int) (bool, error) {
        sa, err := unix.Getsockname(fd)
        if err != nil {
            return false, fmt.Errorf("getsockname: %s", err)
        }

        var fdIP netaddr.IP
        var fdPort int
        switch addr := sa.(type) {
        case *unix.SockaddrInet4:
            fdIP, _ = netaddr.FromStdIP(addr.Addr[:])
            fdPort = addr.Port

        case *unix.SockaddrInet6:
            fdIP = netaddr.IPv6Raw(addr.Addr)
            fdPort = addr.Port

        default:
            return false, nil
        }

        if fdIP.Compare(ip) != 0 {
            return false, nil
        }

        if fdPort != port {
            return false, nil
        }

        return true, nil
    }
}

// InetListener returns a predicate that keeps listening TCP or connected UDP sockets.
//
// It filters out any files that are not sockets.
func InetListener(network string) Predicate {
    return func(fd int) (bool, error) {
        domain, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_DOMAIN)
        if err != nil {
            return false, err
        }
        if domain != unix.AF_INET && domain != unix.AF_INET6 {
            return false, nil
        }

        soType, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_TYPE)
        if err != nil {
            return false, fmt.Errorf("getsockopt(SO_TYPE): %s", err)
        }

        switch network {
        case "udp":
            if soType != unix.SOCK_DGRAM {
                return false, nil
            }

        case "tcp":
            if soType != unix.SOCK_STREAM {
                return false, nil
            }

        default:
            return false, fmt.Errorf("unrecognized network %q", network)
        }

        switch soType {
        case unix.SOCK_STREAM:
            acceptConn, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_ACCEPTCONN)
            if err != nil {
                return false, fmt.Errorf("getsockopt(SO_ACCEPTCONN): %s", err)
            }

            if acceptConn == 0 {
                // Not a listening socket
                return false, nil
            }

        case unix.SOCK_DGRAM:
            sa, err := unix.Getpeername(fd)
            if err != nil && !errors.Is(err, unix.ENOTCONN) {
                return false, fmt.Errorf("getpeername: %s", err)
            }

            if sa != nil {
                // Not a connected socket
                return false, nil
            }

        default:
            return false, nil
        }

        return true, nil
    }
}
