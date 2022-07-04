package types

import (
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	"os"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
)

var (
	nodeName = "localhost"
)

func SetName(name string) {
	nodeName = name
}

// GetName returns the name of the local node. The value returned was either
// previously set with SetName(), retrieved via `os.Hostname()`, or as a last
// resort is hardcoded to "localhost".
func GetName() string {
	return nodeName
}

func init() {
	// Give priority to the environment variable available in the Cilium agent
	if name := os.Getenv(defaults.EnvNodeNameSpec); name != "" {
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
