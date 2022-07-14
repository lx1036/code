package subnet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type Config struct {
	EnableIPv4    bool
	EnableIPv6    bool
	Network       ip.IP4Net
	SubnetMin     ip.IP4
	SubnetMax     ip.IP4
	SubnetLen     uint
	IPv6SubnetLen uint
	BackendType   string          `json:"-"`
	Backend       json.RawMessage `json:",omitempty"`
}

/*
 net-conf.json: |
    {
      "Network": "10.244.0.0/16",
      "Backend": {
        "Type": "vxlan"
      }
    }
*/
func getSubnetConfig(netConfPath string) (*Config, error) {
	netConf, err := ioutil.ReadFile(netConfPath) // /etc/kube-flannel/net-conf.json
	if err != nil {
		return nil, fmt.Errorf("failed to read net conf: %v", err)
	}

	config := Config{
		EnableIPv4: true, // Enable ipv4 by default
	}
	if err = json.Unmarshal([]byte(netConf), &config); err != nil {
		return nil, err
	}

	if config.EnableIPv4 { // check subnet config
		if config.SubnetLen > 0 {
			// SubnetLen needs to allow for a tunnel and bridge device on each host.
			if config.SubnetLen > 30 {
				return nil, errors.New("SubnetLen must be less than /31")
			}

			// SubnetLen needs to fit _more_ than twice into the Network.
			// the first subnet isn't used, so splitting into two one only provide one usable host.
			if config.SubnetLen < config.Network.PrefixLen+2 {
				return nil, errors.New("Network must be able to accommodate at least four subnets")
			}
		} else {
			// If the network is smaller than a /28 then the network isn't big enough for flannel so return an error.
			// Default to giving each host at least a /24 (as long as the network is big enough to support at least four hosts)
			// Otherwise, if the network is too small to give each host a /24 just split the network into four.
			if config.Network.PrefixLen > 28 {
				// Each subnet needs at least four addresses (/30) and the network needs to accommodate at least four
				// since the first subnet isn't used, so splitting into two would only provide one usable host.
				// So the min useful PrefixLen is /28
				return nil, errors.New("Network is too small. Minimum useful network prefix is /28")
			} else if config.Network.PrefixLen <= 22 {
				// Network is big enough to give each host a /24
				config.SubnetLen = 24
			} else {
				// Use +2 to provide four hosts per subnet.
				config.SubnetLen = config.Network.PrefixLen + 2
			}
		}

		subnetSize := ip.IP4(1 << (32 - config.SubnetLen))

		if config.SubnetMin == ip.IP4(0) {
			// skip over the first subnet otherwise it causes problems. e.g.
			// if Network is 10.100.0.0/16, having an interface with 10.100.0.0
			// conflicts with the network address.
			config.SubnetMin = config.Network.IP + subnetSize
		} else if !config.Network.Contains(config.SubnetMin) {
			return nil, errors.New("SubnetMin is not in the range of the Network")
		}

		if config.SubnetMax == ip.IP4(0) {
			config.SubnetMax = config.Network.Next().IP - subnetSize
		} else if !config.Network.Contains(config.SubnetMax) {
			return nil, errors.New("SubnetMax is not in the range of the Network")
		}

		// The SubnetMin and SubnetMax need to be aligned to a SubnetLen boundary
		mask := ip.IP4(0xFFFFFFFF << (32 - config.SubnetLen))
		if config.SubnetMin != config.SubnetMin&mask {
			return nil, fmt.Errorf("SubnetMin is not on a SubnetLen boundary: %v", config.SubnetMin)
		}

		if config.SubnetMax != config.SubnetMax&mask {
			return nil, fmt.Errorf("SubnetMax is not on a SubnetLen boundary: %v", config.SubnetMax)
		}
	}

	// json decode Backend Type
	type BackendType struct {
		Type string
	}
	var backend BackendType
	if err := json.Unmarshal([]byte(config.Backend), &backend); err != nil {
		return nil, fmt.Errorf("error decoding Backend property of config: %v", err)
	}

	config.BackendType = backend.Type // 默认会是 vxlan

	return config, nil
}
