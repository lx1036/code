package main

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "os/signal"
    "syscall"
    "testing"
    "unsafe"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_tcp_socket_cookie.c -- -I.

// go generate .

const (
    CgroupPath = "/sys/fs/cgroup/tcp_socket_cookie"

    PinPath1 = "/sys/fs/bpf/socket_service/tcp_socket_cookie"

    INADDR_LOOPBACK = "127.0.0.1"
)

// CGO_ENABLED=0 go test -v -run ^TestTcpSocketCookie$ .
func TestTcpSocketCookie(test *testing.T) {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    cgroupPath := joinCgroup("tcphdr_opt")
    //defer cleanupCgroup()

    //1.Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: PinPath1, // pin 下 map
        },
    }
    err := loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    defer objs.Close()

    //2.Attach programs to cgroup
    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.SetCookie,
        Attach:  ebpf.AttachCGroupInet4Connect,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    l2, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.UpdateCookie,
        Attach:  ebpf.AttachCGroupSockOps,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l2.Close()

    serverFd := makeServer()
    defer unix.Close(serverFd)
    clientFd := connectToFd(serverFd)
    defer unix.Close(clientFd)

    // validate map
    var value bpfSocketCookie
    // clientFd 是 int, 必须 unsafe.Pointer(), 否则报错: "can't marshal key: binary.Write: invalid type int"
    err = objs.bpfMaps.SocketCookies.Lookup(unsafe.Pointer(&clientFd), &value)
    if err != nil {
        logrus.Fatal(err)
    }
    clientSockAddr, err := unix.Getsockname(clientFd)
    if err != nil {
        logrus.Fatal(err)
    }
    cookieVal := clientSockAddr.(*unix.SockaddrInet4).Port<<8 | 0xFF
    if value.CookieValue != uint32(cookieVal) {
        logrus.Errorf("Unexpected value in map: %x != %x", value.CookieValue, cookieVal)
    } else {
        // d228ff 13773055, d228 为2个字节, hex(6060)=0x17ac, int(0x17ac)=6060
        logrus.Infof("expected value in map: 0x%x %d", cookieVal, cookieVal)
    }

    // Wait
    <-stopCh
}

// tcp listen 127.0.0.1:60123
func makeServer() int {
    serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(serverFd, 5000)

    err = unix.SetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Fatalf("unix.SO_REUSEADDR error: %v", err)
    }

    ip := net.ParseIP(INADDR_LOOPBACK)
    sa := &unix.SockaddrInet4{
        Port: 60123,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(serverFd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.Listen(serverFd, 128)
    if err != nil {
        logrus.Fatal(err)
    }

    return serverFd
}

func connectToFd(serverFd int) int {
    clientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0) // unix.SOCK_STREAM
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    // 非阻塞的
    err = unix.Connect(clientFd, serverSockAddr)
    if err != nil {
        logrus.Fatal(err)
    }

    return clientFd
}

func setSocketTimeout(fd, timeoutMs int) {
    var timeVal *unix.Timeval
    if timeoutMs > 0 {
        timeVal = &unix.Timeval{
            Sec:  int64(timeoutMs / 1000),
            Usec: int64(timeoutMs % 1000 * 1000),
        }
    } else {
        timeVal = &unix.Timeval{
            Sec: 3,
        }
    }

    err := unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, timeVal)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, timeVal)
    if err != nil {
        logrus.Fatal(err)
    }
}

// 把当前进程 pid 写到新建的 connect_force_port cgroup
func joinCgroup(path string) string {
    if len(path) == 0 {
        logrus.Fatalf("path is empty")
    }

    cgroupPath := fmt.Sprintf("%s/%s", CgroupPath, path)
    if err := os.MkdirAll(cgroupPath, 0777); err != nil {
        logrus.Fatalf("os.Mkdir err: %v", err)
    }
    pid := os.Getpid()
    file := fmt.Sprintf("%s/cgroup.procs", cgroupPath)
    if err := os.WriteFile(file, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
        logrus.Fatalf("os.WriteFile err: %v", err)
    }

    return cgroupPath
}

func cleanupCgroup() {
    os.RemoveAll(CgroupPath)
}
