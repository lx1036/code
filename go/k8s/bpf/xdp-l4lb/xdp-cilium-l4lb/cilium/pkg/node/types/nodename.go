package types

import (
    log "github.com/sirupsen/logrus"
    "os"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/defaults"
    k8sConsts "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/k8s/constants"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

var (
    nodeName = "localhost"
)

func init() {
    // Give priority to the environment variable available in the Cilium agent
    if name := os.Getenv(k8sConsts.EnvNodeNameSpec); name != "" {
        nodeName = name
        return
    }
    if h, err := os.Hostname(); err != nil {
        log.WithError(err).Warn("Unable to retrieve local hostname")
    } else {
        log.WithField(logfields.NodeName, h).Debug("os.Hostname() returned")
        nodeName = h
    }
}

// SetName sets the name of the local node. This will overwrite the value that
// is automatically retrieved with `os.Hostname()`.
//
// Note: This function is currently designed to only be called during the
// bootstrapping procedure of the agent where no parallelism exists. If you
// want to use this function in later stages, a mutex must be added first.
func SetName(name string) {
    nodeName = name
}

// GetName returns the name of the local node. The value returned was either
// previously set with SetName(), retrieved via `os.Hostname()`, or as a last
// resort is hardcoded to "localhost".
func GetName() string {
    return nodeName
}

// GetAbsoluteNodeName returns the absolute node name combined of both
// (prefixed)cluster name and the local node name in case of
// clustered environments otherwise returns the name of the local node.
func GetAbsoluteNodeName() string {
    if option.Config.ClusterName != "" &&
        option.Config.ClusterName != defaults.ClusterName {
        return option.Config.ClusterName + "/" + nodeName
    } else {
        return nodeName
    }
}
