//go:build linux
// +build linux

package iptables

import "fmt"

var (
	// apt install -y firewalld && systemctl status firewalld && ls /lib/systemd/system/firewalld.service
	firewallRunning bool // if firewalld service is running
)

// 不考虑防火墙程序 firewalld
func FirewallInit() error {
	return fmt.Errorf("firewall error")
}
