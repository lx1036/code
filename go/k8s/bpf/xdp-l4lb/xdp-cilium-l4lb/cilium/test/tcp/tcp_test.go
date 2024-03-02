package tcp

import (
    "encoding/hex"
    "fmt"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "strconv"
    "testing"
)

func TestHexDecode(test *testing.T) {
    b, err := hex.DecodeString("0xde")
    if err != nil {
        test.Fatal(err) // invalid byte: U+0078 'x'
    }
    u := uint8(b[0])
    fmt.Println(u)
}

func TestUint8(test *testing.T) {
    i := 0xde
    u := uint8(i)
    fmt.Println(u) // 222

    fmt.Println(strconv.Itoa(0x55))       // 85
    fmt.Println(strconv.Itoa(0x55AA * 2)) // 21930 * 2

    fmt.Println(int(^uint(0) >> 1)) // 9223372036854775807
}

func TestGetSocketOptType(test *testing.T) {
    serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        test.Fatal(err)
    }

    socketType, err := unix.GetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_TYPE)
    if err != nil {
        logrus.Fatal(err)
    }

    logrus.Infof("socketType: %d", socketType)
    if socketType == unix.SOCK_STREAM {
        logrus.Info("success")
    } else {
        logrus.Info("fail")
    }
}

// CGO_ENABLED=0 go test -v -run ^TestUdpEchoServer$ .
// udp echo server 打印成功:
// "server unix.Recvmsg from client: a"
// "client unix.Recvmsg from server: a"
func TestUdpEchoServer(test *testing.T) {
    // server
    serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
    if err != nil {
        logrus.Errorf("unix.Socket err: %v", err)
        return
    }
    defer unix.Close(serverFd)
    err = unix.SetsockoptInt(serverFd, unix.SOL_IP, unix.IP_RECVORIGDSTADDR, 1)
    if err != nil {
        logrus.Errorf("unix.SetsockoptInt err: %v", err)
        return
    }
    ipAddr := net.ParseIP("127.0.0.1")
    sa := &unix.SockaddrInet4{
        Port: 8001,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ipAddr)
    err = unix.Bind(serverFd, sa)
    if err != nil {
        logrus.Errorf("unix.Bind err: %v", err)
        return
    }

    // client
    clientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
    if err != nil {
        logrus.Errorf("unix.Socket err: %v", err)
        return
    }
    defer unix.Close(clientFd)
    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Errorf("unix.Getsockname err: %v", err)
        return
    }
    // client > server
    err = unix.Connect(clientFd, serverSockAddr)
    if err != nil {
        logrus.Errorf("unix.Connect err: %v", err)
        return
    }

    echoData := []byte("a")
    err = unix.Send(clientFd, echoData, 0)
    if err != nil {
        logrus.Errorf("unix.Send err: %v", err)
        return
    }
    cbuf := make([]byte, 1024)
    n, _, _, from, err := unix.Recvmsg(serverFd, cbuf, nil, 0)
    cbuf = cbuf[:n]
    logrus.Infof("server unix.Recvmsg from client: %s", string(cbuf))

    // 可以使用 unix.Sendto 直接 server > from(client)
    err = unix.Sendto(serverFd, cbuf, 0, from)
    if err != nil {
        logrus.Errorf("unix.Sendto err: %v", err)
        return
    }
    cbuf2 := make([]byte, 1024)
    n2, _, _, _, err := unix.Recvmsg(clientFd, cbuf2, nil, 0)
    if err != nil {
        logrus.Errorf("unix.Recvmsg clientFd err: %v", err)
        return
    }
    cbuf2 = cbuf2[:n2]
    logrus.Infof("client unix.Recvmsg from server: %s", string(cbuf2))
}

// CGO_ENABLED=0 go test -v -run ^TestUdpBind$ .
func TestUdpBind(test *testing.T) {
    logrus.SetReportCaller(true)

    for i := 0; i < 10; i++ {
        sockfds := makeServer()
        for i := 0; i < len(sockfds); i++ {
            unix.Close(sockfds[i])
        }
    }
}

func makeServer() []int {
    var err error
    sockfds := make([]int, 10)
    for i := 0; i < len(sockfds); i++ {
        sockfds[i], err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
        if err != nil {
            logrus.Errorf("unix.Socket error: %v", err)
            continue
        }

        err = unix.SetsockoptInt(sockfds[i], unix.SOL_IP, unix.IP_RECVORIGDSTADDR, 1)
        if err != nil {
            logrus.Errorf("unix.IP_RECVORIGDSTADDR error: %v", err)
            continue
        }

        err = unix.SetsockoptInt(sockfds[i], unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
        if err != nil {
            logrus.Errorf("unix.SO_REUSEADDR error: %v", err)
            continue
        }

        // unix.SO_REUSEADDR 可以 bind 多次
        ipAddr := net.ParseIP("127.0.0.1")
        sa := &unix.SockaddrInet4{
            Port: 8081,
            Addr: [4]byte{},
        }
        copy(sa.Addr[:], ipAddr)
        err = unix.Bind(sockfds[i], sa)
        if err != nil {
            logrus.Errorf("unix.Bind error: %v", err)
            continue
        }
    }

    return sockfds
}
