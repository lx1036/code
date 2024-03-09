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
// 现在，bpf 修改了，被搞坏了。
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    // 1. listen a server
    serverFd, err := makeServer(unix.SOCK_STREAM, "127.0.0.1", BindPort)
    if err != nil {
        logrus.Errorf("makeServer err: %v", err)
        return
    }
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
    logrus.Infof("client port %d", clientPort) // 5432, 验证出 client port 还是原来的

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

func makeServer(socketType int, ip string, port int) (int, error) {
    var err error
    var sockfd int
    defer func() {
        if err != nil && sockfd > 0 {
            unix.Close(sockfd)
        }
    }()

    sockfd, err = unix.Socket(unix.AF_INET, socketType, 0)
    if err != nil {
        logrus.Errorf("unix.Socket error: %v", err)
        return 0, err
    }
    setSocketTimeout(sockfd, 1000)
    // INFO: 注意 udp IP_RECVORIGDSTADDR 和 tcp SO_REUSEADDR
    if socketType == unix.SOCK_DGRAM {
        err = unix.SetsockoptInt(sockfd, unix.SOL_IP, unix.IP_RECVORIGDSTADDR, 1)
        if err != nil {
            logrus.Errorf("unix.IP_RECVORIGDSTADDR error: %v", err)
            return 0, err
        }
    }

    // ignores TIME-WAIT state using SO_REUSEADDR option
    // https://serverfault.com/questions/329845/how-to-forcibly-close-a-socket-in-time-wait
    err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Errorf("unix.SO_REUSEADDR error: %v", err)
        return 0, err
    }

    ipAddr := net.ParseIP(ip)
    sa := &unix.SockaddrInet4{
        Port: port,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ipAddr)
    err = unix.Bind(sockfd, sa)
    if err != nil {
        logrus.Errorf("unix.Bind error: %v", err)
        return 0, err
    }
    if socketType == unix.SOCK_STREAM {
        err = unix.Listen(sockfd, 128)
        if err != nil {
            logrus.Errorf("unix.Listen error: %v", err)
            return 0, err
        }
    }

    return sockfd, nil
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
