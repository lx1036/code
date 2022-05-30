package main

import (
	"fmt"
	"testing"

	"github.com/moby/ipvs"
	"k8s.io/klog/v2"
)

func TestIPVS(test *testing.T) {
	handle, err := ipvs.New("")
	if err != nil {
		klog.Fatalf("ipvs.New: %s", err)
	}
	svcs, err := handle.GetServices()
	if err != nil {
		klog.Fatalf("handle.GetServices: %s", err)
	}

	for _, svc := range svcs {
		fmt.Println(svc)
	}
}
