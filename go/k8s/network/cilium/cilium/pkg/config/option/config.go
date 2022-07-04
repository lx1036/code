package option

import (
	"bytes"
	"fmt"
	"github.com/cilium/cilium/pkg/defaults"
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ciliumEnvPrefix is the prefix used for environment variables
	ciliumEnvPrefix = "CILIUM_"

	// ConfigFile is the Configuration file (default "$HOME/ciliumd.yaml")
	ConfigFile = "config"

	// ConfigDir is the directory that contains a file for each option where
	// the filename represents the option name and the content of that file
	// represents the value of that option.
	ConfigDir = "config-dir"

	// InstallIptRules sets whether Cilium should install any iptables in general
	InstallIptRules = "install-iptables-rules"
)

// Available option for DaemonConfig.Tunnel
const (
	// TunnelVXLAN specifies VXLAN encapsulation
	TunnelVXLAN = "vxlan"

	// TunnelGeneve specifies Geneve encapsulation
	TunnelGeneve = "geneve"

	// TunnelDisabled specifies to disable encapsulation
	TunnelDisabled = "disabled"

	////////////////////////////// BPF //////////////////////////

	// SockopsEnableName is the name of the option to enable sockops
	SockopsEnableName = "sockops-enable"
)

// DaemonConfig is the configuration used by Daemon.
type DaemonConfig struct {
	////////////////////////////// Base //////////////////////////
	ConfigFile string
	ConfigDir  string
	// StateDir is the directory where runtime state of endpoints is stored
	StateDir string // /var/run/cilium/state/
	// EnableIPv4 is true when IPv4 is enabled
	EnableIPv4 bool
	// EnableIPv6 is true when IPv6 is enabled
	EnableIPv6 bool

	////////////////////////////// Datapath //////////////////////////
	DatapathMode string // Datapath mode
	Tunnel       string // Tunnel mode

	////////////////////////////// Service //////////////////////////
	// EnableNodePort enables k8s NodePort service implementation in BPF
	EnableNodePort bool
	// EnableHostPort enables k8s Pod's hostPort mapping through BPF
	EnableHostPort bool
	// NodePortMode indicates in which mode NodePort implementation should run
	// ("snat", "dsr" or "hybrid")
	NodePortMode string
	// NodePortAcceleration indicates whether NodePort should be accelerated
	// via XDP ("none", "generic" or "native")
	NodePortAcceleration string
	// NodePortHairpin indicates whether the setup is a one-legged LB
	NodePortHairpin bool
	// NodePortBindProtection rejects bind requests to NodePort service ports
	NodePortBindProtection bool
	// EnableAutoProtectNodePortRange enables appending NodePort range to
	// net.ipv4.ip_local_reserved_ports if it overlaps with ephemeral port
	// range (net.ipv4.ip_local_port_range)
	EnableAutoProtectNodePortRange bool
	// KubeProxyReplacement controls how to enable kube-proxy replacement
	// features in BPF datapath
	KubeProxyReplacement string
	// EnableExternalIPs enables implementation of k8s services with externalIPs in datapath
	EnableExternalIPs bool
	// NodePortMin is the minimum port address for the NodePort range
	NodePortMin int
	// NodePortMax is the maximum port address for the NodePort range
	NodePortMax int
	// EnableSessionAffinity enables a support for service sessionAffinity
	EnableSessionAffinity bool
	// K8sEnableEndpointSlice enables k8s endpoint slice feature that is used
	// in kubernetes.
	K8sEnableK8sEndpointSlice bool

	////////////////////////////// CNI //////////////////////////
	// EnableEndpointRoutes enables use of per endpoint routes
	EnableEndpointRoutes bool
	Devices              []string // bpf_host device

	////////////////////////////// BPF //////////////////////////
	// EnableSockOps specifies whether to enable sockops (socket lookup).
	SockopsEnable bool // socket bpf

	InstallIptRules bool
}

func (c *DaemonConfig) Populate() {

	c.AgentHealthPort = viper.GetInt(AgentHealthPort)
	c.AgentLabels = viper.GetStringSlice(AgentLabels)
	c.AllowICMPFragNeeded = viper.GetBool(AllowICMPFragNeeded)
	c.AllowLocalhost = viper.GetString(AllowLocalhost)
	c.AnnotateK8sNode = viper.GetBool(AnnotateK8sNode)
	c.AutoCreateCiliumNodeResource = viper.GetBool(AutoCreateCiliumNodeResource)
	c.BPFCompilationDebug = viper.GetBool(BPFCompileDebugName)
	c.BPFRoot = viper.GetString(BPFRoot)
	c.CertDirectory = viper.GetString(CertsDirectory)
	c.CGroupRoot = viper.GetString(CGroupRoot)
	c.ClusterID = viper.GetInt(ClusterIDName)
	c.ClusterName = viper.GetString(ClusterName)
	c.ClusterMeshConfig = viper.GetString(ClusterMeshConfigName)
	c.DatapathMode = viper.GetString(DatapathMode)
	c.Debug = viper.GetBool(DebugArg)
	c.DebugVerbose = viper.GetStringSlice(DebugVerbose)
	c.DirectRoutingDevice = viper.GetString(DirectRoutingDevice)
	c.DisableConntrack = viper.GetBool(DisableConntrack)
	c.EnableIPv4 = getIPv4Enabled()
	c.EnableIPv6 = viper.GetBool(EnableIPv6Name)
	c.EnableIPSec = viper.GetBool(EnableIPSecName)
	c.EnableWellKnownIdentities = viper.GetBool(EnableWellKnownIdentities)
	c.EndpointInterfaceNamePrefix = viper.GetString(EndpointInterfaceNamePrefix)
	c.DevicePreFilter = viper.GetString(PrefilterDevice)
	c.DisableCiliumEndpointCRD = viper.GetBool(DisableCiliumEndpointCRDName)
	c.DisableK8sServices = viper.GetBool(DisableK8sServices)
	c.EgressMasqueradeInterfaces = viper.GetString(EgressMasqueradeInterfaces)
	c.EnableHostReachableServices = viper.GetBool(EnableHostReachableServices)
	c.EnableRemoteNodeIdentity = viper.GetBool(EnableRemoteNodeIdentity)
	c.K8sHeartbeatTimeout = viper.GetDuration(K8sHeartbeatTimeout)
	c.EnableXTSocketFallback = viper.GetBool(EnableXTSocketFallbackName)
	c.EnableAutoDirectRouting = viper.GetBool(EnableAutoDirectRoutingName)
	c.EnableEndpointRoutes = viper.GetBool(EnableEndpointRoutes)
	c.EnableHealthChecking = viper.GetBool(EnableHealthChecking)
	c.EnableEndpointHealthChecking = viper.GetBool(EnableEndpointHealthChecking)
	c.EnableLocalNodeRoute = viper.GetBool(EnableLocalNodeRoute)
	c.EnablePolicy = strings.ToLower(viper.GetString(EnablePolicy))
	c.EnableExternalIPs = viper.GetBool(EnableExternalIPs)
	c.EnableL7Proxy = viper.GetBool(EnableL7Proxy)
	c.EnableTracing = viper.GetBool(EnableTracing)
	c.EnableNodePort = viper.GetBool(EnableNodePort)
	c.EnableHostPort = viper.GetBool(EnableHostPort)
	c.NodePortMode = viper.GetString(NodePortMode)
	c.NodePortAcceleration = viper.GetString(NodePortAcceleration)
	c.NodePortBindProtection = viper.GetBool(NodePortBindProtection)
	c.EnableAutoProtectNodePortRange = viper.GetBool(EnableAutoProtectNodePortRange)
	c.KubeProxyReplacement = viper.GetString(KubeProxyReplacement)
	c.EnableSessionAffinity = viper.GetBool(EnableSessionAffinity)
	c.EnableHostFirewall = viper.GetBool(EnableHostFirewall)
	c.EncryptInterface = viper.GetString(EncryptInterface)
	c.EncryptNode = viper.GetBool(EncryptNode)
	c.EnvoyLogPath = viper.GetString(EnvoyLog)
	c.ForceLocalPolicyEvalAtSource = viper.GetBool(ForceLocalPolicyEvalAtSource)
	c.HostDevice = getHostDevice()
	c.HTTPIdleTimeout = viper.GetInt(HTTPIdleTimeout)
	c.HTTPMaxGRPCTimeout = viper.GetInt(HTTPMaxGRPCTimeout)
	c.HTTPRequestTimeout = viper.GetInt(HTTPRequestTimeout)
	c.HTTPRetryCount = viper.GetInt(HTTPRetryCount)
	c.HTTPRetryTimeout = viper.GetInt(HTTPRetryTimeout)
	c.IdentityChangeGracePeriod = viper.GetDuration(IdentityChangeGracePeriod)
	c.IPAM = viper.GetString(IPAM)
	c.IPv4Range = viper.GetString(IPv4Range)
	c.IPv4NodeAddr = viper.GetString(IPv4NodeAddr)
	c.IPv4ServiceRange = viper.GetString(IPv4ServiceRange)
	c.IPv6ClusterAllocCIDR = viper.GetString(IPv6ClusterAllocCIDRName)
	c.IPv6NodeAddr = viper.GetString(IPv6NodeAddr)
	c.IPv6Range = viper.GetString(IPv6Range)
	c.IPv6ServiceRange = viper.GetString(IPv6ServiceRange)
	c.K8sAPIServer = viper.GetString(K8sAPIServer)
	c.K8sClientBurst = viper.GetInt(K8sClientBurst)
	c.K8sClientQPSLimit = viper.GetFloat64(K8sClientQPSLimit)
	c.K8sEnableK8sEndpointSlice = viper.GetBool(K8sEnableEndpointSlice)
	c.k8sEnableAPIDiscovery = viper.GetBool(K8sEnableAPIDiscovery)
	c.K8sKubeConfigPath = viper.GetString(K8sKubeConfigPath)
	c.K8sRequireIPv4PodCIDR = viper.GetBool(K8sRequireIPv4PodCIDRName)
	c.K8sRequireIPv6PodCIDR = viper.GetBool(K8sRequireIPv6PodCIDRName)
	c.K8sServiceCacheSize = uint(viper.GetInt(K8sServiceCacheSize))
	c.K8sForceJSONPatch = viper.GetBool(K8sForceJSONPatch)
	c.K8sEventHandover = viper.GetBool(K8sEventHandover)
	c.K8sWatcherQueueSize = uint(viper.GetInt(K8sWatcherQueueSize))
	c.K8sWatcherEndpointSelector = viper.GetString(K8sWatcherEndpointSelector)
	c.KeepConfig = viper.GetBool(KeepConfig)
	c.KVStore = viper.GetString(KVStore)
	c.KVstoreLeaseTTL = viper.GetDuration(KVstoreLeaseTTL)
	c.KVstoreKeepAliveInterval = c.KVstoreLeaseTTL / defaults.KVstoreKeepAliveIntervalFactor
	c.KVstorePeriodicSync = viper.GetDuration(KVstorePeriodicSync)
	c.KVstoreConnectivityTimeout = viper.GetDuration(KVstoreConnectivityTimeout)
	c.IPAllocationTimeout = viper.GetDuration(IPAllocationTimeout)
	c.LabelPrefixFile = viper.GetString(LabelPrefixFile)
	c.Labels = viper.GetStringSlice(Labels)
	c.LibDir = viper.GetString(LibDir)
	c.LogDriver = viper.GetStringSlice(LogDriver)
	c.LogSystemLoadConfig = viper.GetBool(LogSystemLoadConfigName)
	c.Logstash = viper.GetBool(Logstash)
	c.LoopbackIPv4 = viper.GetString(LoopbackIPv4)
	c.Masquerade = viper.GetBool(Masquerade)
	c.EnableBPFMasquerade = viper.GetBool(EnableBPFMasquerade)
	c.EnableBPFClockProbe = viper.GetBool(EnableBPFClockProbe)
	c.EnableIPMasqAgent = viper.GetBool(EnableIPMasqAgent)
	c.IPMasqAgentConfigPath = viper.GetString(IPMasqAgentConfigPath)
	c.InstallIptRules = viper.GetBool(InstallIptRules)
	c.IPTablesLockTimeout = viper.GetDuration(IPTablesLockTimeout)
	c.IPSecKeyFile = viper.GetString(IPSecKeyFileName)
	c.ModePreFilter = viper.GetString(PrefilterMode)
	c.MonitorAggregation = viper.GetString(MonitorAggregationName)
	c.MonitorAggregationInterval = viper.GetDuration(MonitorAggregationInterval)
	c.MonitorQueueSize = viper.GetInt(MonitorQueueSizeName)
	c.MTU = viper.GetInt(MTUName)
	c.NAT46Range = viper.GetString(NAT46Range)
	c.FlannelMasterDevice = viper.GetString(FlannelMasterDevice)
	c.FlannelUninstallOnExit = viper.GetBool(FlannelUninstallOnExit)
	c.PProf = viper.GetBool(PProf)
	c.PreAllocateMaps = viper.GetBool(PreAllocateMapsName)
	c.PrependIptablesChains = viper.GetBool(PrependIptablesChainsName)
	c.PrometheusServeAddr = getPrometheusServerAddr()
	c.ProxyConnectTimeout = viper.GetInt(ProxyConnectTimeout)
	c.BlacklistConflictingRoutes = viper.GetBool(BlacklistConflictingRoutes)
	c.ReadCNIConfiguration = viper.GetString(ReadCNIConfiguration)
	c.RestoreState = viper.GetBool(Restore)
	c.RunDir = viper.GetString(StateDir)
	c.SidecarIstioProxyImage = viper.GetString(SidecarIstioProxyImage)
	c.UseSingleClusterRoute = viper.GetBool(SingleClusterRouteName)
	c.SocketPath = viper.GetString(SocketPath)
	c.SockopsEnable = viper.GetBool(SockopsEnableName)
	c.TracePayloadlen = viper.GetInt(TracePayloadlen)
	c.Tunnel = viper.GetString(TunnelName)
	c.Version = viper.GetString(Version)
	c.WriteCNIConfigurationWhenReady = viper.GetString(WriteCNIConfigurationWhenReady)
	c.PolicyTriggerInterval = viper.GetDuration(PolicyTriggerInterval)
	c.CTMapEntriesTimeoutTCP = viper.GetDuration(CTMapEntriesTimeoutTCPName)
	c.CTMapEntriesTimeoutAny = viper.GetDuration(CTMapEntriesTimeoutAnyName)
	c.CTMapEntriesTimeoutSVCTCP = viper.GetDuration(CTMapEntriesTimeoutSVCTCPName)
	c.CTMapEntriesTimeoutSVCAny = viper.GetDuration(CTMapEntriesTimeoutSVCAnyName)
	c.CTMapEntriesTimeoutSYN = viper.GetDuration(CTMapEntriesTimeoutSYNName)
	c.CTMapEntriesTimeoutFIN = viper.GetDuration(CTMapEntriesTimeoutFINName)
	c.PolicyAuditMode = viper.GetBool(PolicyAuditModeArg)
	c.EnableIPv4FragmentsTracking = viper.GetBool(EnableIPv4FragmentsTrackingName)
	c.FragmentsMapEntries = viper.GetInt(FragmentsMapEntriesName)

}

var (
	// Config represents the daemon configuration
	Config = &DaemonConfig{}
)

// InitConfig reads in config file and ENV variables if set.
func InitConfig(configName string) func() {
	return func() {
		Config.ConfigFile = viper.GetString(ConfigFile) // enable ability to specify config file via flag
		Config.ConfigDir = viper.GetString(ConfigDir)
		viper.SetEnvPrefix("cilium")

		// INFO: 启动时候用的 --config-dir=/tmp/cilium/config-map, 每一个文件名 filename 是 key，文件内容是 value
		if Config.ConfigDir != "" {
			if _, err := os.Stat(Config.ConfigDir); os.IsNotExist(err) {
				log.Fatalf("Non-existent configuration directory %s", Config.ConfigDir)
			}

			if m, err := ReadDirConfig(Config.ConfigDir); err != nil {
				log.Fatalf("Unable to read configuration directory %s: %s", Config.ConfigDir, err)
			} else {
				err := MergeConfig(m)
				if err != nil {
					log.Fatalf("Unable to merge configuration: %s", err)
				}
			}
		}

		if Config.ConfigFile != "" {
			viper.SetConfigFile(Config.ConfigFile)
		} else {
			viper.SetConfigName(configName) // name of config file (without extension)
			viper.AddConfigPath("$HOME")    // adding home directory as first search path
		}

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			log.WithField(logfields.Path, viper.ConfigFileUsed()).
				Info("Using config from file")
		} else if Config.ConfigFile != "" {
			log.WithField(logfields.Path, Config.ConfigFile).
				Fatal("Error reading config file")
		} else {
			log.WithField(logfields.Reason, err).Info("Skipped reading configuration file")
		}
	}
}

// MergeConfig merges the given configuration map with viper's configuration.
func MergeConfig(m map[string]interface{}) error {
	err := viper.MergeConfigMap(m)
	if err != nil {
		return fmt.Errorf("unable to read merge directory configuration: %s", err)
	}
	return nil
}

// ReadDirConfig reads the given directory and returns a map that maps the
// filename to the contents of that file.
func ReadDirConfig(dirName string) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	fi, err := ioutil.ReadDir(dirName)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to read configuration directory: %s", err)
	}
	for _, f := range fi {
		if f.Mode().IsDir() {
			continue
		}
		fName := filepath.Join(dirName, f.Name())

		// the file can still be a symlink to a directory
		if f.Mode()&os.ModeSymlink == 0 {
			absFileName, err := filepath.EvalSymlinks(fName)
			if err != nil {
				log.Warnf("Unable to read configuration file %q: %s", absFileName, err)
				continue
			}
			fName = absFileName
		}

		f, err = os.Stat(fName)
		if err != nil {
			log.Warnf("Unable to read configuration file %q: %s", fName, err)
			continue
		}
		if f.Mode().IsDir() {
			continue
		}

		b, err := ioutil.ReadFile(fName)
		if err != nil {
			log.Warnf("Unable to read configuration file %q: %s", fName, err)
			continue
		}
		m[f.Name()] = string(bytes.TrimSpace(b))
	}
	return m, nil
}

func BindEnv(optName string) {
	viper.BindEnv(optName, getEnvName(optName))
}

// getEnvName returns the environment variable to be used for the given option name.
func getEnvName(option string) string {
	under := strings.Replace(option, "-", "_", -1)
	upper := strings.ToUpper(under)
	return ciliumEnvPrefix + upper
}
