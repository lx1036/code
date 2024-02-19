package main

import (
    "github.com/cilium/ebpf"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
)

const (
    MapPinFile  = "/sys/fs/bpf/tc/globals/server_map"
    BindPort    = 1234
    ConnectPort = 4321
)

// 貌似验证没问题，socket redirect 成功!!!
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    // 1. listen a server
    serverFd := startServer()
    defer unix.Close(serverFd)

    fileName := MapPinFile
    serverMap, err := ebpf.LoadPinnedMap(fileName, nil)
    if err != nil {
        logrus.Fatalf("LoadPinnedMap err: %v", err)
    }
    defer serverMap.Close()

    key := uint32(0)
    err = serverMap.Put(key, uint64(serverFd))
    if err != nil {
        logrus.Fatalf("Put err: %v", err)
    }

    // 2. client connect server, server accept
    clientFd := connectToFd()
    defer unix.Close(clientFd)
    clientServerFd, clientSockAddr, err := unix.Accept(serverFd)
    if err != nil {
        logrus.Fatalf("Accept err: %v", err)
    }
    clientPort := clientSockAddr.(*unix.SockaddrInet4).Port
    logrus.Infof("client port %d", clientPort) // 5432

    // 3. client write, server read
    if _, err = unix.Write(clientFd, []byte("testing")); err != nil {
        logrus.Fatalf("unix.Write err: %v", err)
    }
    cbuf := make([]byte, unix.CmsgSpace(4))
    n, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf, nil, 0)
    if err != nil {
        logrus.Fatalf("unix.Recvmsg err: %v", err)
    }
    cbuf = cbuf[:n]
    logrus.Infof("%s", string(cbuf)) // testing

    // 4. get server info
    serverSockAddr, err := unix.Getsockname(clientServerFd)
    if err != nil {
        logrus.Fatalf("unix.Getsockname clientServerFd err: %v", err)
    }
    serverPort := serverSockAddr.(*unix.SockaddrInet4).Port
    logrus.Infof("server port %d", serverPort)
}

func startServer() int {
    var err error
    var fd int
    defer func() {
        if err != nil && fd > 0 {
            unix.Close(fd)
        }
    }()

    fd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }

    ip := net.ParseIP("127.0.0.1")
    sa := &unix.SockaddrInet4{
        Port: BindPort,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(fd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.Listen(fd, 1)
    if err != nil {
        logrus.Fatal(err)
    }

    return fd
}

func connectToFd() int {
    var err error
    var clientFd int
    defer func() {
        if err != nil && clientFd > 0 {
            unix.Close(clientFd)
        }
    }()

    clientFd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }

    // INFO: 这里客户端指定 ip:port
    ip1 := net.ParseIP("127.0.0.1")
    sa1 := &unix.SockaddrInet4{
        Port: 5432,
        Addr: [4]byte{},
    }
    copy(sa1.Addr[:], ip1)
    err = unix.Bind(clientFd, sa1)
    if err != nil {
        logrus.Fatal(err)
    }

    ip := net.ParseIP("127.0.0.1")
    sa := &unix.SockaddrInet4{
        Port: ConnectPort,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    // 非阻塞的
    err = unix.Connect(clientFd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    return clientFd
}
