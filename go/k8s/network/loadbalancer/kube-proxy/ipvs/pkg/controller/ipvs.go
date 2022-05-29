package controller

import (
	"errors"
	"fmt"
	"github.com/moby/ipvs"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
	"net"
	"os/exec"
	"strings"
	"sync"
)

// INFO: @see https://github.com/kubernetes/kubernetes/blob/master/pkg/util/ipvs/ipvs_linux.go
//  https://github.com/cloudnativelabs/kube-router/blob/master/pkg/controllers/proxy/network_services_controller.go#L131-L140

const (
	KubeDummyIf = "kube-dummy-if"
)

type linuxNetworking struct {
	sync.Mutex

	ipvsHandle *ipvs.Handle
}

func newLinuxNetworking() (*linuxNetworking, error) {
	ln := &linuxNetworking{}
	ipvsHandle, err := ipvs.New("")
	if err != nil {
		return nil, err
	}

	ln.ipvsHandle = ipvsHandle
	return ln, nil
}

func (ln *linuxNetworking) EnsureDummyDevice() (netlink.Link, error) {
	dummyVipInterface, err := netlink.LinkByName(KubeDummyIf)
	if err != nil && errors.Is(err, netlink.LinkNotFoundError{}) {
		klog.Infof("Could not find dummy interface: %s to assign cluster ip, creating one", KubeDummyIf)
		err = netlink.LinkAdd(&netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: KubeDummyIf}})
		if err != nil {
			return nil, errors.New("Failed to add dummy interface:  " + err.Error())
		}
		dummyVipInterface, err = netlink.LinkByName(KubeDummyIf)
		if err != nil {
			return nil, errors.New("Failed to get dummy interface: " + err.Error())
		}
		err = netlink.LinkSetUp(dummyVipInterface)
		if err != nil {
			return nil, errors.New("Failed to bring dummy interface up: " + err.Error())
		}
	}

	return dummyVipInterface, nil
}

func (ln *linuxNetworking) EnsureAddressBind(link netlink.Link, ip string, addRoute bool) error {
	err := netlink.AddrAdd(link, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP(ip),
			Mask: net.IPv4Mask(255, 255, 255, 255), // net.CIDRMask(32, 32),
		},
		Scope: unix.RT_SCOPE_LINK,
	})
	if err != nil && err != unix.EEXIST { // "EEXIST" will be returned if the address is already bound to device
		return err
	}
	if err == unix.EEXIST { // 已经绑定过 ip 了，无需再次添加路由
		return nil
	}

	if !addRoute {
		return nil
	}

	// When a service VIP is assigned to a dummy interface and accessed from host, in some of the
	// case Linux source IP selection logix selects VIP itself as source leading to problems
	// to avoid this an explicit entry is added to use node IP as source IP when accessing
	// VIP from the host. Please see https://github.com/cloudnativelabs/kube-router/issues/376

	// INFO: netlink.RouteReplace which is replacement for below command is not working as expected. Call succeeds but
	//  route is not replaced. For now do it with command. 解释了 "src NodeIP"，且路由在 local table 里。
	//  `ip route replace local ipxxx dev kube-dummy-if table local proto kernel scope host src NodeIP table local`
	out, err := exec.Command("ip", "route", "replace", "local", ip, "dev", KubeDummyIf,
		"table", "local", "proto", "kernel", "scope", "host", "src",
		NodeIP.String(), "table", "local").CombinedOutput()
	if err != nil {
		klog.Errorf("Failed to replace route to service VIP %s configured on %s. Error: %v, Output: %s",
			ip, KubeDummyIf, err, out)
	}
	return nil
}

func (ln *linuxNetworking) AddOrUpdateVirtualServer(ipvsSvc *ipvs.Service) error {
	ln.Lock()
	defer ln.Unlock()

	oldIpvsSvc, _ := ln.ipvsHandle.GetService(ipvsSvc)
	if oldIpvsSvc == nil || !equalIPVSService(oldIpvsSvc, ipvsSvc) {
		if oldIpvsSvc == nil {
			if err := ln.AddVirtualServer(ipvsSvc); err != nil {
				klog.Errorf(fmt.Sprintf("Add new ipvs service for virtual server err: %v", err))
				return err
			}
		} else {
			if err := ln.UpdateVirtualServer(ipvsSvc); err != nil {
				klog.Errorf(fmt.Sprintf("Edit existed ipvs service for virtual server err: %v", err))
				return err
			}
		}
	}

	return nil
}

func (ln *linuxNetworking) GetVirtualServer(ipvsSvc *ipvs.Service) (*ipvs.Service, error) {
	ln.Lock()
	defer ln.Unlock()

	oldIpvsSvc, err := ln.ipvsHandle.GetService(ipvsSvc)
	if err != nil {
		return nil, err
	}
	return oldIpvsSvc, nil
}

// AddVirtualServer `ipvsadm --add-service xxx`
func (ln *linuxNetworking) AddVirtualServer(ipvsSvc *ipvs.Service) error {
	ln.Lock()
	defer ln.Unlock()

	return ln.ipvsHandle.NewService(ipvsSvc)
}

// UpdateVirtualServer `ipvsadm --edit-service xxx`
func (ln *linuxNetworking) UpdateVirtualServer(ipvsSvc *ipvs.Service) error {
	ln.Lock()
	defer ln.Unlock()

	return ln.ipvsHandle.UpdateService(ipvsSvc)
}

func (ln *linuxNetworking) AddRealServer(ipvsSvc *ipvs.Service, dst *ipvs.Destination) error {
	ln.Lock()
	defer ln.Unlock()

	return ln.ipvsHandle.NewDestination(ipvsSvc, dst)
}

func (ln *linuxNetworking) ListRealServer(ipvsSvc *ipvs.Service) ([]*ipvs.Destination, error) {
	ln.Lock()
	defer ln.Unlock()

	return ln.ipvsHandle.GetDestinations(ipvsSvc)
}

func (ln *linuxNetworking) DelRealServer(ipvsSvc *ipvs.Service, dst *ipvs.Destination) error {
	ln.Lock()
	defer ln.Unlock()

	return ln.ipvsHandle.DelDestination(ipvsSvc, dst)
}

func clusterIPToIPVSService(svcInfo serviceInfo) *ipvs.Service {
	ipvsSvc := &ipvs.Service{
		Address:   svcInfo.address, // clusterIP
		Protocol:  stringToProtocol(svcInfo.protocol),
		Port:      uint16(svcInfo.port), // clusterIP
		SchedName: svcInfo.scheduler,
		Flags:     svcInfo.flags,
		Timeout:   svcInfo.sessionAffinityTimeoutSeconds,
	}

	if ip4 := svcInfo.address.To4(); ip4 != nil {
		ipvsSvc.AddressFamily = unix.AF_INET
		ipvsSvc.Netmask = 0xffffffff
	} else {
		ipvsSvc.AddressFamily = unix.AF_INET6
		ipvsSvc.Netmask = 128
	}

	return ipvsSvc
}

func nodePortToIPVSService(svcInfo serviceInfo) *ipvs.Service {
	ipvsSvc := &ipvs.Service{
		Address:   NodeIP, // nodePort
		Protocol:  stringToProtocol(svcInfo.protocol),
		Port:      uint16(svcInfo.nodePort), // nodePort
		SchedName: svcInfo.scheduler,
		Flags:     svcInfo.flags,
		Timeout:   svcInfo.sessionAffinityTimeoutSeconds,
	}

	if ip4 := svcInfo.address.To4(); ip4 != nil {
		ipvsSvc.AddressFamily = unix.AF_INET
		ipvsSvc.Netmask = 0xffffffff
	} else {
		ipvsSvc.AddressFamily = unix.AF_INET6
		ipvsSvc.Netmask = 128
	}

	return ipvsSvc
}

func equalIPVSService(newIpvsSvc, ipvsSvc *ipvs.Service) bool {
	return newIpvsSvc.Address.Equal(ipvsSvc.Address) &&
		newIpvsSvc.Protocol == ipvsSvc.Protocol &&
		newIpvsSvc.Port == ipvsSvc.Port &&
		newIpvsSvc.SchedName == ipvsSvc.SchedName &&
		newIpvsSvc.Flags == ipvsSvc.Flags &&
		newIpvsSvc.Timeout == ipvsSvc.Timeout
}

// stringToProtocolType returns the protocol type for the given name
func stringToProtocol(protocol string) uint16 {
	switch strings.ToLower(protocol) {
	case "tcp":
		return uint16(unix.IPPROTO_TCP)
	case "udp":
		return uint16(unix.IPPROTO_UDP)
	case "sctp":
		return uint16(unix.IPPROTO_SCTP)
	}
	return uint16(0)
}
