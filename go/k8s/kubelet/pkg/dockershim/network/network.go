package network

import (
	"fmt"
	"net"
	"strings"

	utilexec "k8s.io/utils/exec"
)

// TODO: Consider making this value configurable.
const DefaultInterfaceName = "eth0"

// INFO: 这个函数很基础很重要的函数
// getOnePodIP 使用 nsenter 命令进入 pid net namespace 获取 pod ip:
// nsenter --target=${pid} --net -F -- ip -o -4 addr show dev eth0 scope global
func getOnePodIP(execer utilexec.Interface, nsenterPath, netnsPath, interfaceName, addrType string) (net.IP, error) {
	// Try to retrieve ip inside container network namespace
	output, err := execer.Command(nsenterPath, fmt.Sprintf("--net=%s", netnsPath), "-F", "--",
		"ip", "-o", addrType, "addr", "show", "dev", interfaceName, "scope", "global").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("unexpected command output %s with error: %v", output, err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("unexpected command output %s", output)
	}
	fields := strings.Fields(lines[0])
	if len(fields) < 4 {
		return nil, fmt.Errorf("unexpected address output %s ", lines[0])
	}
	ip, _, err := net.ParseCIDR(fields[3])
	if err != nil {
		return nil, fmt.Errorf("CNI failed to parse ip from output %s due to %v", output, err)
	}

	return ip, nil
}

// GetPodIP gets the IP of the pod by inspecting the network info inside the pod's network namespace.
// TODO (khenidak). The "primary ip" in dual stack world does not really exist. For now
// we are defaulting to v4 as primary
func GetPodIPs(execer utilexec.Interface, nsenterPath, netnsPath, interfaceName string) ([]net.IP, error) {
	ip, err := getOnePodIP(execer, nsenterPath, netnsPath, interfaceName, "-4")
	if err != nil {
		// Fall back to IPv6 address if no IPv4 address is present
		ip, err = getOnePodIP(execer, nsenterPath, netnsPath, interfaceName, "-6")
	}
	if err != nil {
		return nil, err
	}

	return []net.IP{ip}, nil
}
