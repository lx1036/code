package tcp

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"log"
	"net"
)

func startServer() int {
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

	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
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

func connectToFd() int {
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

	raddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
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

func fastOpenConnect() int {
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

	data := []byte("FAST!!!")
	unix.Sendto(csk, data, unix.MSG_FASTOPEN, &unix.SockaddrInet4{Port: 8080})

	return csk
}

func startSocketServer() {
	sk := startServer()

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

func getAddr(addr *unix.SockaddrInet4) (string, int) {
	cAddr := addr.Addr
	cPort := addr.Port
	ipv4 := net.IPv4(cAddr[0], cAddr[1], cAddr[2], cAddr[3])

	return ipv4.String(), cPort
}

func handleConnection(clientSocket int) {
	var data []byte
	n, err := unix.Read(clientSocket, data)
	if err != nil {
		log.Printf("read failed: %v\n", err)
		return
	}

	logrus.Infof("read %d number bytes from client", n)

	// unix.Write(clientSocket, data[:n])
}
