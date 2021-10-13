//go:build linux
// +build linux

package iptables

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

// 用户自定义一个 chain
func TestNewChain(test *testing.T) {
	// add 'lx1036' chain into nat table
	natChain, err := NewChain(chainName, Nat, false)
	if err != nil {
		test.Error(err)
	}
	err = Rule(natChain, bridgeName, false, true)
	if err != nil {
		test.Error(err)
	}

	// add 'lx1036' chain into filter table
	filterChain, err := NewChain(chainName, Filter, false)
	if err != nil {
		test.Error(err)
	}
	err = Rule(filterChain, bridgeName, false, true)
	if err != nil {
		test.Error(err)
	}
}

// prerouting chain: 数据包在路由决策之前
func TestPrerouting(test *testing.T) {

}

// output chain: 数据包在被本地进程处理之后
func TestOutput(test *testing.T) {

}

// forward chain: 数据包准发
// nat: -A POSTROUTING -s 172.17.0.2/32 -d 172.17.0.2/32 -p tcp -m tcp --dport 80 -j MASQUERADE
// nat: -A DOCKER ! -i docker0 -p tcp -m tcp --dport 8800 -j DNAT --to-destination 172.17.0.2:80
// filter: -A DOCKER -d 172.17.0.2/32 ! -i docker0 -o docker0 -p tcp -m tcp --dport 80 -j ACCEPT
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
	// "tcp://192.168.1.1:1234" -> "tcp://172.17.0.1:4321"
	err = natChain.Forward(Insert, protocal, bridgeName, ip, port, dstAddr, dstPort)
	if err != nil {
		test.Error(err)
	}
}

func TestLink(test *testing.T) {

}

func TestCleanup(test *testing.T) {

}
