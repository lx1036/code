package main

import "k8s-lx1036/k8s-ui/backend/kubernetes/service/network/iptables/iptables"

func main() {
	err := iptables.SetupIPForward()
	if err != nil {
		panic(err)
	}
}
