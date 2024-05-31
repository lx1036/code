package option

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
