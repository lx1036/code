package main

import (
	"github.com/containernetworking/cni/pkg/skel"
	"testing"
)

func TestCNI(test *testing.T) {

	stdinData := `{ "name":"cilium", "type": "cilium-cni", "cniVersion": "0.3.1", "enable-debug": "false" }`
	expectedCmdArgs := &skel.CmdArgs{
		ContainerID: "abc123",
		Netns:       "/proc/3306/ns/net",
		IfName:      "eth0",
		Args:        "IgnoreUnknown=1;K8S_POD_NAMESPACE=default;K8S_POD_NAME=nginx;K8S_POD_INFRA_CONTAINER_ID=abc123",
		Path:        "/some/cni/path",
		StdinData:   []byte(stdinData),
	}

	parseCmdArgs(expectedCmdArgs)
}
