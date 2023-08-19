package tcp

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
	stopCh := time.After(time.Second * 5)

	go startSocketServer()

	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			clientSendto()
		case <-stopCh:
			return
		}
	}
}

func clientSendto() {
	clientSk := fastOpenConnect()
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

//func fastOpenConnect() int {
//	sk, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, 0)
//	if err != nil {
//		logrus.Errorf("%+v", err)
//		return sk
//	}
//	var optVal = uint32(1)
//	err = unix.SetsockoptInt(sk, unix.IPPROTO_TCP, unix.TCP_FASTOPEN, int(optVal))
//	if err != nil {
//		logrus.Errorf("%+v", err)
//		return sk
//	}
//	var addr unix.SockaddrInet4
//	addr.Family = unix.AF_INET
//	addr.Port = 9000
//	addr.Addr[0] = byte(172)
//	addr.Addr[1] = byte(16)
//	addr.Addr[2] = byte(1)
//	addr.Addr[3] = byte(1)
//	err = unix.Connect(sk, &addr)
//	if err == unix.EINPROGRESS || err == unix.EAL
//
//}
