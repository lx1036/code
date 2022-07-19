package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func hasKey(m map[string]interface{}, k string) bool {
	_, ok := m[k]
	return ok
}

func isString(i interface{}) bool {
	_, ok := i.(string)
	return ok
}

func saveScratchNetConf(containerID, dataDir string, netconf []byte) error {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dataDir, containerID)
	return ioutil.WriteFile(path, netconf, 0600)
}

/* /run/flannel/subnet.env file that looks like this
FLANNEL_NETWORK=10.1.0.0/16
FLANNEL_SUBNET=10.1.17.1/24
FLANNEL_MTU=1472
FLANNEL_IPMASQ=true
*/
type subnetEnv struct {
	network *net.IPNet
	subnet  *net.IPNet
	ip6Nw   *net.IPNet
	ip6Sn   *net.IPNet
	mtu     *uint
	ipmasq  *bool
}

func (se *subnetEnv) missing() string {
	m := []string{}

	if se.network == nil && se.ip6Nw == nil {
		m = append(m, []string{"FLANNEL_NETWORK", "FLANNEL_IPV6_NETWORK"}...)
	}
	if se.subnet == nil && se.ip6Sn == nil {
		m = append(m, []string{"FLANNEL_SUBNET", "FLANNEL_IPV6_SUBNET"}...)
	}
	if se.mtu == nil {
		m = append(m, "FLANNEL_MTU")
	}
	if se.ipmasq == nil {
		m = append(m, "FLANNEL_IPMASQ")
	}
	return strings.Join(m, ", ")
}

func loadFlannelSubnetEnv(fn string) (*subnetEnv, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	se := &subnetEnv{}

	s := bufio.NewScanner(f)
	for s.Scan() {
		parts := strings.SplitN(s.Text(), "=", 2)
		switch parts[0] {
		case "FLANNEL_NETWORK":
			_, se.network, err = net.ParseCIDR(parts[1])
			if err != nil {
				return nil, err
			}

		case "FLANNEL_SUBNET":
			_, se.subnet, err = net.ParseCIDR(parts[1])
			if err != nil {
				return nil, err
			}

		case "FLANNEL_IPV6_NETWORK":
			_, se.ip6Nw, err = net.ParseCIDR(parts[1])
			if err != nil {
				return nil, err
			}

		case "FLANNEL_IPV6_SUBNET":
			_, se.ip6Sn, err = net.ParseCIDR(parts[1])
			if err != nil {
				return nil, err
			}

		case "FLANNEL_MTU":
			mtu, err := strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return nil, err
			}
			se.mtu = new(uint)
			*se.mtu = uint(mtu)

		case "FLANNEL_IPMASQ":
			ipmasq := parts[1] == "true"
			se.ipmasq = &ipmasq
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	if m := se.missing(); m != "" {
		return nil, fmt.Errorf("%v is missing %v", fn, m)
	}

	return se, nil
}
