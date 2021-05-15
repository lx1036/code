package cni

import (
	"strings"
	"sync"

	kubeletconfig "k8s-lx1036/k8s/kubelet/pkg/apis/config"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/network"

	"github.com/containernetworking/cni/libcni"
	utilexec "k8s.io/utils/exec"
)

type cniNetworkPlugin struct {
	network.NoopNetworkPlugin

	loNetwork *cniNetwork

	sync.RWMutex
	defaultNetwork *cniNetwork

	host        network.Host
	execer      utilexec.Interface
	nsenterPath string
	confDir     string
	binDirs     []string
	cacheDir    string
	podCidr     string
}

func (plugin *cniNetworkPlugin) Init(host interface{}, hairpinMode kubeletconfig.HairpinMode, nonMasqueradeCIDR string, mtu int) error {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) Event(name string, details map[string]interface{}) {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) Name() string {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) Capabilities() interface{} {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) SetUpPod(namespace string, name string, podSandboxID interface{}, annotations, options map[string]string) error {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) TearDownPod(namespace string, name string, podSandboxID interface{}) error {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) GetPodNetworkStatus(namespace string, name string, podSandboxID interface{}) (*interface{}, error) {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) Status() error {
	panic("implement me")
}

type cniNetwork struct {
	name          string
	NetworkConfig *libcni.NetworkConfigList
	CNIConfig     libcni.CNI
	Capabilities  []string
}

// SplitDirs : split dirs by ","
func SplitDirs(dirs string) []string {
	// Use comma rather than colon to work better with Windows too
	return strings.Split(dirs, ",")
}

// ProbeNetworkPlugins : get the network plugin based on cni conf file and bin file
func ProbeNetworkPlugins(confDir, cacheDir string, binDirs []string) []network.NetworkPlugin {
	old := binDirs
	binDirs = make([]string, 0, len(binDirs))
	for _, dir := range old {
		if dir != "" {
			binDirs = append(binDirs, dir)
		}
	}

	plugin := &cniNetworkPlugin{
		defaultNetwork: nil,
		loNetwork:      getLoNetwork(binDirs),
		execer:         utilexec.New(),
		confDir:        confDir,
		binDirs:        binDirs,
		cacheDir:       cacheDir,
	}

	// sync NetworkConfig in best effort during probing.
	plugin.syncNetworkConfig()

	return []network.NetworkPlugin{plugin}
}
