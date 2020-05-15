package netlink

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"runtime"
	"testing"
)

// [笔记]《k8s网络权威指南》1.3小节：Linux Bridge
// validate: `brctl show`
func TestAddBridgeWithEth(test *testing.T) {
	attrs := netlink.NewLinkAttrs()
	attrs.Name = "br100"
	// $link= name lx1036 type bridge
	bridge := &netlink.Bridge{
		LinkAttrs: attrs,
	}
	// `ip link add $link`
	err := netlink.LinkAdd(bridge)
	if err != nil  {
		panic(err)
	}
	eth1, err := netlink.LinkByName("eth0")
	if err != nil {
		panic(err)
	}
	// `ip link set $link master $master`
	err = netlink.LinkSetMaster(eth1, bridge)
	if err != nil {
		panic(err)
	}
}

// [笔记]《k8s网络权威指南》1.1小节：crud network namespace
// go test -v -run ^TestNetworkNamespace$ link_test.go
func TestNetworkNamespace(test *testing.T) {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save the current network namespace
	origns, _ := netns.Get()
	defer origns.Close()
	// Create a new network namespace
	newns, _ := netns.New()
	defer newns.Close()
	// Do something with the network namespace
	interfaces, _ := net.Interfaces()
	fmt.Printf("Interfaces: %v\n", interfaces)

	// Switch back to the original namespace
	fmt.Println(netns.Set(origns))
}
