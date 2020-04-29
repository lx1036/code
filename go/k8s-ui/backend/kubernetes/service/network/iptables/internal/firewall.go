package internal

import "fmt"

var (
	// apt install -y firewalld && systemctl status firewalld && ls /lib/systemd/system/firewalld.service
	firewallRunning bool // if firewalld service is running
)

func FirewallInit() error {

	return fmt.Errorf("firewall error")
}
