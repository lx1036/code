package internal

import (
	"fmt"
	"net"
	"os/exec"
	"testing"
)

func TestLookPath(test *testing.T) {
	path, err := exec.LookPath("kubectl")
	if err != nil {
		panic(err)
	}

	// /usr/local/bin/kubectl
	fmt.Println(path)
}

const (
	chainName  = "lx1036"
	bridgeName = "lo"
)

func TestNewChain(test *testing.T) {
	natChain, err := NewChain(chainName, Nat, false)
	if err != nil {
		test.Error(err)
	}
	err = Rule(natChain, bridgeName, false, true)
	if err != nil {
		test.Error(err)
	}

	filterChain, err := NewChain(chainName, Filter, false)
	if err != nil {
		test.Error(err)
	}
	err = Rule(filterChain, bridgeName, false, true)
	if err != nil {
		test.Error(err)
	}
}

func TestForward(test *testing.T) {
	natChain, err := NewChain(chainName, Nat, false)
	if err != nil {
		test.Error(err)
	}

	ip := net.ParseIP("192.168.1.1")
	port := 1234
	dstAddr := "172.17.0.1"
	dstPort := 4321
	protocal := "TCP"
	err = natChain.Forward(Insert, protocal, bridgeName, ip, port, dstAddr, dstPort)
	if err != nil {
		test.Error(err)
	}
}
