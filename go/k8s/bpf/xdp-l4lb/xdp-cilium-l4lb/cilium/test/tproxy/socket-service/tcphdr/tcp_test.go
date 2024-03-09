package main

import (
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "testing"
    "time"
)

// 使用 go 来测试 tcp fastopen，但是必须在 linux 环境里才可以，本地 mac 不行。
// 目前还没有调试成功

// go test -v -run ^TestFastOpen$ .
func TestFastOpen(test *testing.T) {
    expireCh := time.After(time.Second * 5)

    go startSocketServer("tcp", "127.0.0.1:6080") // nginx: listen 6080 fastopen=256;

    tick := time.NewTicker(time.Second)
    defer tick.Stop()
    for {
        select {
        case <-tick.C:
            clientSendto()
        case <-expireCh:
            return
        }
    }
}

func clientSendto() {
    clientSk := FastOpenConnect("tcp", "127.0.0.1:6080")
    defer unix.Close(clientSk)

    data := make([]byte, 1024)
    //unix.Read()
    n, _, err := unix.Recvfrom(clientSk, data, 0)
    if err != nil {
        logrus.Error(err)
        return
    }
    logrus.Infof("recv %d number bytes from server: %s", n, string(data[:n]))

    clientAddr, err := unix.Getsockname(clientSk)
    if err != nil {
        logrus.Error(err)
        return
    }
    cAddrStr, cPort := getAddr(clientAddr.(*unix.SockaddrInet4))
    logrus.Infof("client addr %s:%d", cAddrStr, cPort)

    serverAddr, err := unix.Getpeername(clientSk)
    if err != nil {
        logrus.Error(err)
        return
    }
    sAddrStr, sPort := getAddr(serverAddr.(*unix.SockaddrInet4))
    logrus.Infof("server addr %s:%d", sAddrStr, sPort)
}

func FastOpenConnect(network, serverAddr string) int {
    clientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

    // INFO: 开启该 socket option 支持 client 可以发送 TFO syn
    err = unix.SetsockoptInt(clientFd, unix.SOL_TCP, unix.TCP_FASTOPEN, 256)
    if err != nil {
        logrus.Fatal(err)
    }

    serverAddr2, err := net.ResolveTCPAddr(network, serverAddr)
    if err != nil {
        logrus.Errorf("Error resolving server address: %v", err)
        return -1
    }
    serverSockaddr := &unix.SockaddrInet4{
        Port: serverAddr2.Port,
        Addr: [4]byte{},
    }
    copy(serverSockaddr.Addr[:], serverAddr2.IP.To4())
    err = unix.Connect(clientFd, serverSockaddr)
    if err != nil {
        return -1
    }

    msg := []byte("hello server!")
    err = unix.Sendto(clientFd, msg, unix.MSG_FASTOPEN, serverSockaddr)
    if err != nil {
        logrus.Fatal(err)
    }

    return clientFd
}

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
            logrus.Errorf("accept failed: %v\n", err)
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
        logrus.Errorf("read failed: %v\n", err)
        return
    }
    logrus.Infof("read %d number bytes from client: %s", n, string(data[:n]))

    data = []byte("hello client!")
    n, err = unix.Write(clientSocket, data)
    if err != nil {
        logrus.Errorf("write failed: %v\n", err)
        return
    }
    logrus.Infof("write %d number bytes to client", n)
}

func getAddr(addr *unix.SockaddrInet4) (string, int) {
    cAddr := addr.Addr
    cPort := addr.Port
    ipv4 := net.IPv4(cAddr[0], cAddr[1], cAddr[2], cAddr[3])

    return ipv4.String(), cPort
}
