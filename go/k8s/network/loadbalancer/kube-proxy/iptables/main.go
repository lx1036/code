package main

import (
	"k8s-lx1036/k8s/network/network-policy/iptables/iptables"
)

func main() {
	err := iptables.SetupIPForward()
	if err != nil {
		panic(err)
	}
}
