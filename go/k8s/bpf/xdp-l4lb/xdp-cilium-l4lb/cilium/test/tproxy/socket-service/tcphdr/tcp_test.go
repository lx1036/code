package main

import (
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "testing"
    "time"
)

// 使用 go 来测试 tcp fastopen，但是必须在 linux 环境里才可以，本地 mac 不行。
// 目前还没有调试成功

// go test -v -run ^TestFastOpen$ .
func TestFastOpen(test *testing.T) {
    expireCh := time.After(time.Second * 5)

    go startSocketServer("tcp", ":6080") // nginx: listen 6080 fastopen=256;

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
    clientSk := fastOpenConnect("tcp", ":6080")
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
