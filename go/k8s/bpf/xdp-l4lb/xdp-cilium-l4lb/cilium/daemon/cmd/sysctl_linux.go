package cmd

import (
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/sysctl"
)

func enableIPForwarding() error {
	if err := sysctl.Enable("net.ipv4.ip_forward"); err != nil {
		return err
	}
	if err := sysctl.Enable("net.ipv4.conf.all.forwarding"); err != nil {
		return err
	}

	return nil
}
