//go:build linux
// +build linux

package iptables

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// 开关 linux 内核容许 ipv4 forward 数据包转发
const (
	ipv4ForwardConf     = "/proc/sys/net/ipv4/ip_forward"
	ipv4ForwardConfPerm = 0644
)

func SetupIPForward() error {
	ipv4ForwardData, err := ioutil.ReadFile(ipv4ForwardConf)
	if err != nil {
		return fmt.Errorf("can't read ip forwarding from %s: %v", ipv4ForwardConf, err)
	}
	if ipv4ForwardData[0] != '1' {
		if err := configureIPv4Forwarding(true); err != nil {
			return fmt.Errorf("fail to enable ip forwarding: %v", err)
		}

		if err := SetPolicy(Filter, Forward, Drop); err != nil {
			if err := configureIPv4Forwarding(false); err != nil {
				return fmt.Errorf("fail to disable ip forwarding: %v", err)
			}

			return fmt.Errorf("fail to set %s/%s policy: %v", Filter, Forward, err)
		}
	}

	return nil
}

func configureIPv4Forwarding(enable bool) error {
	var val byte = '0'
	if enable {
		val = '1'
	}
	return ioutil.WriteFile(ipv4ForwardConf, []byte{val, '\n'}, ipv4ForwardConfPerm)
}

// cat /proc/sys/net/ipv4/conf/docker0/route_localnet
// Setup Loopback Addresses Routing
func SetupLoopbackAddressesRouting(bridgeName string) error {
	sysPath := filepath.Join("/proc/sys/net/ipv4/conf", bridgeName, "route_localnet")
	ipv4LoRoutingData, err := ioutil.ReadFile(sysPath)
	if err != nil {
		return fmt.Errorf("can't read IPv4 local routing setup: %v", err)
	}
	// Enable loopback addresses routing only if it isn't already enabled
	if ipv4LoRoutingData[0] != '1' {
		if err := ioutil.WriteFile(sysPath, []byte{'1', '\n'}, 0644); err != nil {
			return fmt.Errorf("can't enable local routing for hairpin mode: %v", err)
		}
	}
	return nil
}
