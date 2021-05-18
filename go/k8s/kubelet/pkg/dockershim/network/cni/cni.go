package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	kubeletconfig "k8s-lx1036/k8s/kubelet/pkg/apis/config"
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/network"

	"github.com/containernetworking/cni/libcni"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/util/bandwidth"
	utilslice "k8s.io/kubernetes/pkg/util/slice"
	utilexec "k8s.io/utils/exec"
)

const (
	// CNIPluginName is the name of CNI plugin
	CNIPluginName = "cni"

	// defaultSyncConfigPeriod is the default period to sync CNI config
	defaultSyncConfigPeriod = time.Second * 5

	// supported capabilities
	// https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md
	portMappingsCapability = "portMappings"
	// { "subnet": "10.1.2.0/24", "rangeStart": "10.1.2.3", "rangeEnd": 10.1.2.99", "gateway": "10.1.2.254" }
	ipRangesCapability  = "ipRanges"
	bandwidthCapability = "bandwidth"
	dnsCapability       = "dns"
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

/*
"portMappings": [
  {"hostPort": 8080, "containerPort": 80, "protocol": "tcp"}
]
*/
// cniPortMapping maps to the standard CNI portmapping Capability
// see: https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md
// https://www.cni.dev/docs/conventions/
type cniPortMapping struct {
	HostPort      int32  `json:"hostPort"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"hostIP"`
}

// cniBandwidthEntry maps to the standard CNI bandwidth Capability
// https://github.com/containernetworking/plugins/blob/master/plugins/meta/bandwidth/README.md
// https://www.cni.dev/docs/conventions/
type cniBandwidthEntry struct {
	// IngressRate is the bandwidth rate in bits per second for traffic through container. 0 for no limit. If IngressRate is set, IngressBurst must also be set
	IngressRate int `json:"ingressRate,omitempty"`
	// IngressBurst is the bandwidth burst in bits for traffic through container. 0 for no limit. If IngressBurst is set, IngressRate must also be set
	// NOTE: it's not used for now and defaults to 0. If IngressRate is set IngressBurst will be math.MaxInt32 ~ 2Gbit
	IngressBurst int `json:"ingressBurst,omitempty"`
	// EgressRate is the bandwidth is the bandwidth rate in bits per second for traffic through container. 0 for no limit. If EgressRate is set, EgressBurst must also be set
	EgressRate int `json:"egressRate,omitempty"`
	// EgressBurst is the bandwidth burst in bits for traffic through container. 0 for no limit. If EgressBurst is set, EgressRate must also be set
	// NOTE: it's not used for now and defaults to 0. If EgressRate is set EgressBurst will be math.MaxInt32 ~ 2Gbit
	EgressBurst int `json:"egressBurst,omitempty"`
}

// `nsenter` bin path
func (plugin *cniNetworkPlugin) platformInit() error {
	var err error
	plugin.nsenterPath, err = plugin.execer.LookPath("nsenter")
	if err != nil {
		return err
	}
	return nil
}

func (plugin *cniNetworkPlugin) Init(host network.Host, hairpinMode kubeletconfig.HairpinMode, nonMasqueradeCIDR string, mtu int) error {
	err := plugin.platformInit()
	if err != nil {
		return err
	}

	// INFO: 这一步很重要，为了获取 net namespace，其实就是 "/proc/${pid}/ns/net"
	plugin.host = host

	plugin.syncNetworkConfig()

	// INFO: 这里定期更新 plugin.defaultNetwork
	// start a goroutine to sync network config from confDir periodically to detect network config updates in every 5 seconds
	go wait.Forever(plugin.syncNetworkConfig, defaultSyncConfigPeriod)

	return nil
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

func (plugin *cniNetworkPlugin) checkInitialized() error {
	if plugin.getDefaultNetwork() == nil {
		return fmt.Errorf("cni config uninitialized")
	}

	if utilslice.ContainsString(plugin.getDefaultNetwork().Capabilities, ipRangesCapability, nil) && plugin.podCidr == "" {
		return fmt.Errorf("cni config needs ipRanges but no PodCIDR set")
	}

	return nil
}

func (plugin *cniNetworkPlugin) SetUpPod(namespace string, name string, podSandboxID kubecontainer.ContainerID, annotations, options map[string]string) error {
	if err := plugin.checkInitialized(); err != nil {
		return err
	}

	// /proc/${pid}/ns/net
	netnsPath, err := plugin.host.GetNetNS(podSandboxID.ID)
	if err != nil {
		return fmt.Errorf("CNI failed to retrieve network namespace path: %v", err)
	}

	cniTimeoutCtx, cancelFunc := context.WithTimeout(context.Background(), network.CNITimeoutSec*time.Second)
	defer cancelFunc()

	// Windows doesn't have loNetwork. It comes only with Linux
	if plugin.loNetwork != nil {
		if _, err = plugin.addToNetwork(cniTimeoutCtx, plugin.loNetwork, name, namespace, podSandboxID, netnsPath, annotations, options); err != nil {
			return err
		}
	}

	_, err = plugin.addToNetwork(cniTimeoutCtx, plugin.getDefaultNetwork(), name, namespace, podSandboxID, netnsPath, annotations, options)
	return err
}

// @see https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md
/*
[
	[
		{ "subnet": "10.1.2.0/24", "rangeStart": "10.1.2.3", "rangeEnd": 10.1.2.99", "gateway": "10.1.2.254" }
	]
]
*/
// cniIPRange maps to the standard CNI ip range Capability
type cniIPRange struct {
	Subnet string `json:"subnet"`
}

func (plugin *cniNetworkPlugin) buildCNIRuntimeConf(podName string, podNs string, podSandboxID kubecontainer.ContainerID, podNetnsPath string, annotations, options map[string]string) (*libcni.RuntimeConf, error) {
	rt := &libcni.RuntimeConf{
		ContainerID: podSandboxID.ID,
		NetNS:       podNetnsPath,
		IfName:      network.DefaultInterfaceName,
		CacheDir:    plugin.cacheDir,
		Args: [][2]string{
			{"IgnoreUnknown", "1"},
			{"K8S_POD_NAMESPACE", podNs},
			{"K8S_POD_NAME", podName},
			{"K8S_POD_INFRA_CONTAINER_ID", podSandboxID.ID},
		},
	}

	// INFO: port mappings
	// port mappings are a cni capability-based args, rather than parameters
	// to a specific plugin
	portMappings, err := plugin.host.GetPodPortMappings(podSandboxID.ID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve port mappings: %v", err)
	}
	portMappingsParam := make([]cniPortMapping, 0, len(portMappings))
	for _, p := range portMappings {
		if p.HostPort <= 0 {
			continue
		}
		portMappingsParam = append(portMappingsParam, cniPortMapping{
			HostPort:      p.HostPort,
			ContainerPort: p.ContainerPort,
			Protocol:      strings.ToLower(string(p.Protocol)),
			HostIP:        p.HostIP,
		})
	}
	rt.CapabilityArgs = map[string]interface{}{
		portMappingsCapability: portMappingsParam,
	}

	// INFO: bandwidth
	// { "ingressRate": 2048, "ingressBurst": 1600, "egressRate": 4096, "egressBurst": 1600 }
	ingress, egress, err := bandwidth.ExtractPodBandwidthResources(annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod bandwidth from annotations: %v", err)
	}
	if ingress != nil || egress != nil {
		bandwidthParam := cniBandwidthEntry{}
		if ingress != nil {
			// https://www.cni.dev/docs/conventions/
			// Rates are in bits per second, burst values are in bits.
			bandwidthParam.IngressRate = int(ingress.Value())
			// Limit IngressBurst to math.MaxInt32, in practice limiting to 2Gbit is the equivalent of setting no limit
			bandwidthParam.IngressBurst = math.MaxInt32
		}
		if egress != nil {
			bandwidthParam.EgressRate = int(egress.Value())
			// Limit EgressBurst to math.MaxInt32, in practice limiting to 2Gbit is the equivalent of setting no limit
			bandwidthParam.EgressBurst = math.MaxInt32
		}
		rt.CapabilityArgs[bandwidthCapability] = bandwidthParam
	}

	// INFO: ipRanges
	// Set the PodCIDR
	rt.CapabilityArgs[ipRangesCapability] = [][]cniIPRange{{{Subnet: plugin.podCidr}}}

	// INFO: dns
	// Set dns capability args.
	if dnsOptions, ok := options["dns"]; ok {
		dnsConfig := runtimeapi.DNSConfig{}
		err := json.Unmarshal([]byte(dnsOptions), &dnsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal dns config %q: %v", dnsOptions, err)
		}
		if dnsParam := buildDNSCapabilities(&dnsConfig); dnsParam != nil {
			rt.CapabilityArgs[dnsCapability] = *dnsParam
		}
	}

	return rt, nil
}

// cniDNSConfig maps to the windows CNI dns Capability.
// see: https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md
// https://www.cni.dev/docs/conventions/
// Note that dns capability is only used for Windows containers.
type cniDNSConfig struct {
	// List of DNS servers of the cluster.
	Servers []string `json:"servers,omitempty"`
	// List of DNS search domains of the cluster.
	Searches []string `json:"searches,omitempty"`
	// List of DNS options.
	Options []string `json:"options,omitempty"`
}

// buildDNSCapabilities builds cniDNSConfig from runtimeapi.DNSConfig.
func buildDNSCapabilities(dnsConfig *runtimeapi.DNSConfig) *cniDNSConfig {
	return nil
}

func podDesc(namespace, name string, id kubecontainer.ContainerID) string {
	return fmt.Sprintf("%s_%s/%s", namespace, name, id.ID)
}

func (plugin *cniNetworkPlugin) addToNetwork(ctx context.Context, network *cniNetwork, podName string, podNamespace string,
	podSandboxID kubecontainer.ContainerID, podNetnsPath string, annotations, options map[string]string) (cnitypes.Result, error) {
	rt, err := plugin.buildCNIRuntimeConf(podName, podNamespace, podSandboxID, podNetnsPath, annotations, options)
	if err != nil {
		klog.Errorf("Error adding network when building cni runtime conf: %v", err)
		return nil, err
	}

	pdesc := podDesc(podNamespace, podName, podSandboxID)
	netConf, cniNet := network.NetworkConfig, network.CNIConfig
	klog.V(4).Infof("Adding %s to network %s/%s netns %q", pdesc, netConf.Plugins[0].Network.Type, netConf.Name, podNetnsPath)
	res, err := cniNet.AddNetworkList(ctx, netConf, rt)
	if err != nil {
		klog.Errorf("Error adding %s to network %s/%s: %v", pdesc, netConf.Plugins[0].Network.Type, netConf.Name, err)
		return nil, err
	}
	klog.V(4).Infof("Added %s to network %s: %v", pdesc, netConf.Name, res)

	return res, nil
}

func (plugin *cniNetworkPlugin) TearDownPod(namespace string, name string, podSandboxID kubecontainer.ContainerID) error {
	if err := plugin.checkInitialized(); err != nil {
		return err
	}

	// Lack of namespace should not be fatal on teardown
	netnsPath, err := plugin.host.GetNetNS(podSandboxID.ID)
	if err != nil {
		klog.Warningf("CNI failed to retrieve network namespace path: %v", err)
	}

	cniTimeoutCtx, cancelFunc := context.WithTimeout(context.Background(), network.CNITimeoutSec*time.Second)
	defer cancelFunc()
	// Windows doesn't have loNetwork. It comes only with Linux
	if plugin.loNetwork != nil {
		// Loopback network deletion failure should not be fatal on teardown
		if err := plugin.deleteFromNetwork(cniTimeoutCtx, plugin.loNetwork, name, namespace, podSandboxID, netnsPath, nil); err != nil {
			klog.Warningf("CNI failed to delete loopback network: %v", err)
		}
	}

	return plugin.deleteFromNetwork(cniTimeoutCtx, plugin.getDefaultNetwork(), name, namespace, podSandboxID, netnsPath, nil)
}

func (plugin *cniNetworkPlugin) deleteFromNetwork(ctx context.Context, network *cniNetwork, podName string,
	podNamespace string, podSandboxID kubecontainer.ContainerID, podNetnsPath string, annotations map[string]string) error {
	rt, err := plugin.buildCNIRuntimeConf(podName, podNamespace, podSandboxID, podNetnsPath, annotations, nil)
	if err != nil {
		klog.Errorf("Error deleting network when building cni runtime conf: %v", err)
		return err
	}

	pdesc := podDesc(podNamespace, podName, podSandboxID)
	netConf, cniNet := network.NetworkConfig, network.CNIConfig
	klog.V(4).Infof("Deleting %s from network %s/%s netns %q", pdesc, netConf.Plugins[0].Network.Type, netConf.Name, podNetnsPath)
	err = cniNet.DelNetworkList(ctx, netConf, rt)
	// The pod may not get deleted successfully at the first time.
	// Ignore "no such file or directory" error in case the network has already been deleted in previous attempts.
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		klog.Errorf("Error deleting %s from network %s/%s: %v", pdesc, netConf.Plugins[0].Network.Type, netConf.Name, err)
		return err
	}

	klog.V(4).Infof("Deleted %s from network %s/%s", pdesc, netConf.Plugins[0].Network.Type, netConf.Name)

	return nil
}

// INFO: 实际上就是运行 `nsenter --target=${pid} --net -F -- ip -o -4 addr show dev eth0 scope global` 命令获得 pod ip
func (plugin *cniNetworkPlugin) GetPodNetworkStatus(namespace string, name string, podSandboxID kubecontainer.ContainerID) (*network.PodNetworkStatus, error) {
	// INFO: 获取当前 sandbox container net namespace path netnsPath="/proc/${pid}/ns/net"
	netnsPath, err := plugin.host.GetNetNS(podSandboxID.ID)
	if err != nil {
		return nil, fmt.Errorf("CNI failed to retrieve network namespace path: %v", err)
	}
	if netnsPath == "" {
		return nil, fmt.Errorf("cannot find the network namespace, skipping pod network status for container %q", podSandboxID)
	}

	// `nsenter --net=/proc/${pid}/ns/net -F -- ip -o -4 addr show dev eth0 scope global`
	ips, err := network.GetPodIPs(plugin.execer, plugin.nsenterPath, netnsPath, network.DefaultInterfaceName)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("cannot find pod IPs in the network namespace, skipping pod network status for container %q", podSandboxID)
	}

	return &network.PodNetworkStatus{
		IP:  ips[0],
		IPs: ips,
	}, nil
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

func (plugin *cniNetworkPlugin) getDefaultNetwork() *cniNetwork {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.defaultNetwork
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
