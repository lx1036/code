package types

import (
    "encoding/json"
    "fmt"
    "os"

    ipamTypes "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/ipam/types"

    cniTypes "github.com/containernetworking/cni/pkg/types"
    current "github.com/containernetworking/cni/pkg/types/100"
    "github.com/containernetworking/cni/pkg/version"
)

// IPAM is the Cilium specific CNI IPAM configuration
type IPAM struct {
    cniTypes.IPAM
    ipamTypes.IPAMSpec
}

// Args contains arbitrary information a scheduler
// can pass to the cni plugin
type Args struct{}

// NetConf is the Cilium specific CNI network configuration
type NetConf struct {
    cniTypes.NetConf
    MTU         int    `json:"mtu"`
    Args        Args   `json:"args"`
    IPAM        IPAM   `json:"ipam,omitempty"` // Shadows the JSON field "ipam" in cniTypes.NetConf.
    EnableDebug bool   `json:"enable-debug"`
    LogFormat   string `json:"log-format"`
    LogFile     string `json:"log-file"`
}

// NetConfList is a CNI chaining configuration
type NetConfList struct {
    Plugins []*NetConf `json:"plugins,omitempty"`
}

// ReadNetConf reads a CNI configuration file and returns the corresponding
// NetConf structure
func ReadNetConf(path string) (*NetConf, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("unable to read CNI configuration '%s': %s", path, err)
    }

    netConfList := &NetConfList{}
    if err := json.Unmarshal(b, netConfList); err == nil {
        for _, plugin := range netConfList.Plugins {
            if plugin.Type == "cilium-cni" {
                return parsePrevResult(plugin)
            }
        }
    }

    return LoadNetConf(b)
}

func parsePrevResult(n *NetConf) (*NetConf, error) {
    if n.RawPrevResult != nil {
        resultBytes, err := json.Marshal(n.RawPrevResult)
        if err != nil {
            return nil, fmt.Errorf("could not serialize prevResult: %v", err)
        }
        res, err := version.NewResult(n.CNIVersion, resultBytes)
        if err != nil {
            return nil, fmt.Errorf("could not parse prevResult: %v", err)
        }
        n.PrevResult, err = current.NewResultFromResult(res)
        if err != nil {
            return nil, fmt.Errorf("could not convert result to current version: %v", err)
        }
    }

    return n, nil
}

// LoadNetConf unmarshals a Cilium network configuration from JSON and returns
// a NetConf together with the CNI version
func LoadNetConf(bytes []byte) (*NetConf, error) {
    n := &NetConf{}
    if err := json.Unmarshal(bytes, n); err != nil {
        return nil, fmt.Errorf("failed to load netconf: %s", err)
    }

    return parsePrevResult(n)
}
