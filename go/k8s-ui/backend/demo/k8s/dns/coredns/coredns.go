package main

import (
	"k8s-lx1036/k8s-ui/backend/demo/k8s/dns/coredns/coremain"
	_ "k8s-lx1036/k8s-ui/backend/demo/k8s/dns/coredns/plugin"
)

func main() {
	coremain.Run()
}