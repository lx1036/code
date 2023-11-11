package option

import (
	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/defaults"
	"github.com/cilium/cilium/pkg/lock"
	"github.com/cilium/cilium/pkg/metrics"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// AgentHealthPort is the TCP port for the agent health status API.
	AgentHealthPort = "agent-health-port"

	// AgentLabels are additional labels to identify this agent
	AgentLabels = "agent-labels"

	// AllowICMPFragNeeded allows ICMP Fragmentation Needed type packets in policy.
	AllowICMPFragNeeded = "allow-icmp-frag-needed"

	// AllowLocalhost is the policy when to allow local stack to reach local endpoints { auto | always | policy }
	AllowLocalhost = "allow-localhost"

	// AllowLocalhostAuto defaults to policy except when running in
	// Kubernetes where it then defaults to "always"
	AllowLocalhostAuto = "auto"

	// AllowLocalhostAlways always allows the local stack to reach local
	// endpoints
	AllowLocalhostAlways = "always"

	// AllowLocalhostPolicy requires a policy rule to allow the local stack
	// to reach particular endpoints or policy enforcement must be
	// disabled.
	AllowLocalhostPolicy = "policy"

	// AnnotateK8sNode enables annotating a kubernetes node while bootstrapping
	// the daemon, which can also be disbled using this option.
	AnnotateK8sNode = "annotate-k8s-node"

	// ARPPingRefreshPeriod is the ARP entries refresher period
	ARPPingRefreshPeriod = "arping-refresh-period"

	// EnableL2NeighDiscovery determines if cilium should perform L2 neighbor
	// discovery.
	EnableL2NeighDiscovery = "enable-l2-neigh-discovery"

	// BPFRoot is the Path to BPF filesystem
	BPFRoot = "bpf-root"

	// CertsDirectory is the root directory used to find out certificates used
	// in L7 HTTPs policy enforcement.
	CertsDirectory = "certificates-directory"

	// CGroupRoot is the path to Cgroup2 filesystem
	CGroupRoot = "cgroup-root"

	// CompilerFlags allow to specify extra compiler commands for advanced debugging
	CompilerFlags = "cflags"

	// ConfigFile is the Configuration file (default "$HOME/ciliumd.yaml")
	ConfigFile = "config"

	// ConfigDir is the directory that contains a file for each option where
	// the filename represents the option name and the content of that file
	// represents the value of that option.
	ConfigDir = "config-dir"

	// ConntrackGCInterval is the name of the ConntrackGCInterval option
	ConntrackGCInterval = "conntrack-gc-interval"

	// DebugArg is the argument enables debugging mode
	DebugArg = "debug"

	// DebugVerbose is the argument enables verbose log message for particular subsystems
	DebugVerbose = "debug-verbose"

	// Devices facing cluster/external network for attaching bpf_host
	Devices = "devices"

	// DirectRoutingDevice is the name of a device used to connect nodes in
	// direct routing mode (only required by BPF NodePort)
	DirectRoutingDevice = "direct-routing-device"

	// LBDevInheritIPAddr is device name which IP addr is inherited by devices
	// running BPF loadbalancer program
	LBDevInheritIPAddr = "bpf-lb-dev-ip-addr-inherit"

	// DisableConntrack disables connection tracking
	DisableConntrack = "disable-conntrack"

	// DisableEnvoyVersionCheck do not perform Envoy binary version check on startup
	DisableEnvoyVersionCheck = "disable-envoy-version-check"

	// EnablePolicy enables policy enforcement in the agent.
	EnablePolicy = "enable-policy"

	// EnableExternalIPs enables implementation of k8s services with externalIPs in datapath
	EnableExternalIPs = "enable-external-ips"

	// K8sEnableEndpointSlice enables the k8s EndpointSlice feature into Cilium
	K8sEnableEndpointSlice = "enable-k8s-endpoint-slice"

	// EnableL7Proxy is the name of the option to enable L7 proxy
	EnableL7Proxy = "enable-l7-proxy"

	// EnableTracing enables tracing mode in the agent.
	EnableTracing = "enable-tracing"

	// EncryptInterface enables encryption on specified interface
	EncryptInterface = "encrypt-interface"

	// EncryptNode enables node IP encryption
	EncryptNode = "encrypt-node"

	// EnvoyLog sets the path to a separate Envoy log file, if any
	EnvoyLog = "envoy-log"

	// GopsPort is the TCP port for the gops server.
	GopsPort = "gops-port"

	// ProxyPrometheusPort specifies the port to serve Cilium host proxy metrics on.
	ProxyPrometheusPort = "proxy-prometheus-port"

	// FixedIdentityMapping is the key-value for the fixed identity mapping
	// which allows to use reserved label for fixed identities
	FixedIdentityMapping = "fixed-identity-mapping"

	// IPv4Range is the per-node IPv4 endpoint prefix, e.g. 10.16.0.0/16
	IPv4Range = "ipv4-range"

	// IPv6Range is the per-node IPv6 endpoint prefix, must be /96, e.g. fd02:1:1::/96
	IPv6Range = "ipv6-range"

	// IPv4ServiceRange is the Kubernetes IPv4 services CIDR if not inside cluster prefix
	IPv4ServiceRange = "ipv4-service-range"

	// IPv6ServiceRange is the Kubernetes IPv6 services CIDR if not inside cluster prefix
	IPv6ServiceRange = "ipv6-service-range"

	// ModePreFilterNative for loading progs with xdpdrv
	ModePreFilterNative = XDPModeNative

	// ModePreFilterGeneric for loading progs with xdpgeneric
	ModePreFilterGeneric = XDPModeGeneric

	// IPv6ClusterAllocCIDRName is the name of the IPv6ClusterAllocCIDR option
	IPv6ClusterAllocCIDRName = "ipv6-cluster-alloc-cidr"

	// K8sRequireIPv4PodCIDRName is the name of the K8sRequireIPv4PodCIDR option
	K8sRequireIPv4PodCIDRName = "k8s-require-ipv4-pod-cidr"

	// K8sRequireIPv6PodCIDRName is the name of the K8sRequireIPv6PodCIDR option
	K8sRequireIPv6PodCIDRName = "k8s-require-ipv6-pod-cidr"

	// K8sForceJSONPatch when set, uses JSON Patch to update CNP and CEP
	// status in kube-apiserver.
	K8sForceJSONPatch = "k8s-force-json-patch"

	// K8sWatcherEndpointSelector specifies the k8s endpoints that Cilium
	// should watch for.
	K8sWatcherEndpointSelector = "k8s-watcher-endpoint-selector"

	// K8sAPIServer is the kubernetes api address server (for https use --k8s-kubeconfig-path instead)
	K8sAPIServer = "k8s-api-server"

	// K8sKubeConfigPath is the absolute path of the kubernetes kubeconfig file
	K8sKubeConfigPath = "k8s-kubeconfig-path"

	// K8sServiceCacheSize is service cache size for cilium k8s package.
	K8sServiceCacheSize = "k8s-service-cache-size"

	// K8sSyncTimeout is the timeout to synchronize all resources with k8s.
	K8sSyncTimeoutName = "k8s-sync-timeout"

	// AllocatorListTimeout is the timeout to list initial allocator state.
	AllocatorListTimeoutName = "allocator-list-timeout"

	// KeepConfig when restoring state, keeps containers' configuration in place
	KeepConfig = "keep-config"

	// KVStore key-value store type
	KVStore = "kvstore"

	// KVStoreOpt key-value store options
	KVStoreOpt = "kvstore-opt"

	// Labels is the list of label prefixes used to determine identity of an endpoint
	Labels = "labels"

	// LabelPrefixFile is the valid label prefixes file path
	LabelPrefixFile = "label-prefix-file"

	// EnableHostFirewall enables network policies for the host
	EnableHostFirewall = "enable-host-firewall"

	// EnableHostPort enables HostPort forwarding implemented by Cilium in BPF
	EnableHostPort = "enable-host-port"

	// EnableHostLegacyRouting enables the old routing path via stack.
	EnableHostLegacyRouting = "enable-host-legacy-routing"

	// EnableNodePort enables NodePort services implemented by Cilium in BPF
	EnableNodePort = "enable-node-port"

	// EnableSVCSourceRangeCheck enables check of service source range checks
	EnableSVCSourceRangeCheck = "enable-svc-source-range-check"

	// NodePortMode indicates in which mode NodePort implementation should run
	// ("snat", "dsr" or "hybrid")
	NodePortMode = "node-port-mode"

	// NodePortAlg indicates which algorithm is used for backend selection
	// ("random" or "maglev")
	NodePortAlg = "node-port-algorithm"

	// NodePortAcceleration indicates whether NodePort should be accelerated
	// via XDP ("none", "generic" or "native")
	NodePortAcceleration = "node-port-acceleration"

	// Alias to NodePortMode
	LoadBalancerMode = "bpf-lb-mode"

	// Alias to DSR dispatch method
	LoadBalancerDSRDispatch = "bpf-lb-dsr-dispatch"

	// Alias to DSR L4 translation method
	LoadBalancerDSRL4Xlate = "bpf-lb-dsr-l4-xlate"

	// Alias to DSR/IPIP IPv4 source CIDR
	LoadBalancerRSSv4CIDR = "bpf-lb-rss-ipv4-src-cidr"

	// Alias to DSR/IPIP IPv6 source CIDR
	LoadBalancerRSSv6CIDR = "bpf-lb-rss-ipv6-src-cidr"

	// Alias to NodePortAlg
	LoadBalancerAlg = "bpf-lb-algorithm"

	// Alias to NodePortAcceleration
	LoadBalancerAcceleration = "bpf-lb-acceleration"

	LoadBalancerPreserveWorldID = "bpf-lb-preserve-world-id"

	// MaglevTableSize determines the size of the backend table per service
	MaglevTableSize = "bpf-lb-maglev-table-size"

	// MaglevHashSeed contains the cluster-wide seed for the hash
	MaglevHashSeed = "bpf-lb-maglev-hash-seed"

	// NodePortBindProtection rejects bind requests to NodePort service ports
	NodePortBindProtection = "node-port-bind-protection"

	// NodePortRange defines a custom range where to look up NodePort services
	NodePortRange = "node-port-range"

	// EnableAutoProtectNodePortRange enables appending NodePort range to
	// net.ipv4.ip_local_reserved_ports if it overlaps with ephemeral port
	// range (net.ipv4.ip_local_port_range)
	EnableAutoProtectNodePortRange = "enable-auto-protect-node-port-range"

	// KubeProxyReplacement controls how to enable kube-proxy replacement
	// features in BPF datapath
	KubeProxyReplacement = "kube-proxy-replacement"

	// EnableSessionAffinity enables a support for service sessionAffinity
	EnableSessionAffinity = "enable-session-affinity"

	// EnableIdentityMark enables setting the mark field with the identity for
	// local traffic. This may be disabled if chaining modes and Cilium use
	// conflicting marks.
	EnableIdentityMark = "enable-identity-mark"

	// EnableBandwidthManager enables EDT-based pacing
	EnableBandwidthManager = "enable-bandwidth-manager"

	// EnableRecorder enables the datapath pcap recorder
	EnableRecorder = "enable-recorder"

	// EnableLocalRedirectPolicy enables support for local redirect policy
	EnableLocalRedirectPolicy = "enable-local-redirect-policy"

	// EnableMKE enables MKE specific 'chaining' for kube-proxy replacement
	EnableMKE = "enable-mke"

	// CgroupPathMKE points to the cgroupv1 net_cls mount instance
	CgroupPathMKE = "mke-cgroup-mount"

	// LibDir enables the directory path to store runtime build environment
	LibDir = "lib-dir"

	// LogDriver sets logging endpoints to use for example syslog, fluentd
	LogDriver = "log-driver"

	// LogOpt sets log driver options for cilium
	LogOpt = "log-opt"

	// Logstash enables logstash integration
	Logstash = "logstash"

	// NAT46Range is the IPv6 prefix to map IPv4 addresses to
	NAT46Range = "nat46-range"

	// Masquerade are the packets from endpoints leaving the host
	Masquerade = "masquerade"

	// EnableIPv4Masquerade masquerades IPv4 packets from endpoints leaving the host.
	EnableIPv4Masquerade = "enable-ipv4-masquerade"

	// EnableIPv6Masquerade masquerades IPv6 packets from endpoints leaving the host.
	EnableIPv6Masquerade = "enable-ipv6-masquerade"

	// EnableBPFClockProbe selects a more efficient source clock (jiffies vs ktime)
	EnableBPFClockProbe = "enable-bpf-clock-probe"

	// EnableBPFMasquerade masquerades packets from endpoints leaving the host with BPF instead of iptables
	EnableBPFMasquerade = "enable-bpf-masquerade"

	// DeriveMasqIPAddrFromDevice is device name which IP addr is used for BPF masquerades
	DeriveMasqIPAddrFromDevice = "derive-masquerade-ip-addr-from-device"

	// EnableIPMasqAgent enables BPF ip-masq-agent
	EnableIPMasqAgent = "enable-ip-masq-agent"

	// EnableEgressGateway enables egress-gateway
	EnableEgressGateway = "enable-egress-gateway"

	// IPMasqAgentConfigPath is the configuration file path
	IPMasqAgentConfigPath = "ip-masq-agent-config-path"

	// InstallIptRules sets whether Cilium should install any iptables in general
	InstallIptRules = "install-iptables-rules"

	// InstallNoConntrackIptRules instructs Cilium to install Iptables rules
	// to skip netfilter connection tracking on all pod traffic.
	InstallNoConntrackIptRules = "install-no-conntrack-iptables-rules"

	IPTablesLockTimeout = "iptables-lock-timeout"

	// IPTablesRandomFully sets iptables flag random-fully on masquerading rules
	IPTablesRandomFully = "iptables-random-fully"

	// IPv6NodeAddr is the IPv6 address of node
	IPv6NodeAddr = "ipv6-node"

	// IPv4NodeAddr is the IPv4 address of node
	IPv4NodeAddr = "ipv4-node"

	// Restore restores state, if possible, from previous daemon
	Restore = "restore"

	// SidecarIstioProxyImage regular expression matching compatible Istio sidecar istio-proxy container image names
	SidecarIstioProxyImage = "sidecar-istio-proxy-image"

	// SocketPath sets daemon's socket path to listen for connections
	SocketPath = "socket-path"

	// StateDir is the directory path to store runtime state
	StateDir = "state-dir"

	// TracePayloadlen length of payload to capture when tracing
	TracePayloadlen = "trace-payloadlen"

	// Version prints the version information
	Version = "version"

	// PProf enables serving the pprof debugging API
	PProf = "pprof"

	// PProfPort is the port that the pprof listens on
	PProfPort = "pprof-port"

	// EnableXDPPrefilter enables XDP-based prefiltering
	EnableXDPPrefilter = "enable-xdp-prefilter"

	// PrefilterDevice is the device facing external network for XDP prefiltering
	PrefilterDevice = "prefilter-device"

	// PrefilterMode { "+ModePreFilterNative+" | "+ModePreFilterGeneric+" } (default: "+option.ModePreFilterNative+")
	PrefilterMode = "prefilter-mode"

	// PrometheusServeAddr IP:Port on which to serve prometheus metrics (pass ":Port" to bind on all interfaces, "" is off)
	PrometheusServeAddr = "prometheus-serve-addr"

	// CMDRef is the path to cmdref output directory
	CMDRef = "cmdref"

	// DNSMaxIPsPerRestoredRule defines the maximum number of IPs to maintain
	// for each FQDN selector in endpoint's restored DNS rules
	DNSMaxIPsPerRestoredRule = "dns-max-ips-per-restored-rule"

	// ToFQDNsMinTTL is the minimum time, in seconds, to use DNS data for toFQDNs policies.
	ToFQDNsMinTTL = "tofqdns-min-ttl"

	// ToFQDNsProxyPort is the global port on which the in-agent DNS proxy should listen. Default 0 is a OS-assigned port.
	ToFQDNsProxyPort = "tofqdns-proxy-port"

	// ToFQDNsMaxIPsPerHost defines the maximum number of IPs to maintain
	// for each FQDN name in an endpoint's FQDN cache
	ToFQDNsMaxIPsPerHost = "tofqdns-endpoint-max-ip-per-hostname"

	// ToFQDNsMaxDeferredConnectionDeletes defines the maximum number of IPs to
	// retain for expired DNS lookups with still-active connections"
	ToFQDNsMaxDeferredConnectionDeletes = "tofqdns-max-deferred-connection-deletes"

	// ToFQDNsIdleConnectionGracePeriod defines the connection idle time during which
	// previously active connections with expired DNS lookups are still considered alive
	ToFQDNsIdleConnectionGracePeriod = "tofqdns-idle-connection-grace-period"

	// ToFQDNsPreCache is a path to a file with DNS cache data to insert into the
	// global cache on startup.
	// The file is not re-read after agent start.
	ToFQDNsPreCache = "tofqdns-pre-cache"

	// ToFQDNsEnableDNSCompression allows the DNS proxy to compress responses to
	// endpoints that are larger than 512 Bytes or the EDNS0 option, if present.
	ToFQDNsEnableDNSCompression = "tofqdns-enable-dns-compression"

	// DNSProxyConcurrencyLimit limits parallel processing of DNS messages in
	// DNS proxy at any given point in time.
	DNSProxyConcurrencyLimit = "dnsproxy-concurrency-limit"

	// DNSProxyConcurrencyProcessingGracePeriod is the amount of grace time to
	// wait while processing DNS messages when the DNSProxyConcurrencyLimit has
	// been reached.
	DNSProxyConcurrencyProcessingGracePeriod = "dnsproxy-concurrency-processing-grace-period"

	// MTUName is the name of the MTU option
	MTUName = "mtu"

	// DatapathMode is the name of the DatapathMode option
	DatapathMode = "datapath-mode"

	// IpvlanMasterDevice is the name of the IpvlanMasterDevice option
	IpvlanMasterDevice = "ipvlan-master-device"

	// EnableHostReachableServices is the name of the EnableHostReachableServices option
	EnableHostReachableServices = "enable-host-reachable-services"

	// BPFSocketLBHostnsOnly is the name of the BPFSocketLBHostnsOnly option
	BPFSocketLBHostnsOnly = "bpf-lb-sock-hostns-only"

	// HostReachableServicesProtos is the name of the HostReachableServicesProtos option
	HostReachableServicesProtos = "host-reachable-services-protos"

	// HostServicesTCP is the name of EnableHostServicesTCP config
	HostServicesTCP = "tcp"

	// HostServicesUDP is the name of EnableHostServicesUDP config
	HostServicesUDP = "udp"

	// TunnelName is the name of the Tunnel option
	TunnelName = "tunnel"

	// SingleClusterRouteName is the name of the SingleClusterRoute option
	//
	// SingleClusterRoute enables use of a single route covering the entire
	// cluster CIDR to point to the cilium_host interface instead of using
	// a separate route for each cluster node CIDR. This option is not
	// compatible with Tunnel=TunnelDisabled
	SingleClusterRouteName = "single-cluster-route"

	// MonitorAggregationName specifies the MonitorAggregationLevel on the
	// comandline.
	MonitorAggregationName = "monitor-aggregation"

	// MonitorAggregationInterval configures interval for monitor-aggregation
	MonitorAggregationInterval = "monitor-aggregation-interval"

	// MonitorAggregationFlags configures TCP flags used by monitor aggregation.
	MonitorAggregationFlags = "monitor-aggregation-flags"

	// ciliumEnvPrefix is the prefix used for environment variables
	ciliumEnvPrefix = "CILIUM_"

	// ClusterName is the name of the ClusterName option
	ClusterName = "cluster-name"

	// ClusterIDName is the name of the ClusterID option
	ClusterIDName = "cluster-id"

	// ClusterMeshConfigName is the name of the ClusterMeshConfig option
	ClusterMeshConfigName = "clustermesh-config"

	// CNIChainingMode configures which CNI plugin Cilium is chained with.
	CNIChainingMode = "cni-chaining-mode"

	// BPFCompileDebugName is the name of the option to enable BPF compiliation debugging
	BPFCompileDebugName = "bpf-compile-debug"

	// CTMapEntriesGlobalTCPDefault is the default maximum number of entries
	// in the TCP CT table.
	CTMapEntriesGlobalTCPDefault = 2 << 18 // 512Ki

	// CTMapEntriesGlobalAnyDefault is the default maximum number of entries
	// in the non-TCP CT table.
	CTMapEntriesGlobalAnyDefault = 2 << 17 // 256Ki

	// CTMapEntriesGlobalTCPName configures max entries for the TCP CT
	// table.
	CTMapEntriesGlobalTCPName = "bpf-ct-global-tcp-max"

	// CTMapEntriesGlobalAnyName configures max entries for the non-TCP CT
	// table.
	CTMapEntriesGlobalAnyName = "bpf-ct-global-any-max"

	// CTMapEntriesTimeout* name option and default value mappings
	CTMapEntriesTimeoutSYNName    = "bpf-ct-timeout-regular-tcp-syn"
	CTMapEntriesTimeoutFINName    = "bpf-ct-timeout-regular-tcp-fin"
	CTMapEntriesTimeoutTCPName    = "bpf-ct-timeout-regular-tcp"
	CTMapEntriesTimeoutAnyName    = "bpf-ct-timeout-regular-any"
	CTMapEntriesTimeoutSVCTCPName = "bpf-ct-timeout-service-tcp"
	CTMapEntriesTimeoutSVCAnyName = "bpf-ct-timeout-service-any"

	// NATMapEntriesGlobalDefault holds the default size of the NAT map
	// and is 2/3 of the full CT size as a heuristic
	NATMapEntriesGlobalDefault = int((CTMapEntriesGlobalTCPDefault + CTMapEntriesGlobalAnyDefault) * 2 / 3)

	// SockRevNATMapEntriesDefault holds the default size of the SockRev NAT map
	// and is the same size of CTMapEntriesGlobalAnyDefault as a heuristic given
	// that sock rev NAT is mostly used for UDP and getpeername only.
	SockRevNATMapEntriesDefault = CTMapEntriesGlobalAnyDefault

	// MapEntriesGlobalDynamicSizeRatioName is the name of the option to
	// set the ratio of total system memory to use for dynamic sizing of the
	// CT, NAT, Neighbor and SockRevNAT BPF maps.
	MapEntriesGlobalDynamicSizeRatioName = "bpf-map-dynamic-size-ratio"

	// LimitTableAutoGlobalTCPMin defines the minimum TCP CT table limit for
	// dynamic size ration calculation.
	LimitTableAutoGlobalTCPMin = 1 << 17 // 128Ki entries

	// LimitTableAutoGlobalAnyMin defines the minimum UDP CT table limit for
	// dynamic size ration calculation.
	LimitTableAutoGlobalAnyMin = 1 << 16 // 64Ki entries

	// LimitTableAutoNatGlobalMin defines the minimum NAT limit for dynamic size
	// ration calculation.
	LimitTableAutoNatGlobalMin = 1 << 17 // 128Ki entries

	// LimitTableAutoSockRevNatMin defines the minimum SockRevNAT limit for
	// dynamic size ration calculation.
	LimitTableAutoSockRevNatMin = 1 << 16 // 64Ki entries

	// LimitTableMin defines the minimum CT or NAT table limit
	LimitTableMin = 1 << 10 // 1Ki entries

	// LimitTableMax defines the maximum CT or NAT table limit
	LimitTableMax = 1 << 24 // 16Mi entries (~1GiB of entries per map)

	// PolicyMapMin defines the minimum policy map limit.
	PolicyMapMin = 1 << 8

	// PolicyMapMax defines the maximum policy map limit.
	PolicyMapMax = 1 << 16

	// FragmentsMapMin defines the minimum fragments map limit.
	FragmentsMapMin = 1 << 8

	// FragmentsMapMax defines the maximum fragments map limit.
	FragmentsMapMax = 1 << 16

	// NATMapEntriesGlobalName configures max entries for BPF NAT table
	NATMapEntriesGlobalName = "bpf-nat-global-max"

	// NeighMapEntriesGlobalName configures max entries for BPF neighbor table
	NeighMapEntriesGlobalName = "bpf-neigh-global-max"

	// PolicyMapEntriesName configures max entries for BPF policymap.
	PolicyMapEntriesName = "bpf-policy-map-max"

	// SockRevNatEntriesName configures max entries for BPF sock reverse nat
	// entries.
	SockRevNatEntriesName = "bpf-sock-rev-map-max"

	// LogSystemLoadConfigName is the name of the option to enable system
	// load loggging
	LogSystemLoadConfigName = "log-system-load"

	// PrependIptablesChainsName is the name of the option to enable
	// prepending iptables chains instead of appending
	PrependIptablesChainsName = "prepend-iptables-chains"

	// DisableCiliumEndpointCRDName is the name of the option to disable
	// use of the CEP CRD
	DisableCiliumEndpointCRDName = "disable-endpoint-crd"

	// MaxCtrlIntervalName and MaxCtrlIntervalNameEnv allow configuration
	// of MaxControllerInterval.
	MaxCtrlIntervalName = "max-controller-interval"

	// SockopsEnableName is the name of the option to enable sockops
	SockopsEnableName = "sockops-enable"

	// K8sNamespaceName is the name of the K8sNamespace option
	K8sNamespaceName = "k8s-namespace"

	// JoinClusterName is the name of the JoinCluster Option
	JoinClusterName = "join-cluster"

	// EnableIPv4Name is the name of the option to enable IPv4 support
	EnableIPv4Name = "enable-ipv4"

	// EnableIPv6Name is the name of the option to enable IPv6 support
	EnableIPv6Name = "enable-ipv6"

	// EnableIPv6NDPName is the name of the option to enable IPv6 NDP support
	EnableIPv6NDPName = "enable-ipv6-ndp"

	// IPv6MCastDevice is the name of the option to select IPv6 multicast device
	IPv6MCastDevice = "ipv6-mcast-device"

	// EnableMonitor is the name of the option to enable the monitor socket
	EnableMonitorName = "enable-monitor"

	// MonitorQueueSizeName is the name of the option MonitorQueueSize
	MonitorQueueSizeName = "monitor-queue-size"

	//FQDNRejectResponseCode is the name for the option for dns-proxy reject response code
	FQDNRejectResponseCode = "tofqdns-dns-reject-response-code"

	// FQDNProxyDenyWithNameError is useful when stub resolvers, like the one
	// in Alpine Linux's libc (musl), treat a REFUSED as a resolution error.
	// This happens when trying a DNS search list, as in kubernetes, and breaks
	// even whitelisted DNS names.
	FQDNProxyDenyWithNameError = "nameError"

	// FQDNProxyDenyWithRefused is the response code for Domain refused. It is
	// the default for denied DNS requests.
	FQDNProxyDenyWithRefused = "refused"

	// FQDNProxyResponseMaxDelay is the maximum time the proxy holds back a response
	FQDNProxyResponseMaxDelay = "tofqdns-proxy-response-max-delay"

	// FQDNRegexCompileLRUSize is the size of the FQDN regex compilation LRU.
	// Useful for heavy but repeated FQDN MatchName or MatchPattern use.
	FQDNRegexCompileLRUSize = "fqdn-regex-compile-lru-size"

	// PreAllocateMapsName is the name of the option PreAllocateMaps
	PreAllocateMapsName = "preallocate-bpf-maps"

	// EnableBPFTProxy option supports enabling or disabling BPF TProxy.
	EnableBPFTProxy = "enable-bpf-tproxy"

	// EnableXTSocketFallbackName is the name of the EnableXTSocketFallback option
	EnableXTSocketFallbackName = "enable-xt-socket-fallback"

	// EnableAutoDirectRoutingName is the name for the EnableAutoDirectRouting option
	EnableAutoDirectRoutingName = "auto-direct-node-routes"

	// EnableIPSecName is the name of the option to enable IPSec
	EnableIPSecName = "enable-ipsec"

	// IPSecKeyFileName is the name of the option for ipsec key file
	IPSecKeyFileName = "ipsec-key-file"

	// EnableWireguard is the name of the option to enable wireguard
	EnableWireguard = "enable-wireguard"

	// KVstoreLeaseTTL is the time-to-live for lease in kvstore.
	KVstoreLeaseTTL = "kvstore-lease-ttl"

	// KVstorePeriodicSync is the time interval in which periodic
	// synchronization with the kvstore occurs
	KVstorePeriodicSync = "kvstore-periodic-sync"

	// KVstoreConnectivityTimeout is the timeout when performing kvstore operations
	KVstoreConnectivityTimeout = "kvstore-connectivity-timeout"

	// IPAllocationTimeout is the timeout when allocating CIDRs
	IPAllocationTimeout = "ip-allocation-timeout"

	// IdentityChangeGracePeriod is the name of the
	// IdentityChangeGracePeriod option
	IdentityChangeGracePeriod = "identity-change-grace-period"

	// IdentityRestoreGracePeriod is the name of the
	// IdentityRestoreGracePeriod option
	IdentityRestoreGracePeriod = "identity-restore-grace-period"

	// EnableHealthChecking is the name of the EnableHealthChecking option
	EnableHealthChecking = "enable-health-checking"

	// EnableEndpointHealthChecking is the name of the EnableEndpointHealthChecking option
	EnableEndpointHealthChecking = "enable-endpoint-health-checking"

	// EnableHealthCheckNodePort is the name of the EnableHealthCheckNodePort option
	EnableHealthCheckNodePort = "enable-health-check-nodeport"

	// PolicyQueueSize is the size of the queues utilized by the policy
	// repository.
	PolicyQueueSize = "policy-queue-size"

	// EndpointQueueSize is the size of the EventQueue per-endpoint.
	EndpointQueueSize = "endpoint-queue-size"

	// EndpointGCInterval interval to attempt garbage collection of
	// endpoints that are no longer alive and healthy.
	EndpointGCInterval = "endpoint-gc-interval"

	// SelectiveRegeneration specifies whether only the endpoints which policy
	// changes select should be regenerated upon policy changes.
	SelectiveRegeneration = "enable-selective-regeneration"

	// K8sEventHandover is the name of the K8sEventHandover option
	K8sEventHandover = "enable-k8s-event-handover"

	// Metrics represents the metrics subsystem that Cilium should expose
	// to prometheus.
	Metrics = "metrics"

	// LoopbackIPv4 is the address to use for service loopback SNAT
	LoopbackIPv4 = "ipv4-service-loopback-address"

	// LocalRouterIPv4 is the link-local IPv4 address to use for Cilium router device
	LocalRouterIPv4 = "local-router-ipv4"

	// LocalRouterIPv6 is the link-local IPv6 address to use for Cilium router device
	LocalRouterIPv6 = "local-router-ipv6"

	// EndpointInterfaceNamePrefix is the prefix name of the interface
	// names shared by all endpoints
	EndpointInterfaceNamePrefix = "endpoint-interface-name-prefix"

	// ForceLocalPolicyEvalAtSource forces a policy decision at the source
	// endpoint for all local communication
	ForceLocalPolicyEvalAtSource = "force-local-policy-eval-at-source"

	// SkipCRDCreation specifies whether the CustomResourceDefinition will be
	// created by the daemon
	SkipCRDCreation = "skip-crd-creation"

	// EnableEndpointRoutes enables use of per endpoint routes
	EnableEndpointRoutes = "enable-endpoint-routes"

	// ExcludeLocalAddress excludes certain addresses to be recognized as a
	// local address
	ExcludeLocalAddress = "exclude-local-address"

	// IPv4PodSubnets A list of IPv4 subnets that pods may be
	// assigned from. Used with CNI chaining where IPs are not directly managed
	// by Cilium.
	IPv4PodSubnets = "ipv4-pod-subnets"

	// IPv6PodSubnets A list of IPv6 subnets that pods may be
	// assigned from. Used with CNI chaining where IPs are not directly managed
	// by Cilium.
	IPv6PodSubnets = "ipv6-pod-subnets"

	// IPAM is the IPAM method to use
	IPAM = "ipam"

	// XDPModeNative for loading progs with XDPModeLinkDriver
	XDPModeNative = "native"

	// XDPModeGeneric for loading progs with XDPModeLinkGeneric
	XDPModeGeneric = "testing-only"

	// XDPModeDisabled for not having XDP enabled
	XDPModeDisabled = "disabled"

	// XDPModeLinkDriver is the tc selector for native XDP
	XDPModeLinkDriver = "xdpdrv"

	// XDPModeLinkGeneric is the tc selector for generic XDP
	XDPModeLinkGeneric = "xdpgeneric"

	// XDPModeLinkNone for not having XDP enabled
	XDPModeLinkNone = XDPModeDisabled

	// K8sClientQPSLimit is the queries per second limit for the K8s client. Defaults to k8s client defaults.
	K8sClientQPSLimit = "k8s-client-qps"

	// K8sClientBurst is the burst value allowed for the K8s client. Defaults to k8s client defaults.
	K8sClientBurst = "k8s-client-burst"

	// AutoCreateCiliumNodeResource enables automatic creation of a
	// CiliumNode resource for the local node
	AutoCreateCiliumNodeResource = "auto-create-cilium-node-resource"

	// IPv4NativeRoutingCIDR describes a CIDR in which pod IPs are routable
	IPv4NativeRoutingCIDR = "native-routing-cidr"

	// EgressMasqueradeInterfaces is the selector used to select interfaces
	// subject to egress masquerading
	EgressMasqueradeInterfaces = "egress-masquerade-interfaces"

	// PolicyTriggerInterval is the amount of time between triggers of policy
	// updates are invoked.
	PolicyTriggerInterval = "policy-trigger-interval"

	// IdentityAllocationMode specifies what mode to use for identity
	// allocation
	IdentityAllocationMode = "identity-allocation-mode"

	// IdentityAllocationModeKVstore enables use of a key-value store such
	// as etcd or consul for identity allocation
	IdentityAllocationModeKVstore = "kvstore"

	// IdentityAllocationModeCRD enables use of Kubernetes CRDs for
	// identity allocation
	IdentityAllocationModeCRD = "crd"

	// DisableCNPStatusUpdates disables updating of CNP NodeStatus in the CNP
	// CRD.
	DisableCNPStatusUpdates = "disable-cnp-status-updates"

	// EnableLocalNodeRoute controls installation of the route which points
	// the allocation prefix of the local node.
	EnableLocalNodeRoute = "enable-local-node-route"

	// EnableWellKnownIdentities enables the use of well-known identities.
	// This is requires if identiy resolution is required to bring up the
	// control plane, e.g. when using the managed etcd feature
	EnableWellKnownIdentities = "enable-well-known-identities"

	// EnableRemoteNodeIdentity enables use of the remote-node identity
	EnableRemoteNodeIdentity = "enable-remote-node-identity"

	// PolicyAuditModeArg argument enables policy audit mode.
	PolicyAuditModeArg = "policy-audit-mode"

	// EnableHubble enables hubble in the agent.
	EnableHubble = "enable-hubble"

	// HubbleSocketPath specifies the UNIX domain socket for Hubble server to listen to.
	HubbleSocketPath = "hubble-socket-path"

	// HubbleListenAddress specifies address for Hubble server to listen to.
	HubbleListenAddress = "hubble-listen-address"

	// HubbleTLSDisabled allows the Hubble server to run on the given listen
	// address without TLS.
	HubbleTLSDisabled = "hubble-disable-tls"

	// HubbleTLSCertFile specifies the path to the public key file for the
	// Hubble server. The file must contain PEM encoded data.
	HubbleTLSCertFile = "hubble-tls-cert-file"

	// HubbleTLSKeyFile specifies the path to the private key file for the
	// Hubble server. The file must contain PEM encoded data.
	HubbleTLSKeyFile = "hubble-tls-key-file"

	// HubbleTLSClientCAFiles specifies the path to one or more client CA
	// certificates to use for TLS with mutual authentication (mTLS). The files
	// must contain PEM encoded data.
	HubbleTLSClientCAFiles = "hubble-tls-client-ca-files"

	// HubbleFlowBufferSize specifies the maximum number of flows in Hubble's buffer.
	// Deprecated: please, use HubbleEventBufferCapacity instead.
	HubbleFlowBufferSize = "hubble-flow-buffer-size"

	// HubbleEventBufferCapacity specifies the capacity of Hubble events buffer.
	HubbleEventBufferCapacity = "hubble-event-buffer-capacity"

	// HubbleEventQueueSize specifies the buffer size of the channel to receive monitor events.
	HubbleEventQueueSize = "hubble-event-queue-size"

	// HubbleMetricsServer specifies the addresses to serve Hubble metrics on.
	HubbleMetricsServer = "hubble-metrics-server"

	// HubbleMetrics specifies enabled metrics and their configuration options.
	HubbleMetrics = "hubble-metrics"

	// HubbleExportFilePath specifies the filepath to write Hubble events to.
	// e.g. "/var/run/cilium/hubble/events.log"
	HubbleExportFilePath = "hubble-export-file-path"

	// HubbleExportFileMaxSizeMB specifies the file size in MB at which to rotate
	// the Hubble export file.
	HubbleExportFileMaxSizeMB = "hubble-export-file-max-size-mb"

	// HubbleExportFileMaxBacks specifies the number of rotated files to keep.
	HubbleExportFileMaxBackups = "hubble-export-file-max-backups"

	// HubbleExportFileCompress specifies whether rotated files are compressed.
	HubbleExportFileCompress = "hubble-export-file-compress"

	// EnableHubbleRecorderAPI specifies if the Hubble Recorder API should be served
	EnableHubbleRecorderAPI = "enable-hubble-recorder-api"

	// HubbleRecorderStoragePath specifies the directory in which pcap files
	// created via the Hubble Recorder API are stored
	HubbleRecorderStoragePath = "hubble-recorder-storage-path"

	// HubbleRecorderSinkQueueSize is the queue size for each recorder sink
	HubbleRecorderSinkQueueSize = "hubble-recorder-sink-queue-size"

	// DisableIptablesFeederRules specifies which chains will be excluded
	// when installing the feeder rules
	DisableIptablesFeederRules = "disable-iptables-feeder-rules"

	// K8sHeartbeatTimeout configures the timeout for apiserver heartbeat
	K8sHeartbeatTimeout = "k8s-heartbeat-timeout"

	// EndpointStatus enables population of information in the
	// CiliumEndpoint.Status resource
	EndpointStatus = "endpoint-status"

	// EndpointStatusPolicy enables CiliumEndpoint.Status.Policy
	EndpointStatusPolicy = "policy"

	// EndpointStatusHealth enables CilliumEndpoint.Status.Health
	EndpointStatusHealth = "health"

	// EndpointStatusControllers enables CiliumEndpoint.Status.Controllers
	EndpointStatusControllers = "controllers"

	// EndpointStatusLog enables CiliumEndpoint.Status.Log
	EndpointStatusLog = "log"

	// EndpointStatusState enables CiliumEndpoint.Status.State
	EndpointStatusState = "state"

	// EnableIPv4FragmentsTrackingName is the name of the option to enable
	// IPv4 fragments tracking for L4-based lookups. Needs LRU map support.
	EnableIPv4FragmentsTrackingName = "enable-ipv4-fragment-tracking"

	// FragmentsMapEntriesName configures max entries for BPF fragments
	// tracking map.
	FragmentsMapEntriesName = "bpf-fragments-map-max"

	// K8sEnableAPIDiscovery enables Kubernetes API discovery
	K8sEnableAPIDiscovery = "enable-k8s-api-discovery"

	// LBMapEntriesName configures max entries for BPF lbmap.
	LBMapEntriesName = "bpf-lb-map-max"

	// K8sServiceProxyName instructs Cilium to handle service objects only when
	// service.kubernetes.io/service-proxy-name label equals the provided value.
	K8sServiceProxyName = "k8s-service-proxy-name"

	// APIRateLimitName enables configuration of the API rate limits
	APIRateLimitName = "api-rate-limit"

	// CRDWaitTimeout is the timeout in which Cilium will exit if CRDs are not
	// available.
	CRDWaitTimeout = "crd-wait-timeout"

	// EgressMultiHomeIPRuleCompat instructs Cilium to use a new scheme to
	// store rules and routes under ENI and Azure IPAM modes, if false.
	// Otherwise, it will use the old scheme.
	EgressMultiHomeIPRuleCompat = "egress-multi-home-ip-rule-compat"

	// EnableBPFBypassFIBLookup instructs Cilium to enable the FIB lookup bypass optimization for nodeport reverse NAT handling.
	EnableBPFBypassFIBLookup = "bpf-lb-bypass-fib-lookup"

	// EnableCustomCallsName is the name of the option to enable tail calls
	// for user-defined custom eBPF programs.
	EnableCustomCallsName = "enable-custom-calls"

	// BGPAnnounceLBIP announces service IPs of type LoadBalancer via BGP
	BGPAnnounceLBIP = "bgp-announce-lb-ip"

	// BGPConfigPath is the file path to the BGP configuration. It is
	// compatible with MetalLB's configuration.
	BGPConfigPath = "bgp-config-path"

	// ExternalClusterIPName is the name of the option to enable
	// cluster external access to ClusterIP services.
	ExternalClusterIPName = "bpf-lb-external-clusterip"

	// BypassIPAvailabilityUponRestore bypasses the IP availability error
	// within IPAM upon endpoint restore and allows the use of the restored IP
	// regardless of whether it's available in the pool.
	BypassIPAvailabilityUponRestore = "bypass-ip-availability-upon-restore"
)

const (
	// NodePortMinDefault is the minimal port to listen for NodePort requests
	NodePortMinDefault = 30000

	// NodePortMaxDefault is the maximum port to listen for NodePort requests
	NodePortMaxDefault = 32767

	// NodePortModeSNAT is for SNATing requests to remote nodes
	NodePortModeSNAT = "snat"

	// NodePortModeDSR is for performing DSR for requests to remote nodes
	NodePortModeDSR = "dsr"

	// NodePortAlgRandom is for randomly selecting a backend
	NodePortAlgRandom = "random"

	// NodePortAlgMaglev is for using maglev consistent hashing for backend selection
	NodePortAlgMaglev = "maglev"

	// NodePortModeHybrid is a dual mode of the above, that is, DSR for TCP and SNAT for UDP
	NodePortModeHybrid = "hybrid"

	// DSR dispatch mode to encode service into IP option or extension header
	DSRDispatchOption = "opt"

	// DSR dispatch mode to encapsulate to IPIP
	DSRDispatchIPIP = "ipip"

	// DSR L4 translation to frontend port
	DSRL4XlateFrontend = "frontend"

	// DSR L4 translation to backend port
	DSRL4XlateBackend = "backend"

	// NodePortAccelerationDisabled means we do not accelerate NodePort via XDP
	NodePortAccelerationDisabled = XDPModeDisabled

	// NodePortAccelerationGeneric means we accelerate NodePort via generic XDP
	NodePortAccelerationGeneric = XDPModeGeneric

	// NodePortAccelerationNative means we accelerate NodePort via native XDP in the driver (preferred)
	NodePortAccelerationNative = XDPModeNative

	// KubeProxyReplacementProbe specifies to auto-enable available features for
	// kube-proxy replacement
	KubeProxyReplacementProbe = "probe"

	// KubeProxyReplacementPartial specifies to enable only selected kube-proxy
	// replacement features (might panic)
	KubeProxyReplacementPartial = "partial"

	// KubeProxyReplacementStrict specifies to enable all kube-proxy replacement
	// features (might panic)
	KubeProxyReplacementStrict = "strict"

	// KubeProxyReplacementDisabled specified to completely disable kube-proxy
	// replacement
	KubeProxyReplacementDisabled = "disabled"

	// KubeProxyReplacement healthz server bind address
	KubeProxyReplacementHealthzBindAddr = "kube-proxy-replacement-healthz-bind-address"
)

var (
	// Config represents the daemon configuration
	Config = &DaemonConfig{
		CreationTime:                 time.Now(),
		Opts:                         NewIntOptions(&DaemonOptionLibrary),
		Monitor:                      &models.MonitorStatus{Cpus: int64(runtime.NumCPU()), Npages: 64, Pagesize: int64(os.Getpagesize()), Lost: 0, Unknown: 0},
		IPv6ClusterAllocCIDR:         defaults.IPv6ClusterAllocCIDR,
		IPv6ClusterAllocCIDRBase:     defaults.IPv6ClusterAllocCIDRBase,
		EnableHostIPRestore:          defaults.EnableHostIPRestore,
		EnableHealthChecking:         defaults.EnableHealthChecking,
		EnableEndpointHealthChecking: defaults.EnableEndpointHealthChecking,
		EnableHealthCheckNodePort:    defaults.EnableHealthCheckNodePort,
		EnableIPv4:                   defaults.EnableIPv4,
		EnableIPv6:                   defaults.EnableIPv6,
		EnableIPv6NDP:                defaults.EnableIPv6NDP,
		EnableL7Proxy:                defaults.EnableL7Proxy,
		EndpointStatus:               make(map[string]struct{}),
		DNSMaxIPsPerRestoredRule:     defaults.DNSMaxIPsPerRestoredRule,
		ToFQDNsMaxIPsPerHost:         defaults.ToFQDNsMaxIPsPerHost,
		KVstorePeriodicSync:          defaults.KVstorePeriodicSync,
		KVstoreConnectivityTimeout:   defaults.KVstoreConnectivityTimeout,
		IPAllocationTimeout:          defaults.IPAllocationTimeout,
		IdentityChangeGracePeriod:    defaults.IdentityChangeGracePeriod,
		IdentityRestoreGracePeriod:   defaults.IdentityRestoreGracePeriod,
		FixedIdentityMapping:         make(map[string]string),
		KVStoreOpt:                   make(map[string]string),
		LogOpt:                       make(map[string]string),
		SelectiveRegeneration:        defaults.SelectiveRegeneration,
		LoopbackIPv4:                 defaults.LoopbackIPv4,
		EndpointInterfaceNamePrefix:  defaults.EndpointInterfaceNamePrefix,
		ForceLocalPolicyEvalAtSource: defaults.ForceLocalPolicyEvalAtSource,
		EnableEndpointRoutes:         defaults.EnableEndpointRoutes,
		AnnotateK8sNode:              defaults.AnnotateK8sNode,
		K8sServiceCacheSize:          defaults.K8sServiceCacheSize,
		AutoCreateCiliumNodeResource: defaults.AutoCreateCiliumNodeResource,
		IdentityAllocationMode:       IdentityAllocationModeKVstore,
		AllowICMPFragNeeded:          defaults.AllowICMPFragNeeded,
		EnableWellKnownIdentities:    defaults.EnableWellKnownIdentities,
		K8sEnableK8sEndpointSlice:    defaults.K8sEnableEndpointSlice,
		k8sEnableAPIDiscovery:        defaults.K8sEnableAPIDiscovery,
		AllocatorListTimeout:         defaults.AllocatorListTimeout,

		k8sEnableLeasesFallbackDiscovery: defaults.K8sEnableLeasesFallbackDiscovery,
		APIRateLimit:                     make(map[string]string),

		ExternalClusterIP: defaults.ExternalClusterIP,
	}
)

// DaemonConfig is the configuration used by Daemon.
type DaemonConfig struct {
	CreationTime        time.Time
	BpfDir              string     // BPF template files directory
	LibDir              string     // Cilium library files directory
	RunDir              string     // Cilium runtime directory
	NAT46Prefix         *net.IPNet // NAT46 IPv6 Prefix
	Devices             []string   // bpf_host device
	DirectRoutingDevice string     // Direct routing device (used only by NodePort BPF)
	LBDevInheritIPAddr  string     // Device which IP addr used by bpf_host devices
	EnableXDPPrefilter  bool       // Enable XDP-based prefiltering
	DevicePreFilter     string     // Prefilter device
	ModePreFilter       string     // Prefilter mode
	XDPMode             string     // XDP mode, values: { xdpdrv | xdpgeneric | none }
	HostV4Addr          net.IP     // Host v4 address of the snooping device
	HostV6Addr          net.IP     // Host v6 address of the snooping device
	EncryptInterface    []string   // Set of network facing interface to encrypt over
	EncryptNode         bool       // Set to true for encrypting node IP traffic

	Ipvlan IpvlanConfig // Ipvlan related configuration

	DatapathMode string // Datapath mode
	Tunnel       string // Tunnel mode

	DryMode bool // Do not create BPF maps, devices, ..

	// RestoreState enables restoring the state from previous running daemons.
	RestoreState bool

	// EnableHostIPRestore enables restoring the host IPs based on state
	// left behind by previous Cilium runs.
	EnableHostIPRestore bool

	KeepConfig bool // Keep configuration of existing endpoints when starting up.

	// AllowLocalhost defines when to allows the local stack to local endpoints
	// values: { auto | always | policy }
	AllowLocalhost string

	// StateDir is the directory where runtime state of endpoints is stored
	StateDir string

	// Options changeable at runtime
	Opts *IntOptions

	// Mutex for serializing configuration updates to the daemon.
	ConfigPatchMutex lock.RWMutex

	// Monitor contains the configuration for the node monitor.
	Monitor *models.MonitorStatus

	// AgentHealthPort is the TCP port for the agent health status API.
	AgentHealthPort int

	// AgentLabels contains additional labels to identify this agent in monitor events.
	AgentLabels []string

	// IPv6ClusterAllocCIDR is the base CIDR used to allocate IPv6 node
	// CIDRs if allocation is not performed by an orchestration system
	IPv6ClusterAllocCIDR string

	// IPv6ClusterAllocCIDRBase is derived from IPv6ClusterAllocCIDR and
	// contains the CIDR without the mask, e.g. "fdfd::1/64" -> "fdfd::"
	//
	// This variable should never be written to, it is initialized via
	// DaemonConfig.Validate()
	IPv6ClusterAllocCIDRBase string

	// K8sRequireIPv4PodCIDR requires the k8s node resource to specify the
	// IPv4 PodCIDR. Cilium will block bootstrapping until the information
	// is available.
	K8sRequireIPv4PodCIDR bool

	// K8sRequireIPv6PodCIDR requires the k8s node resource to specify the
	// IPv6 PodCIDR. Cilium will block bootstrapping until the information
	// is available.
	K8sRequireIPv6PodCIDR bool

	// K8sServiceCacheSize is the service cache size for cilium k8s package.
	K8sServiceCacheSize uint

	// K8sForceJSONPatch when set, uses JSON Patch to update CNP and CEP
	// status in kube-apiserver.
	K8sForceJSONPatch bool

	// MTU is the maximum transmission unit of the underlying network
	MTU int

	// ClusterName is the name of the cluster
	ClusterName string

	// ClusterID is the unique identifier of the cluster
	ClusterID int

	// ClusterMeshConfig is the path to the clustermesh configuration directory
	ClusterMeshConfig string

	// CTMapEntriesGlobalTCP is the maximum number of conntrack entries
	// allowed in each TCP CT table for IPv4/IPv6.
	CTMapEntriesGlobalTCP int

	// CTMapEntriesGlobalAny is the maximum number of conntrack entries
	// allowed in each non-TCP CT table for IPv4/IPv6.
	CTMapEntriesGlobalAny int

	// CTMapEntriesTimeout* values configured by the user.
	CTMapEntriesTimeoutTCP    time.Duration
	CTMapEntriesTimeoutAny    time.Duration
	CTMapEntriesTimeoutSVCTCP time.Duration
	CTMapEntriesTimeoutSVCAny time.Duration
	CTMapEntriesTimeoutSYN    time.Duration
	CTMapEntriesTimeoutFIN    time.Duration

	// EnableMonitor enables the monitor unix domain socket server
	EnableMonitor bool

	// MonitorAggregationInterval configures the interval between monitor
	// messages when monitor aggregation is enabled.
	MonitorAggregationInterval time.Duration

	// MonitorAggregationFlags determines which TCP flags that the monitor
	// aggregation ensures reports are generated for when monitor-aggragation
	// is enabled. Network byte-order.
	MonitorAggregationFlags uint16

	// BPFMapsDynamicSizeRatio is ratio of total system memory to use for
	// dynamic sizing of the CT, NAT, Neighbor and SockRevNAT BPF maps.
	BPFMapsDynamicSizeRatio float64

	// NATMapEntriesGlobal is the maximum number of NAT mappings allowed
	// in the BPF NAT table
	NATMapEntriesGlobal int

	// NeighMapEntriesGlobal is the maximum number of neighbor mappings
	// allowed in the BPF neigh table
	NeighMapEntriesGlobal int

	// PolicyMapEntries is the maximum number of peer identities that an
	// endpoint may allow traffic to exchange traffic with.
	PolicyMapEntries int

	// SockRevNatEntries is the maximum number of sock rev nat mappings
	// allowed in the BPF rev nat table
	SockRevNatEntries int

	// DisableCiliumEndpointCRD disables the use of CiliumEndpoint CRD
	DisableCiliumEndpointCRD bool

	// MaxControllerInterval is the maximum value for a controller's
	// RunInterval. Zero means unlimited.
	MaxControllerInterval int

	// UseSingleClusterRoute specifies whether to use a single cluster route
	// instead of per-node routes.
	UseSingleClusterRoute bool

	// HTTPNormalizePath switches on Envoy HTTP path normalization options, which currently
	// includes RFC 3986 path normalization, Envoy merge slashes option, and unescaping and
	// redirecting for paths that contain escaped slashes. These are necessary to keep path based
	// access control functional, and should not interfere with normal operation. Set this to
	// false only with caution.
	HTTPNormalizePath bool

	// HTTP403Message is the error message to return when a HTTP 403 is returned
	// by the proxy, if L7 policy is configured.
	HTTP403Message string

	// HTTPRequestTimeout is the time in seconds after which Envoy responds with an
	// error code on a request that has not yet completed. This needs to be longer
	// than the HTTPIdleTimeout
	HTTPRequestTimeout int

	// HTTPIdleTimeout is the time in seconds of a HTTP stream having no traffic after
	// which Envoy responds with an error code. This needs to be shorter than the
	// HTTPRequestTimeout
	HTTPIdleTimeout int

	// HTTPMaxGRPCTimeout is the upper limit to which "grpc-timeout" headers in GRPC
	// requests are honored by Envoy. If 0 there is no limit. GRPC requests are not
	// bound by the HTTPRequestTimeout, but ARE affected by the idle timeout!
	HTTPMaxGRPCTimeout int

	// HTTPRetryCount is the upper limit on how many times Envoy retries failed requests.
	HTTPRetryCount int

	// HTTPRetryTimeout is the time in seconds before an uncompleted request is retried.
	HTTPRetryTimeout int

	// ProxyConnectTimeout is the time in seconds after which Envoy considers a TCP
	// connection attempt to have timed out.
	ProxyConnectTimeout int

	// ProxyGID specifies the group ID that has access to unix domain sockets opened by Cilium
	// agent for proxy configuration and access logging.
	ProxyGID int

	// ProxyPrometheusPort specifies the port to serve Envoy metrics on.
	ProxyPrometheusPort int

	// EnvoyLogPath specifies where to store the Envoy proxy logs when Envoy
	// runs in the same container as Cilium.
	EnvoyLogPath string

	// EnableSockOps specifies whether to enable sockops (socket lookup).
	SockopsEnable bool

	// PrependIptablesChains is the name of the option to enable prepending
	// iptables chains instead of appending
	PrependIptablesChains bool

	// IPTablesLockTimeout defines the "-w" iptables option when the
	// iptables CLI is directly invoked from the Cilium agent.
	IPTablesLockTimeout time.Duration

	// IPTablesRandomFully defines the "--random-fully" iptables option when the
	// iptables CLI is directly invoked from the Cilium agent.
	IPTablesRandomFully bool

	// K8sNamespace is the name of the namespace in which Cilium is
	// deployed in when running in Kubernetes mode
	K8sNamespace string

	// JoinCluster is 'true' if the agent should join a Cilium cluster via kvstore
	// registration
	JoinCluster bool

	// EnableIPv4 is true when IPv4 is enabled
	EnableIPv4 bool

	// EnableIPv6 is true when IPv6 is enabled
	EnableIPv6 bool

	// EnableIPv6NDP is true when NDP is enabled for IPv6
	EnableIPv6NDP bool

	// IPv6MCastDevice is the name of device that joins IPv6's solicitation multicast group
	IPv6MCastDevice string

	// EnableL7Proxy is the option to enable L7 proxy
	EnableL7Proxy bool

	// EnableIPSec is true when IPSec is enabled
	EnableIPSec bool

	// IPSec key file for stored keys
	IPSecKeyFile string

	// EnableWireguard enables Wireguard encryption
	EnableWireguard bool

	// MonitorQueueSize is the size of the monitor event queue
	MonitorQueueSize int

	// CLI options

	BPFRoot                       string
	BPFSocketLBHostnsOnly         bool
	CGroupRoot                    string
	BPFCompileDebug               string
	CompilerFlags                 []string
	ConfigFile                    string
	ConfigDir                     string
	Debug                         bool
	DebugVerbose                  []string
	DisableConntrack              bool
	EnableHostReachableServices   bool
	EnableHostServicesTCP         bool
	EnableHostServicesUDP         bool
	EnableHostServicesPeer        bool
	EnablePolicy                  string
	EnableTracing                 bool
	EnvoyLog                      string
	DisableEnvoyVersionCheck      bool
	FixedIdentityMapping          map[string]string
	FixedIdentityMappingValidator func(val string) (string, error) `json:"-"`
	IPv4Range                     string
	IPv6Range                     string
	IPv4ServiceRange              string
	IPv6ServiceRange              string
	K8sAPIServer                  string
	K8sKubeConfigPath             string
	K8sClientBurst                int
	K8sClientQPSLimit             float64
	K8sSyncTimeout                time.Duration
	AllocatorListTimeout          time.Duration
	K8sWatcherEndpointSelector    string
	KVStore                       string
	KVStoreOpt                    map[string]string
	LabelPrefixFile               string
	Labels                        []string
	LogDriver                     []string
	LogOpt                        map[string]string
	Logstash                      bool
	LogSystemLoadConfig           bool
	NAT46Range                    string

	// Masquerade specifies whether or not to masquerade packets from endpoints
	// leaving the host.
	EnableIPv4Masquerade       bool
	EnableIPv6Masquerade       bool
	EnableBPFMasquerade        bool
	DeriveMasqIPAddrFromDevice string
	EnableBPFClockProbe        bool
	EnableIPMasqAgent          bool
	EnableEgressGateway        bool
	IPMasqAgentConfigPath      string
	InstallIptRules            bool
	MonitorAggregation         string
	PreAllocateMaps            bool
	IPv6NodeAddr               string
	IPv4NodeAddr               string
	SidecarIstioProxyImage     string
	SocketPath                 string
	TracePayloadlen            int
	Version                    string
	PProf                      bool
	PProfPort                  int
	PrometheusServeAddr        string
	ToFQDNsMinTTL              int

	// DNSMaxIPsPerRestoredRule defines the maximum number of IPs to maintain
	// for each FQDN selector in endpoint's restored DNS rules
	DNSMaxIPsPerRestoredRule int

	// ToFQDNsProxyPort is the user-configured global, shared, DNS listen port used
	// by the DNS Proxy. Both UDP and TCP are handled on the same port. When it
	// is 0 a random port will be assigned, and can be obtained from
	// DefaultDNSProxy below.
	ToFQDNsProxyPort int

	// ToFQDNsMaxIPsPerHost defines the maximum number of IPs to maintain
	// for each FQDN name in an endpoint's FQDN cache
	ToFQDNsMaxIPsPerHost int

	// ToFQDNsMaxIPsPerHost defines the maximum number of IPs to retain for
	// expired DNS lookups with still-active connections
	ToFQDNsMaxDeferredConnectionDeletes int

	// ToFQDNsIdleConnectionGracePeriod Time during which idle but
	// previously active connections with expired DNS lookups are
	// still considered alive
	ToFQDNsIdleConnectionGracePeriod time.Duration

	// FQDNRejectResponse is the dns-proxy response for invalid dns-proxy request
	FQDNRejectResponse string

	// FQDNProxyResponseMaxDelay The maximum time the DNS proxy holds an allowed
	// DNS response before sending it along. Responses are sent as soon as the
	// datapath is updated with the new IP information.
	FQDNProxyResponseMaxDelay time.Duration

	// FQDNRegexCompileLRUSize is the size of the FQDN regex compilation LRU.
	// Useful for heavy but repeated FQDN MatchName or MatchPattern use.
	FQDNRegexCompileLRUSize int

	// Path to a file with DNS cache data to preload on startup
	ToFQDNsPreCache string

	// ToFQDNsEnableDNSCompression allows the DNS proxy to compress responses to
	// endpoints that are larger than 512 Bytes or the EDNS0 option, if present.
	ToFQDNsEnableDNSCompression bool

	// DNSProxyConcurrencyLimit limits parallel processing of DNS messages in
	// DNS proxy at any given point in time.
	DNSProxyConcurrencyLimit int

	// DNSProxyConcurrencyProcessingGracePeriod is the amount of grace time to
	// wait while processing DNS messages when the DNSProxyConcurrencyLimit has
	// been reached.
	DNSProxyConcurrencyProcessingGracePeriod time.Duration

	// HostDevice will be device used by Cilium to connect to the outside world.
	HostDevice string

	// EnableXTSocketFallback allows disabling of kernel's ip_early_demux
	// sysctl option if `xt_socket` kernel module is not available.
	EnableXTSocketFallback bool

	// EnableBPFTProxy enables implementing proxy redirection via BPF
	// mechanisms rather than iptables rules.
	EnableBPFTProxy bool

	// EnableAutoDirectRouting enables installation of direct routes to
	// other nodes when available
	EnableAutoDirectRouting bool

	// EnableLocalNodeRoute controls installation of the route which points
	// the allocation prefix of the local node.
	EnableLocalNodeRoute bool

	// EnableHealthChecking enables health checking between nodes and
	// health endpoints
	EnableHealthChecking bool

	// EnableEndpointHealthChecking enables health checking between virtual
	// health endpoints
	EnableEndpointHealthChecking bool

	// EnableHealthCheckNodePort enables health checking of NodePort by
	// cilium
	EnableHealthCheckNodePort bool

	// KVstoreKeepAliveInterval is the interval in which the lease is being
	// renewed. This must be set to a value lesser than the LeaseTTL ideally
	// by a factor of 3.
	KVstoreKeepAliveInterval time.Duration

	// KVstoreLeaseTTL is the time-to-live for kvstore lease.
	KVstoreLeaseTTL time.Duration

	// KVstorePeriodicSync is the time interval in which periodic
	// synchronization with the kvstore occurs
	KVstorePeriodicSync time.Duration

	// KVstoreConnectivityTimeout is the timeout when performing kvstore operations
	KVstoreConnectivityTimeout time.Duration

	// IPAllocationTimeout is the timeout when allocating CIDRs
	IPAllocationTimeout time.Duration

	// IdentityChangeGracePeriod is the grace period that needs to pass
	// before an endpoint that has changed its identity will start using
	// that new identity. During the grace period, the new identity has
	// already been allocated and other nodes in the cluster have a chance
	// to whitelist the new upcoming identity of the endpoint.
	IdentityChangeGracePeriod time.Duration

	// IdentityRestoreGracePeriod is the grace period that needs to pass before CIDR identities
	// restored during agent restart are released. If any of the restored identities remains
	// unused after this time, they will be removed from the IP cache. Any of the restored
	// identities that are used in network policies will remain in the IP cache until all such
	// policies are removed.
	IdentityRestoreGracePeriod time.Duration

	// PolicyQueueSize is the size of the queues for the policy repository.
	// A larger queue means that more events related to policy can be buffered.
	PolicyQueueSize int

	// EndpointQueueSize is the size of the EventQueue per-endpoint. A larger
	// queue means that more events can be buffered per-endpoint. This is useful
	// in the case where a cluster might be under high load for endpoint-related
	// events, specifically those which cause many regenerations.
	EndpointQueueSize int

	// EndpointGCInterval is interval to attempt garbage collection of
	// endpoints that are no longer alive and healthy.
	EndpointGCInterval time.Duration

	// SelectiveRegeneration, when true, enables the functionality to only
	// regenerate endpoints which are selected by the policy rules that have
	// been changed (added, deleted, or updated). If false, then all endpoints
	// are regenerated upon every policy change regardless of the scope of the
	// policy change.
	SelectiveRegeneration bool

	// ConntrackGCInterval is the connection tracking garbage collection
	// interval
	ConntrackGCInterval time.Duration

	// K8sEventHandover enables use of the kvstore to optimize Kubernetes
	// event handling by listening for k8s events in the operator and
	// mirroring it into the kvstore for reduced overhead in large
	// clusters.
	K8sEventHandover bool

	// MetricsConfig is the configuration set in metrics
	MetricsConfig metrics.Configuration

	// LoopbackIPv4 is the address to use for service loopback SNAT
	LoopbackIPv4 string

	// LocalRouterIPv4 is the link-local IPv4 address used for Cilium's router device
	LocalRouterIPv4 string

	// LocalRouterIPv6 is the link-local IPv6 address used for Cilium's router device
	LocalRouterIPv6 string

	// EndpointInterfaceNamePrefix is the prefix name of the interface
	// names shared by all endpoints
	EndpointInterfaceNamePrefix string

	// ForceLocalPolicyEvalAtSource forces a policy decision at the source
	// endpoint for all local communication
	ForceLocalPolicyEvalAtSource bool

	// SkipCRDCreation disables creation of the CustomResourceDefinition
	// on daemon startup
	// Deprecated: this option is not used by the cilium-agents anymore.
	SkipCRDCreation bool

	// EnableEndpointRoutes enables use of per endpoint routes
	EnableEndpointRoutes bool

	// Specifies wheather to annotate the kubernetes nodes or not
	AnnotateK8sNode bool

	// RunMonitorAgent indicates whether to run the monitor agent
	RunMonitorAgent bool

	// ReadCNIConfiguration reads the CNI configuration file and extracts
	// Cilium relevant information. This can be used to pass per node
	// configuration to Cilium.
	ReadCNIConfiguration string

	// WriteCNIConfigurationWhenReady writes the CNI configuration to the
	// specified location once the agent is ready to serve requests. This
	// allows to keep a Kubernetes node NotReady until Cilium is up and
	// running and able to schedule endpoints.
	WriteCNIConfigurationWhenReady string

	// EnableNodePort enables k8s NodePort service implementation in BPF
	EnableNodePort bool

	// EnableSVCSourceRangeCheck enables check of loadBalancerSourceRanges
	EnableSVCSourceRangeCheck bool

	// EnableHealthDatapath enables IPIP health probes data path
	EnableHealthDatapath bool

	// EnableHostPort enables k8s Pod's hostPort mapping through BPF
	EnableHostPort bool

	// EnableHostLegacyRouting enables the old routing path via stack.
	EnableHostLegacyRouting bool

	// NodePortMode indicates in which mode NodePort implementation should run
	// ("snat", "dsr" or "hybrid")
	NodePortMode string

	// NodePortAlg indicates which backend selection algorithm is used
	// ("random" or "maglev")
	NodePortAlg string

	// LoadBalancerDSRDispatch indicates the method for pushing packets to
	// backends under DSR ("opt" or "ipip")
	LoadBalancerDSRDispatch string

	// LoadBalancerDSRL4Xlate indicates the method for L4 DNAT translation
	// under IPIP dispatch, that is, whether the inner packet will be
	// translated to the frontend or backend port.
	LoadBalancerDSRL4Xlate string

	// LoadBalancerRSSv4CIDR defines the outer source IPv4 prefix for DSR/IPIP
	LoadBalancerRSSv4CIDR string
	LoadBalancerRSSv4     net.IPNet

	// LoadBalancerRSSv4CIDR defines the outer source IPv6 prefix for DSR/IPIP
	LoadBalancerRSSv6CIDR string
	LoadBalancerRSSv6     net.IPNet

	// LoadBalancerPMTUDiscovery indicates whether LB should reply with ICMP
	// frag needed messages to client (when needed)
	LoadBalancerPMTUDiscovery bool

	// LoadBalancerPreserveWorldID indicates whether world security ID should
	// be set for to be forwarded via tunnel LB requests
	LoadBalancerPreserveWorldID bool

	// Maglev backend table size (M) per service. Must be prime number.
	MaglevTableSize int

	// MaglevHashSeed contains the cluster-wide seed for the hash(es).
	MaglevHashSeed string

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

	// EnableBandwidthManager enables EDT-based pacing
	EnableBandwidthManager bool

	// ResetQueueMapping resets the Pod's skb queue mapping
	ResetQueueMapping bool

	// EnableRecorder enables the datapath pcap recorder
	EnableRecorder bool

	// EnableMKE enables MKE specific 'chaining' for kube-proxy replacement
	EnableMKE bool

	// CgroupPathMKE points to the cgroupv1 net_cls mount instance
	CgroupPathMKE string

	// KubeProxyReplacementHealthzBindAddr is the KubeProxyReplacement healthz server bind addr
	KubeProxyReplacementHealthzBindAddr string

	// EnableExternalIPs enables implementation of k8s services with externalIPs in datapath
	EnableExternalIPs bool

	// EnableHostFirewall enables network policies for the host
	EnableHostFirewall bool

	// EnableLocalRedirectPolicy enables redirect policies to redirect traffic within nodes
	EnableLocalRedirectPolicy bool

	// K8sEnableEndpointSlice enables k8s endpoint slice feature that is used
	// in kubernetes.
	K8sEnableK8sEndpointSlice bool

	// NodePortMin is the minimum port address for the NodePort range
	NodePortMin int

	// NodePortMax is the maximum port address for the NodePort range
	NodePortMax int

	// EnableSessionAffinity enables a support for service sessionAffinity
	EnableSessionAffinity bool

	// Selection of BPF main clock source (ktime vs jiffies)
	ClockSource BPFClockSource

	// EnableIdentityMark enables setting the mark field with the identity for
	// local traffic. This may be disabled if chaining modes and Cilium use
	// conflicting marks.
	EnableIdentityMark bool

	// KernelHz is the HZ rate the kernel is operating in
	KernelHz int

	// excludeLocalAddresses excludes certain addresses to be recognized as
	// a local address
	excludeLocalAddresses []*net.IPNet

	// IPv4PodSubnets available subnets to be assign IPv4 addresses to pods from
	IPv4PodSubnets []*net.IPNet

	// IPv6PodSubnets available subnets to be assign IPv6 addresses to pods from
	IPv6PodSubnets []*net.IPNet

	// IPAM is the IPAM method to use
	IPAM string

	// Enable chaining with another CNI plugin.
	CNIChainingMode string

	// AutoCreateCiliumNodeResource enables automatic creation of a
	// CiliumNode resource for the local node
	AutoCreateCiliumNodeResource bool

	// ipv4NativeRoutingCIDR describes a CIDR in which pod IPs are routable
	ipv4NativeRoutingCIDR *cidr.CIDR

	// EgressMasqueradeInterfaces is the selector used to select interfaces
	// subject to egress masquerading
	EgressMasqueradeInterfaces string

	// PolicyTriggerInterval is the amount of time between when policy updates
	// are triggered.
	PolicyTriggerInterval time.Duration

	// IdentityAllocationMode specifies what mode to use for identity
	// allocation
	IdentityAllocationMode string

	// DisableCNPStatusUpdates disables updating of CNP NodeStatus in the CNP
	// CRD.
	DisableCNPStatusUpdates bool

	// AllowICMPFragNeeded allows ICMP Fragmentation Needed type packets in
	// the network policy for cilium-agent.
	AllowICMPFragNeeded bool

	// EnableWellKnownIdentities enables the use of well-known identities.
	// This is requires if identiy resolution is required to bring up the
	// control plane, e.g. when using the managed etcd feature
	EnableWellKnownIdentities bool

	// CertsDirectory is the root directory to be used by cilium to find
	// certificates locally.
	CertDirectory string

	// EnableRemoteNodeIdentity enables use of the remote-node identity
	EnableRemoteNodeIdentity bool

	// Azure options

	// PolicyAuditMode enables non-drop mode for installed policies. In
	// audit mode packets affected by policies will not be dropped.
	// Policy related decisions can be checked via the poicy verdict messages.
	PolicyAuditMode bool

	// EnableHubble specifies whether to enable the hubble server.
	EnableHubble bool

	// HubbleSocketPath specifies the UNIX domain socket for Hubble server to listen to.
	HubbleSocketPath string

	// HubbleListenAddress specifies address for Hubble to listen to.
	HubbleListenAddress string

	// HubbleTLSDisabled allows the Hubble server to run on the given listen
	// address without TLS.
	HubbleTLSDisabled bool

	// HubbleTLSCertFile specifies the path to the public key file for the
	// Hubble server. The file must contain PEM encoded data.
	HubbleTLSCertFile string

	// HubbleTLSKeyFile specifies the path to the private key file for the
	// Hubble server. The file must contain PEM encoded data.
	HubbleTLSKeyFile string

	// HubbleTLSClientCAFiles specifies the path to one or more client CA
	// certificates to use for TLS with mutual authentication (mTLS). The files
	// must contain PEM encoded data.
	HubbleTLSClientCAFiles []string

	// HubbleFlowBufferSize specifies the maximum number of flows in Hubble's buffer.
	// Deprecated: please, use HubbleEventBufferCapacity instead.
	HubbleFlowBufferSize int

	// HubbleEventBufferCapacity specifies the capacity of Hubble events buffer.
	HubbleEventBufferCapacity int

	// HubbleEventQueueSize specifies the buffer size of the channel to receive monitor events.
	HubbleEventQueueSize int

	// HubbleMetricsServer specifies the addresses to serve Hubble metrics on.
	HubbleMetricsServer string

	// HubbleMetrics specifies enabled metrics and their configuration options.
	HubbleMetrics []string

	// HubbleExportFilePath specifies the filepath to write Hubble events to.
	// e.g. "/var/run/cilium/hubble/events.log"
	HubbleExportFilePath string

	// HubbleExportFileMaxSizeMB specifies the file size in MB at which to rotate
	// the Hubble export file.
	HubbleExportFileMaxSizeMB int

	// HubbleExportFileMaxBacks specifies the number of rotated files to keep.
	HubbleExportFileMaxBackups int

	// HubbleExportFileCompress specifies whether rotated files are compressed.
	HubbleExportFileCompress bool

	// EnableHubbleRecorderAPI specifies if the Hubble Recorder API should be served
	EnableHubbleRecorderAPI bool

	// HubbleRecorderStoragePath specifies the directory in which pcap files
	// created via the Hubble Recorder API are stored
	HubbleRecorderStoragePath string

	// HubbleRecorderSinkQueueSize is the queue size for each recorder sink
	HubbleRecorderSinkQueueSize int

	// K8sHeartbeatTimeout configures the timeout for apiserver heartbeat
	K8sHeartbeatTimeout time.Duration

	// EndpointStatus enables population of information in the
	// CiliumEndpoint.Status resource
	EndpointStatus map[string]struct{}

	// DisableIptablesFeederRules specifies which chains will be excluded
	// when installing the feeder rules
	DisableIptablesFeederRules []string

	// EnableIPv4FragmentsTracking enables IPv4 fragments tracking for
	// L4-based lookups. Needs LRU map support.
	EnableIPv4FragmentsTracking bool

	// FragmentsMapEntries is the maximum number of fragmented datagrams
	// that can simultaneously be tracked in order to retrieve their L4
	// ports for all fragments.
	FragmentsMapEntries int

	// sizeofCTElement is the size of an element (key + value) in the CT map.
	sizeofCTElement int

	// sizeofNATElement is the size of an element (key + value) in the NAT map.
	sizeofNATElement int

	// sizeofNeighElement is the size of an element (key + value) in the neigh
	// map.
	sizeofNeighElement int

	// sizeofSockRevElement is the size of an element (key + value) in the neigh
	// map.
	sizeofSockRevElement int

	k8sEnableAPIDiscovery bool

	// k8sEnableLeasesFallbackDiscovery enables k8s to fallback to API probing to check
	// for the support of Leases in Kubernetes when there is an error in discovering
	// API groups using Discovery API.
	// We require to check for Leases capabilities in operator only, which uses Leases for leader
	// election purposes in HA mode.
	// This is only enabled for cilium-operator
	k8sEnableLeasesFallbackDiscovery bool

	// LBMapEntries is the maximum number of entries allowed in BPF lbmap.
	LBMapEntries int

	// k8sServiceProxyName is the value of service.kubernetes.io/service-proxy-name label,
	// that identifies the service objects Cilium should handle.
	// If the provided value is an empty string, Cilium will manage service objects when
	// the label is not present. For more details -
	// https://github.com/kubernetes/enhancements/blob/master/keps/sig-network/0031-20181017-kube-proxy-services-optional.md
	k8sServiceProxyName string

	// APIRateLimitName enables configuration of the API rate limits
	APIRateLimit map[string]string

	// CRDWaitTimeout is the timeout in which Cilium will exit if CRDs are not
	// available.
	CRDWaitTimeout time.Duration

	// EgressMultiHomeIPRuleCompat instructs Cilium to use a new scheme to
	// store rules and routes under ENI and Azure IPAM modes, if false.
	// Otherwise, it will use the old scheme.
	EgressMultiHomeIPRuleCompat bool

	// EnableBPFBypassFIBLookup instructs Cilium to enable the FIB lookup bypass optimization for nodeport reverse NAT handling.
	EnableBPFBypassFIBLookup bool

	// InstallNoConntrackIptRules instructs Cilium to install Iptables rules to skip netfilter connection tracking on all pod traffic.
	InstallNoConntrackIptRules bool

	// EnableCustomCalls enables tail call hooks for user-defined custom
	// eBPF programs, typically used to collect custom per-endpoint
	// metrics.
	EnableCustomCalls bool

	// BGPAnnounceLBIP announces service IPs of type LoadBalancer via BGP.
	BGPAnnounceLBIP bool

	// BGPConfigPath is the file path to the BGP configuration. It is
	// compatible with MetalLB's configuration.
	BGPConfigPath string

	// ExternalClusterIP enables routing to ClusterIP services from outside
	// the cluster. This mirrors the behaviour of kube-proxy.
	ExternalClusterIP bool

	// ARPPingRefreshPeriod is the ARP entries refresher period.
	ARPPingRefreshPeriod time.Duration

	// EnableL2NeighDiscovery determines if cilium should perform L2 neighbor
	// discovery.
	EnableL2NeighDiscovery bool

	// BypassIPAvailabilityUponRestore bypasses the IP availability error
	// within IPAM upon endpoint restore and allows the use of the restored IP
	// regardless of whether it's available in the pool.
	BypassIPAvailabilityUponRestore bool
}

const nodeConfigFile = "node_config.h"

// GetNodeConfigPath /var/run/cilium/state/globals/node_config.h
func (c *DaemonConfig) GetNodeConfigPath() string {
	return filepath.Join(c.GetGlobalsDir(), nodeConfigFile)
}

func (c *DaemonConfig) GetGlobalsDir() string {
	return filepath.Join(c.StateDir, "globals")
}
