package tcp

import (
    "encoding/hex"
    "fmt"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
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
    fmt.Println(u)
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
