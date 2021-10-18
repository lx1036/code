//go:build linux
// +build linux

package k8s

import (
	"fmt"
	"testing"
)

const (
	// FlagPersistent specify IPVS service session affinity
	FlagPersistent = 0x1
	// FlagHashed specify IPVS service hash flag
	FlagHashed = 0x2
	// IPVSProxyMode is match set up cluster with ipvs proxy model
	IPVSProxyMode = "ipvs"
)

func TestGetVirtualServers(test *testing.T) {
	runner, err := New()
	if err != nil {
		test.Error(err)
	}

	virtualServers, err := runner.GetVirtualServers()
	if err != nil {
		test.Error(err)
	}

	for _, virtualServer := range virtualServers {
		fmt.Println(virtualServer)
	}
}

func TestEqual(test *testing.T) {

}
