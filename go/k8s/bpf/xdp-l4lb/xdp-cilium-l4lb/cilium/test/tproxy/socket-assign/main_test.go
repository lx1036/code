package main

import (
    "github.com/cilium/ebpf"
    "github.com/sirupsen/logrus"
    "testing"
)

func TestUpdateMap(test *testing.T) {
    fileName := MapPinFile
    serverMap, err := ebpf.LoadPinnedMap(fileName, nil)
    if err != nil {
        logrus.Fatalf("LoadPinnedMap err: %v", err)
    }
    defer serverMap.Close()

    key := uint32(0)
    serverFd := 12
    err = serverMap.Put(key, uint64(serverFd))
    if err != nil {
        logrus.Fatalf("Put err: %v", err)
    }
}

// CGO_ENABLED=0 go test -v -run ^TestByte$ .
func TestByte(test *testing.T) {
    data := [8]byte{}
    copy(data[:], "00010001")
    logrus.Infof("data: %s", data)

    data2 := make([]byte, 1024)
    copy(data2[:], data[:])

    logrus.Infof("data2: %s", data2)
}
