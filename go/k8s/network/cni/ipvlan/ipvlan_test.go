package main

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/plugins/pkg/testutils"
	"k8s.io/klog/v2"
	"testing"
)

func TestIPVlan(test *testing.T) {
	targetNs, _ := testutils.NewNS()
	defer targetNs.Close()
	ifName := "eth0"
	conf := fmt.Sprintf(`{
			    "cniVersion": "0.3.0",
			    "name": "mynet",
			    "type": "ipvlan",
			    "master": "eth0",
			    "ipam": {
					"type": "host-local",
					"subnet": "100.1.2.0/24",
			    },
				"prevResult": {
					"cniVersion": "0.3.0",
					"interfaces": [
						{
							"name": "%s",
							"sandbox": "%s"
						}
					],
					"ips": [
						{
							"version": "4",
							"address": "192.168.1.209/24",
							"gateway": "192.168.1.1",
							"interface": 0
						}
					],
					"routes": []
				}
			}`, ifName, targetNs.Path())
	args := &skel.CmdArgs{
		ContainerID: "dummy",
		Netns:       targetNs.Path(),
		IfName:      ifName,
		StdinData:   []byte(conf),
	}
	_, _, err := testutils.CmdAddWithArgs(args, func() error { return cmdAdd(args) })
	if err != nil {
		klog.Error(err)
	}
}
