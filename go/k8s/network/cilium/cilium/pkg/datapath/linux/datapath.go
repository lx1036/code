package linux

import (
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath/linux/config"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath/loader"
)

// DatapathConfiguration is the static configuration of the datapath. The
// configuration cannot change throughout the lifetime of a datapath object.
type DatapathConfiguration struct {
	// HostDevice is the name of the device to be used to access the host.
	HostDevice string
	// EncryptInterface is the name of the device to be used for direct ruoting encryption
	EncryptInterface string
}

type linuxDatapath struct {
	datapath.ConfigWriter

	loader *loader.Loader
}

// NewDatapath creates a new Linux datapath
func NewDatapath(cfg DatapathConfiguration, ruleManager datapath.IptablesManager) datapath.Datapath {

	dp := &linuxDatapath{
		ConfigWriter: &config.HeaderfileWriter{},
		loader:       loader.NewLoader(),
	}

}

func (l *linuxDatapath) Loader() datapath.Loader {
	return l.loader
}
