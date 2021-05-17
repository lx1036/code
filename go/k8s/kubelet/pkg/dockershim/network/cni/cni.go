package cni

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	kubeletconfig "k8s-lx1036/k8s/kubelet/pkg/apis/config"
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/network"

	"github.com/containernetworking/cni/libcni"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	utilexec "k8s.io/utils/exec"
)

const (
	// CNIPluginName is the name of CNI plugin
	CNIPluginName = "cni"
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

func (plugin *cniNetworkPlugin) Init(host network.Host, hairpinMode kubeletconfig.HairpinMode, nonMasqueradeCIDR string, mtu int) error {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) Event(name string, details map[string]interface{}) {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) Name() string {
	return CNIPluginName
}

func (plugin *cniNetworkPlugin) Capabilities() sets.Int {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) SetUpPod(namespace string, name string, podSandboxID kubecontainer.ContainerID, annotations, options map[string]string) error {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) TearDownPod(namespace string, name string, podSandboxID kubecontainer.ContainerID) error {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) GetPodNetworkStatus(namespace string, name string, podSandboxID kubecontainer.ContainerID) (*network.PodNetworkStatus, error) {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) Status() error {
	panic("implement me")
}

func (plugin *cniNetworkPlugin) syncNetworkConfig() {
	defaultNetwork, err := getDefaultCNINetwork(plugin.confDir, plugin.binDirs)
	if err != nil {
		klog.Warningf("Unable to update cni config: %s", err)
		return
	}

	plugin.setDefaultNetwork(defaultNetwork)
}

func (plugin *cniNetworkPlugin) setDefaultNetwork(n *cniNetwork) {
	plugin.Lock()
	defer plugin.Unlock()
	plugin.defaultNetwork = n
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

func getLoNetwork(binDirs []string) *cniNetwork {
	loConfig, err := libcni.ConfListFromBytes([]byte(`{
  "cniVersion": "0.2.0",
  "name": "cni-loopback",
  "plugins":[{
    "type": "loopback"
  }]
}`))
	if err != nil {
		// The hardcoded config above should always be valid and unit tests will
		// catch this
		panic(err)
	}
	loNetwork := &cniNetwork{
		name:          "lo",
		NetworkConfig: loConfig,
		CNIConfig:     &libcni.CNIConfig{Path: binDirs},
	}

	return loNetwork
}

func getDefaultCNINetwork(confDir string, binDirs []string) (*cniNetwork, error) {
	files, err := libcni.ConfFiles(confDir, []string{".conf", ".conflist", ".json"})
	switch {
	case err != nil:
		return nil, err
	case len(files) == 0:
		return nil, fmt.Errorf("no networks found in %s", confDir)
	}

	cniConfig := &libcni.CNIConfig{Path: binDirs}

	sort.Strings(files)
	for _, confFile := range files {
		var confList *libcni.NetworkConfigList
		if strings.HasSuffix(confFile, ".conflist") {
			confList, err = libcni.ConfListFromFile(confFile)
			if err != nil {
				klog.Warningf("Error loading CNI config list file %s: %v", confFile, err)
				continue
			}
		} else {
			conf, err := libcni.ConfFromFile(confFile)
			if err != nil {
				klog.Warningf("Error loading CNI config file %s: %v", confFile, err)
				continue
			}
			// Ensure the config has a "type" so we know what plugin to run.
			// Also catches the case where somebody put a conflist into a conf file.
			if conf.Network.Type == "" {
				klog.Warningf("Error loading CNI config file %s: no 'type'; perhaps this is a .conflist?", confFile)
				continue
			}

			confList, err = libcni.ConfListFromConf(conf)
			if err != nil {
				klog.Warningf("Error converting CNI config file %s to list: %v", confFile, err)
				continue
			}
		}
		if len(confList.Plugins) == 0 {
			klog.Warningf("CNI config list %s has no networks, skipping", string(confList.Bytes[:maxStringLengthInLog(len(confList.Bytes))]))
			continue
		}

		// Before using this CNI config, we have to validate it to make sure that
		// all plugins of this config exist on disk
		caps, err := cniConfig.ValidateNetworkList(context.TODO(), confList)
		if err != nil {
			klog.Warningf("Error validating CNI config list %s: %v", string(confList.Bytes[:maxStringLengthInLog(len(confList.Bytes))]), err)
			continue
		}

		klog.V(4).Infof("Using CNI configuration file %s", confFile)

		return &cniNetwork{
			name:          confList.Name,
			NetworkConfig: confList,
			CNIConfig:     cniConfig,
			Capabilities:  caps,
		}, nil
	}
	return nil, fmt.Errorf("no valid networks found in %s", confDir)
}

func maxStringLengthInLog(length int) int {
	// we allow no more than 4096-length strings to be logged
	const maxStringLength = 4096

	if length < maxStringLength {
		return length
	}
	return maxStringLength
}
