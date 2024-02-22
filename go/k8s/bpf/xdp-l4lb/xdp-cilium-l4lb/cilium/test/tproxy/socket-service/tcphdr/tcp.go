package main

import (
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "log"
    "net"
)

func startServer(network, serverAddr string) int {
    var sk int
    var err error
    defer func() {
        if err != nil {
            logrus.Errorf("start server err:%v", err)
            unix.Close(sk)
        }
    }()
    sk, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        return -1
    }
    err = unix.SetsockoptInt(sk, unix.SOL_TCP, unix.TCP_FASTOPEN, 256)
    if err != nil {
        return -1
    }

    laddr, err := net.ResolveTCPAddr(network, serverAddr)
    // laddr, err := net.ResolveTCPAddr("tcp", "[::]:8080")
    if err != nil {
        logrus.Errorf("Error resolving local address: %v", err)
        return -1
    }

    // 报错 "transport endpoint is not connected"
    // unix.SetsockoptInt(sk, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    lsockaddr := &unix.SockaddrInet4{
        Port: laddr.Port,
        Addr: [4]byte{},
    }
    copy(lsockaddr.Addr[:], laddr.IP.To4())
    err = unix.Bind(sk, lsockaddr)
    if err != nil {
        return -1
    }

    err = unix.Listen(sk, 5)
    if err != nil {
        return -1
    }

    return sk
}

func startSocketServer(network, serverAddr string) {
    sk := startServer(network, serverAddr)

    for {
        clientSocket, clientAddr, err := unix.Accept(sk)
        if err != nil {
            log.Printf("accept failed: %v\n", err)
            continue
        }
        cAddrStr, cPort := getAddr(clientAddr.(*unix.SockaddrInet4))
        logrus.Infof("accept a new connection from client %s:%d", cAddrStr, cPort)

        go handleConnection(clientSocket)
    }
}

func handleConnection(clientSocket int) {
    data := make([]byte, 1024)
    // INFO: Recvfrom() / Read() 都可以, 因为已经 establish
    // n, _, err := unix.Recvfrom(clientSocket, data, 0)
    n, err := unix.Read(clientSocket, data)
    if err != nil {
        log.Printf("read failed: %v\n", err)
        return
    }
    logrus.Infof("read %d number bytes from client: %s", n, string(data[:n]))

    data = []byte("hello client!")
    n, err = unix.Write(clientSocket, data)
    if err != nil {
        log.Printf("write failed: %v\n", err)
        return
    }
    logrus.Infof("write %d number bytes to client", n)
}

func connectToFd(network, serverAddr string) int {
    var csk int
    var err error
    defer func() {
        if err != nil {
            logrus.Error(err)
            unix.Close(csk)
        }
    }()

    csk, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        return -1
    }

    raddr, err := net.ResolveTCPAddr(network, serverAddr)
    if err != nil {
        logrus.Errorf("Error resolving local address: %v", err)
        return -1
    }
    rsockaddr := &unix.SockaddrInet4{
        Port: raddr.Port,
        Addr: [4]byte{},
    }
    copy(rsockaddr.Addr[:], raddr.IP.To4())
    err = unix.Connect(csk, rsockaddr)

    return csk
}

func fastOpenConnect(network, serverAddr string) int {
    var csk int
    var err error
    defer func() {
        if err != nil {
            logrus.Errorf("connect %v", err)
            unix.Close(csk)
        }
    }()

    csk, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        return -1
    }
    err = unix.SetsockoptInt(csk, unix.SOL_TCP, unix.TCP_FASTOPEN, 256)
    if err != nil {
        return -1
    }

    raddr, err := net.ResolveTCPAddr(network, serverAddr)
    rsockAddr := &unix.SockaddrInet4{
        Port: raddr.Port,
        Addr: [4]byte{},
    }
    copy(rsockAddr.Addr[:], raddr.IP.To4())
    data := []byte("hello server!")
    unix.Sendto(csk, data, unix.MSG_FASTOPEN, rsockAddr)
    // INFO: Write() 不行，使用 Write() 必须要先 Connect()
    //unix.Write(csk, data)

    return csk
}

func getAddr(addr *unix.SockaddrInet4) (string, int) {
    cAddr := addr.Addr
    cPort := addr.Port
    ipv4 := net.IPv4(cAddr[0], cAddr[1], cAddr[2], cAddr[3])

    return ipv4.String(), cPort
}
