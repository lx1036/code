package allocator

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	cnitypes "github.com/containernetworking/cni/pkg/types"
)

const (
	IPAMType = "host-local-cluster-wide"

	AddTimeLimit = 2 * time.Minute

	// Allocate operation identifier
	Allocate = 0
	// Deallocate operation identifier
	Deallocate = 1
)

type Net struct {
	Name       string      `json:"name"`
	CNIVersion string      `json:"cniVersion"`
	IPAM       *IPAMConfig `json:"ipam"`
}

type IPAMConfig struct {
	Name       string
	Type       string            `json:"type"`
	Routes     []*cnitypes.Route `json:"routes"`
	ResolvConf string            `json:"resolvConf"` // /etc/resolv.conf

	Range      string `json:"range"`
	RangeStart net.IP `json:"range_start,omitempty"`
	RangeEnd   net.IP `json:"range_end,omitempty"`
	Gateway    string `json:"gateway"`
}

func LoadIPAMConfig(bytes []byte, envArgs string) (*IPAMConfig, string, error) {
	n := Net{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return nil, "", err
	}

	if n.IPAM == nil {
		return nil, "", fmt.Errorf("IPAM config missing 'ipam' key")
	} else if n.IPAM.Type != IPAMType {
		return nil, "", fmt.Errorf(fmt.Sprintf("ipam type %s is not valid", n.IPAM.Type))
	}

	// Copy net name into IPAM so not to drag Net struct around
	n.IPAM.Name = n.Name

	return n.IPAM, n.CNIVersion, nil
}
