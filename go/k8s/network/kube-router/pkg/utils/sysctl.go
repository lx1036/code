package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	// Network Routes Configuration Paths
	BridgeNFCallIPTables  = "net/bridge/bridge-nf-call-iptables"
	BridgeNFCallIP6Tables = "net/bridge/bridge-nf-call-ip6tables"
)

func SetSysctl(path string, value int) error {
	sysctlPath := fmt.Sprintf("/proc/sys/%s", path)
	if _, err := os.Stat(sysctlPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("option not found, Does your kernel version support this feature? err: %v", err)
		}
		return fmt.Errorf("path existed, but could not be stat'd err:%v", err)
	}
	err := ioutil.WriteFile(sysctlPath, []byte(strconv.Itoa(value)), 0640)
	if err != nil {
		return fmt.Errorf("path could not be set %v", err)
	}

	return nil
}
