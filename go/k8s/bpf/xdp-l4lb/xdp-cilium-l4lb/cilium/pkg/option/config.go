package option

import (
    "fmt"
    "google.golang.org/appengine/log"
    "net"
    "os"
    "path/filepath"
    "runtime"
    "strconv"
    "strings"
    "time"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/cidr"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/command"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/common"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/defaults"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/ip"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/metrics"

    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
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

const (
    // TunnelVXLAN specifies VXLAN encapsulation
    TunnelVXLAN = "vxlan"

    // TunnelGeneve specifies Geneve encapsulation
    TunnelGeneve = "geneve"

    // TunnelDisabled specifies to disable encapsulation
    TunnelDisabled = "disabled"
)

const nodeConfigFile = "node_config.h"

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

    RegisteredOptions = map[string]struct{}{}
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

    devicesMu lock.RWMutex // Protects devices
    devices   []string     // bpf_host device

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

    // LBMapEntries is the maximum number of entries allowed in BPF lbmap.
    LBMapEntries int

    // LBServiceMapEntries is the maximum number of entries allowed in BPF lbmap for services.
    LBServiceMapEntries int

    // LBBackendMapEntries is the maximum number of entries allowed in BPF lbmap for service backends.
    LBBackendMapEntries int

    // LBRevNatEntries is the maximum number of entries allowed in BPF lbmap for reverse NAT.
    LBRevNatEntries int

    // LBAffinityMapEntries is the maximum number of entries allowed in BPF lbmap for session affinities.
    LBAffinityMapEntries int

    // LBSourceRangeMapEntries is the maximum number of entries allowed in BPF lbmap for source ranges.
    LBSourceRangeMapEntries int

    // LBMaglevMapEntries is the maximum number of entries allowed in BPF lbmap for maglev.
    LBMaglevMapEntries int

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

// Populate sets all options with the values from viper
func (c *DaemonConfig) Populate() {
    var err error

    c.AgentHealthPort = viper.GetInt(AgentHealthPort)
    c.ClusterHealthPort = viper.GetInt(ClusterHealthPort)
    c.ClusterMeshHealthPort = viper.GetInt(ClusterMeshHealthPort)
    c.AgentLabels = viper.GetStringSlice(AgentLabels)
    c.AllowICMPFragNeeded = viper.GetBool(AllowICMPFragNeeded)
    c.AllowLocalhost = viper.GetString(AllowLocalhost)
    c.AnnotateK8sNode = viper.GetBool(AnnotateK8sNode)
    c.ARPPingRefreshPeriod = viper.GetDuration(ARPPingRefreshPeriod)
    c.EnableL2NeighDiscovery = viper.GetBool(EnableL2NeighDiscovery)
    c.AutoCreateCiliumNodeResource = viper.GetBool(AutoCreateCiliumNodeResource)
    c.BPFRoot = viper.GetString(BPFRoot)
    c.CertDirectory = viper.GetString(CertsDirectory)
    c.CGroupRoot = viper.GetString(CGroupRoot)
    c.ClusterID = viper.GetInt(ClusterIDName)
    c.ClusterName = viper.GetString(ClusterName)
    c.ClusterMeshConfig = viper.GetString(ClusterMeshConfigName)
    c.CNIChainingMode = viper.GetString(CNIChainingMode)
    c.DatapathMode = viper.GetString(DatapathMode)
    c.Debug = viper.GetBool(DebugArg)
    c.DebugVerbose = viper.GetStringSlice(DebugVerbose)
    c.DirectRoutingDevice = viper.GetString(DirectRoutingDevice)
    c.LBDevInheritIPAddr = viper.GetString(LBDevInheritIPAddr)
    c.EnableIPv4 = viper.GetBool(EnableIPv4Name)
    c.EnableIPv6 = viper.GetBool(EnableIPv6Name)
    c.EnableIPv6NDP = viper.GetBool(EnableIPv6NDPName)
    c.IPv6MCastDevice = viper.GetString(IPv6MCastDevice)
    c.EnableIPSec = viper.GetBool(EnableIPSecName)
    c.EnableWireguard = viper.GetBool(EnableWireguard)
    c.EnableWireguardUserspaceFallback = viper.GetBool(EnableWireguardUserspaceFallback)
    c.EnableWellKnownIdentities = viper.GetBool(EnableWellKnownIdentities)
    c.EnableXDPPrefilter = viper.GetBool(EnableXDPPrefilter)
    c.DisableCiliumEndpointCRD = viper.GetBool(DisableCiliumEndpointCRDName)
    c.EgressMasqueradeInterfaces = viper.GetString(EgressMasqueradeInterfaces)
    c.BPFSocketLBHostnsOnly = viper.GetBool(BPFSocketLBHostnsOnly)
    c.EnableSocketLB = viper.GetBool(EnableHostReachableServices) || viper.GetBool(EnableSocketLB)
    c.EnableRemoteNodeIdentity = viper.GetBool(EnableRemoteNodeIdentity)
    c.K8sHeartbeatTimeout = viper.GetDuration(K8sHeartbeatTimeout)
    c.EnableBPFTProxy = viper.GetBool(EnableBPFTProxy)
    c.EnableXTSocketFallback = viper.GetBool(EnableXTSocketFallbackName)
    c.EnableAutoDirectRouting = viper.GetBool(EnableAutoDirectRoutingName)
    c.EnableEndpointRoutes = viper.GetBool(EnableEndpointRoutes)
    c.EnableHealthChecking = viper.GetBool(EnableHealthChecking)
    c.EnableEndpointHealthChecking = viper.GetBool(EnableEndpointHealthChecking)
    c.EnableHealthCheckNodePort = viper.GetBool(EnableHealthCheckNodePort)
    c.EnableLocalNodeRoute = viper.GetBool(EnableLocalNodeRoute)
    c.EnablePolicy = strings.ToLower(viper.GetString(EnablePolicy))
    c.EnableExternalIPs = viper.GetBool(EnableExternalIPs)
    c.EnableL7Proxy = viper.GetBool(EnableL7Proxy)
    c.EnableTracing = viper.GetBool(EnableTracing)
    c.EnableUnreachableRoutes = viper.GetBool(EnableUnreachableRoutes)
    c.EnableNodePort = viper.GetBool(EnableNodePort)
    c.EnableSVCSourceRangeCheck = viper.GetBool(EnableSVCSourceRangeCheck)
    c.EnableHostPort = viper.GetBool(EnableHostPort)
    c.EnableHostLegacyRouting = viper.GetBool(EnableHostLegacyRouting)
    c.MaglevTableSize = viper.GetInt(MaglevTableSize)
    c.MaglevHashSeed = viper.GetString(MaglevHashSeed)
    c.NodePortBindProtection = viper.GetBool(NodePortBindProtection)
    c.EnableAutoProtectNodePortRange = viper.GetBool(EnableAutoProtectNodePortRange)
    c.KubeProxyReplacement = viper.GetString(KubeProxyReplacement)
    c.EnableSessionAffinity = viper.GetBool(EnableSessionAffinity)
    c.EnableServiceTopology = viper.GetBool(EnableServiceTopology)
    c.EnableBandwidthManager = viper.GetBool(EnableBandwidthManager)
    c.EnableBBR = viper.GetBool(EnableBBR)
    c.EnableRecorder = viper.GetBool(EnableRecorder)
    c.EnableMKE = viper.GetBool(EnableMKE)
    c.CgroupPathMKE = viper.GetString(CgroupPathMKE)
    c.EnableHostFirewall = viper.GetBool(EnableHostFirewall)
    c.EnableLocalRedirectPolicy = viper.GetBool(EnableLocalRedirectPolicy)
    c.EncryptInterface = viper.GetStringSlice(EncryptInterface)
    c.EncryptNode = viper.GetBool(EncryptNode)
    c.EnvoyLogPath = viper.GetString(EnvoyLog)
    c.ForceLocalPolicyEvalAtSource = viper.GetBool(ForceLocalPolicyEvalAtSource)
    c.HTTPNormalizePath = viper.GetBool(HTTPNormalizePath)
    c.HTTPIdleTimeout = viper.GetInt(HTTPIdleTimeout)
    c.HTTPMaxGRPCTimeout = viper.GetInt(HTTPMaxGRPCTimeout)
    c.HTTPRequestTimeout = viper.GetInt(HTTPRequestTimeout)
    c.HTTPRetryCount = viper.GetInt(HTTPRetryCount)
    c.HTTPRetryTimeout = viper.GetInt(HTTPRetryTimeout)
    c.IdentityChangeGracePeriod = viper.GetDuration(IdentityChangeGracePeriod)
    c.IdentityRestoreGracePeriod = viper.GetDuration(IdentityRestoreGracePeriod)
    c.IPAM = viper.GetString(IPAM)
    c.IPv4Range = viper.GetString(IPv4Range)
    c.IPv4NodeAddr = viper.GetString(IPv4NodeAddr)
    c.IPv4ServiceRange = viper.GetString(IPv4ServiceRange)
    c.IPv6ClusterAllocCIDR = viper.GetString(IPv6ClusterAllocCIDRName)
    c.IPv6NodeAddr = viper.GetString(IPv6NodeAddr)
    c.IPv6Range = viper.GetString(IPv6Range)
    c.IPv6ServiceRange = viper.GetString(IPv6ServiceRange)
    c.JoinCluster = viper.GetBool(JoinClusterName)
    c.K8sAPIServer = viper.GetString(K8sAPIServer)
    c.K8sClientBurst = viper.GetInt(K8sClientBurst)
    c.K8sClientQPSLimit = viper.GetFloat64(K8sClientQPSLimit)
    c.K8sEnableK8sEndpointSlice = viper.GetBool(K8sEnableEndpointSlice)
    c.K8sEnableAPIDiscovery = viper.GetBool(K8sEnableAPIDiscovery)
    c.K8sKubeConfigPath = viper.GetString(K8sKubeConfigPath)
    c.K8sRequireIPv4PodCIDR = viper.GetBool(K8sRequireIPv4PodCIDRName)
    c.K8sRequireIPv6PodCIDR = viper.GetBool(K8sRequireIPv6PodCIDRName)
    c.K8sServiceCacheSize = uint(viper.GetInt(K8sServiceCacheSize))
    c.K8sEventHandover = viper.GetBool(K8sEventHandover)
    c.K8sSyncTimeout = viper.GetDuration(K8sSyncTimeoutName)
    c.AllocatorListTimeout = viper.GetDuration(AllocatorListTimeoutName)
    c.K8sWatcherEndpointSelector = viper.GetString(K8sWatcherEndpointSelector)
    c.KeepConfig = viper.GetBool(KeepConfig)
    c.KVStore = viper.GetString(KVStore)
    c.KVstoreLeaseTTL = viper.GetDuration(KVstoreLeaseTTL)
    c.KVstoreKeepAliveInterval = c.KVstoreLeaseTTL / defaults.KVstoreKeepAliveIntervalFactor
    c.KVstorePeriodicSync = viper.GetDuration(KVstorePeriodicSync)
    c.KVstoreConnectivityTimeout = viper.GetDuration(KVstoreConnectivityTimeout)
    c.KVstoreMaxConsecutiveQuorumErrors = viper.GetInt(KVstoreMaxConsecutiveQuorumErrorsName)
    c.IPAllocationTimeout = viper.GetDuration(IPAllocationTimeout)
    c.LabelPrefixFile = viper.GetString(LabelPrefixFile)
    c.Labels = viper.GetStringSlice(Labels)
    c.LibDir = viper.GetString(LibDir)
    c.LogDriver = viper.GetStringSlice(LogDriver)
    c.LogSystemLoadConfig = viper.GetBool(LogSystemLoadConfigName)
    c.Logstash = viper.GetBool(Logstash)
    c.LoopbackIPv4 = viper.GetString(LoopbackIPv4)
    c.LocalRouterIPv4 = viper.GetString(LocalRouterIPv4)
    c.LocalRouterIPv6 = viper.GetString(LocalRouterIPv6)
    c.EnableBPFClockProbe = viper.GetBool(EnableBPFClockProbe)
    c.EnableIPMasqAgent = viper.GetBool(EnableIPMasqAgent)
    c.EnableIPv4EgressGateway = viper.GetBool(EnableIPv4EgressGateway)
    c.InstallEgressGatewayRoutes = viper.GetBool(InstallEgressGatewayRoutes)
    c.EnableEnvoyConfig = viper.GetBool(EnableEnvoyConfig)
    c.EnableIngressController = viper.GetBool(EnableIngressController)
    c.EnvoyConfigTimeout = viper.GetDuration(EnvoyConfigTimeout)
    c.IPMasqAgentConfigPath = viper.GetString(IPMasqAgentConfigPath)
    c.InstallIptRules = viper.GetBool(InstallIptRules)
    c.IPTablesLockTimeout = viper.GetDuration(IPTablesLockTimeout)
    c.IPTablesRandomFully = viper.GetBool(IPTablesRandomFully)
    c.IPSecKeyFile = viper.GetString(IPSecKeyFileName)
    c.EnableMonitor = viper.GetBool(EnableMonitorName)
    c.MonitorAggregation = viper.GetString(MonitorAggregationName)
    c.MonitorAggregationInterval = viper.GetDuration(MonitorAggregationInterval)
    c.MonitorQueueSize = viper.GetInt(MonitorQueueSizeName)
    c.MTU = viper.GetInt(MTUName)
    c.PProf = viper.GetBool(PProf)
    c.PProfPort = viper.GetInt(PProfPort)
    c.PreAllocateMaps = viper.GetBool(PreAllocateMapsName)
    c.PrependIptablesChains = viper.GetBool(PrependIptablesChainsName)
    c.ProcFs = viper.GetString(ProcFs)
    c.PrometheusServeAddr = viper.GetString(PrometheusServeAddr)
    c.ProxyConnectTimeout = viper.GetInt(ProxyConnectTimeout)
    c.ProxyGID = viper.GetInt(ProxyGID)
    c.ProxyPrometheusPort = viper.GetInt(ProxyPrometheusPort)
    c.ProxyMaxRequestsPerConnection = viper.GetInt(ProxyMaxRequestsPerConnection)
    c.ProxyMaxConnectionDuration = time.Duration(viper.GetInt64(ProxyMaxConnectionDuration))
    c.ReadCNIConfiguration = viper.GetString(ReadCNIConfiguration)
    c.RestoreState = viper.GetBool(Restore)
    c.RouteMetric = viper.GetInt(RouteMetric)
    c.RunDir = viper.GetString(StateDir)
    c.SidecarIstioProxyImage = viper.GetString(SidecarIstioProxyImage)
    c.UseSingleClusterRoute = viper.GetBool(SingleClusterRouteName)
    c.SocketPath = viper.GetString(SocketPath)
    c.SockopsEnable = viper.GetBool(SockopsEnableName)
    c.TracePayloadlen = viper.GetInt(TracePayloadlen)
    c.Version = viper.GetString(Version)
    c.WriteCNIConfigurationWhenReady = viper.GetString(WriteCNIConfigurationWhenReady)
    c.PolicyTriggerInterval = viper.GetDuration(PolicyTriggerInterval)
    c.CTMapEntriesTimeoutTCP = viper.GetDuration(CTMapEntriesTimeoutTCPName)
    c.CTMapEntriesTimeoutAny = viper.GetDuration(CTMapEntriesTimeoutAnyName)
    c.CTMapEntriesTimeoutSVCTCP = viper.GetDuration(CTMapEntriesTimeoutSVCTCPName)
    c.CTMapEntriesTimeoutSVCTCPGrace = viper.GetDuration(CTMapEntriesTimeoutSVCTCPGraceName)
    c.CTMapEntriesTimeoutSVCAny = viper.GetDuration(CTMapEntriesTimeoutSVCAnyName)
    c.CTMapEntriesTimeoutSYN = viper.GetDuration(CTMapEntriesTimeoutSYNName)
    c.CTMapEntriesTimeoutFIN = viper.GetDuration(CTMapEntriesTimeoutFINName)
    c.PolicyAuditMode = viper.GetBool(PolicyAuditModeArg)
    c.EnableIPv4FragmentsTracking = viper.GetBool(EnableIPv4FragmentsTrackingName)
    c.FragmentsMapEntries = viper.GetInt(FragmentsMapEntriesName)
    c.K8sServiceProxyName = viper.GetString(K8sServiceProxyName)
    c.CRDWaitTimeout = viper.GetDuration(CRDWaitTimeout)
    c.LoadBalancerDSRDispatch = viper.GetString(LoadBalancerDSRDispatch)
    c.LoadBalancerDSRL4Xlate = viper.GetString(LoadBalancerDSRL4Xlate)
    c.LoadBalancerRSSv4CIDR = viper.GetString(LoadBalancerRSSv4CIDR)
    c.LoadBalancerRSSv6CIDR = viper.GetString(LoadBalancerRSSv6CIDR)
    c.InstallNoConntrackIptRules = viper.GetBool(InstallNoConntrackIptRules)
    c.EnableCustomCalls = viper.GetBool(EnableCustomCallsName)
    c.BGPAnnounceLBIP = viper.GetBool(BGPAnnounceLBIP)
    c.BGPAnnouncePodCIDR = viper.GetBool(BGPAnnouncePodCIDR)
    c.BGPConfigPath = viper.GetString(BGPConfigPath)
    c.ExternalClusterIP = viper.GetBool(ExternalClusterIPName)
    c.TCFilterPriority = viper.GetInt(TCFilterPriority)

    c.EnableIPv4Masquerade = viper.GetBool(EnableIPv4Masquerade) && c.EnableIPv4
    c.EnableIPv6Masquerade = viper.GetBool(EnableIPv6Masquerade) && c.EnableIPv6
    c.EnableBPFMasquerade = viper.GetBool(EnableBPFMasquerade)
    c.DeriveMasqIPAddrFromDevice = viper.GetString(DeriveMasqIPAddrFromDevice)

    c.populateLoadBalancerSettings()
    c.populateDevices()
    c.EnableRuntimeDeviceDetection = viper.GetBool(EnableRuntimeDeviceDetection)
    c.EgressMultiHomeIPRuleCompat = viper.GetBool(EgressMultiHomeIPRuleCompat)

    vlanBPFBypassIDs := viper.GetStringSlice(VLANBPFBypass)
    c.VLANBPFBypass = make([]int, 0, len(vlanBPFBypassIDs))
    for _, vlanIDStr := range vlanBPFBypassIDs {
        vlanID, err := strconv.Atoi(vlanIDStr)
        if err != nil {
            log.WithError(err).Fatalf("Cannot parse vlan ID integer from --%s option", VLANBPFBypass)
        }
        c.VLANBPFBypass = append(c.VLANBPFBypass, vlanID)
    }

    c.Tunnel = viper.GetString(TunnelName)
    c.TunnelPort = viper.GetInt(TunnelPortName)

    if c.TunnelPort == 0 {
        switch c.Tunnel {
        case TunnelDisabled:
            // tunnel might still be used by eg. EgressGW
            c.TunnelPort = defaults.TunnelPortVXLAN
        case TunnelVXLAN:
            c.TunnelPort = defaults.TunnelPortVXLAN
        case TunnelGeneve:
            c.TunnelPort = defaults.TunnelPortGeneve
        }
    }

    if viper.IsSet(AddressScopeMax) {
        c.AddressScopeMax, err = ip.ParseScope(viper.GetString(AddressScopeMax))
        if err != nil {
            log.WithError(err).Fatalf("Cannot parse scope integer from --%s option", AddressScopeMax)
        }
    } else {
        c.AddressScopeMax = defaults.AddressScopeMax
    }

    ipv4NativeRoutingCIDR := viper.GetString(IPv4NativeRoutingCIDR)

    if ipv4NativeRoutingCIDR != "" {
        c.IPv4NativeRoutingCIDR = cidr.MustParseCIDR(ipv4NativeRoutingCIDR)

        if len(c.IPv4NativeRoutingCIDR.IP) != net.IPv4len {
            log.Fatalf("%s must be an IPv4 CIDR", IPv4NativeRoutingCIDR)
        }
    }

    if c.EnableIPv4 && ipv4NativeRoutingCIDR == "" && c.EnableAutoDirectRouting {
        log.Warnf("If %s is enabled, then you are recommended to also configure %s. If %s is not configured, this may lead to pod to pod traffic being masqueraded, "+
            "which can cause problems with performance, observability and policy", EnableAutoDirectRoutingName, IPv4NativeRoutingCIDR, IPv4NativeRoutingCIDR)
    }

    ipv6NativeRoutingCIDR := viper.GetString(IPv6NativeRoutingCIDR)

    if ipv6NativeRoutingCIDR != "" {
        c.IPv6NativeRoutingCIDR = cidr.MustParseCIDR(ipv6NativeRoutingCIDR)

        if len(c.IPv6NativeRoutingCIDR.IP) != net.IPv6len {
            log.Fatalf("%s must be an IPv6 CIDR", IPv6NativeRoutingCIDR)
        }
    }

    if c.EnableIPv6 && ipv6NativeRoutingCIDR == "" && c.EnableAutoDirectRouting {
        log.Warnf("If %s is enabled, then you are recommended to also configure %s. If %s is not configured, this may lead to pod to pod traffic being masqueraded, "+
            "which can cause problems with performance, observability and policy", EnableAutoDirectRoutingName, IPv6NativeRoutingCIDR, IPv6NativeRoutingCIDR)
    }

    if err := c.calculateBPFMapSizes(); err != nil {
        log.Fatal(err)
    }

    c.ClockSource = ClockSourceKtime
    c.EnableIdentityMark = viper.GetBool(EnableIdentityMark)

    // toFQDNs options
    c.DNSMaxIPsPerRestoredRule = viper.GetInt(DNSMaxIPsPerRestoredRule)
    c.DNSPolicyUnloadOnShutdown = viper.GetBool(DNSPolicyUnloadOnShutdown)
    c.FQDNRegexCompileLRUSize = viper.GetInt(FQDNRegexCompileLRUSize)
    c.ToFQDNsMaxIPsPerHost = viper.GetInt(ToFQDNsMaxIPsPerHost)
    if maxZombies := viper.GetInt(ToFQDNsMaxDeferredConnectionDeletes); maxZombies >= 0 {
        c.ToFQDNsMaxDeferredConnectionDeletes = viper.GetInt(ToFQDNsMaxDeferredConnectionDeletes)
    } else {
        log.Fatalf("%s must be positive, or 0 to disable deferred connection deletion",
            ToFQDNsMaxDeferredConnectionDeletes)
    }
    switch {
    case viper.IsSet(ToFQDNsMinTTL): // set by user
        c.ToFQDNsMinTTL = viper.GetInt(ToFQDNsMinTTL)
    default:
        c.ToFQDNsMinTTL = defaults.ToFQDNsMinTTL
    }
    c.ToFQDNsProxyPort = viper.GetInt(ToFQDNsProxyPort)
    c.ToFQDNsPreCache = viper.GetString(ToFQDNsPreCache)
    c.ToFQDNsEnableDNSCompression = viper.GetBool(ToFQDNsEnableDNSCompression)
    c.DNSProxyConcurrencyLimit = viper.GetInt(DNSProxyConcurrencyLimit)
    c.DNSProxyConcurrencyProcessingGracePeriod = viper.GetDuration(DNSProxyConcurrencyProcessingGracePeriod)

    // Convert IP strings into net.IPNet types
    subnets, invalid := ip.ParseCIDRs(viper.GetStringSlice(IPv4PodSubnets))
    if len(invalid) > 0 {
        log.WithFields(
            logrus.Fields{
                "Subnets": invalid,
            }).Warning("IPv4PodSubnets parameter can not be parsed.")
    }
    c.IPv4PodSubnets = subnets

    subnets, invalid = ip.ParseCIDRs(viper.GetStringSlice(IPv6PodSubnets))
    if len(invalid) > 0 {
        log.WithFields(
            logrus.Fields{
                "Subnets": invalid,
            }).Warning("IPv6PodSubnets parameter can not be parsed.")
    }
    c.IPv6PodSubnets = subnets

    c.XDPMode = XDPModeLinkNone

    err = c.populateNodePortRange()
    if err != nil {
        log.WithError(err).Fatal("Failed to populate NodePortRange")
    }

    err = c.populateHostServicesProtos()
    if err != nil {
        log.WithError(err).Fatal("Failed to populate HostReachableServicesProtos")
    }

    monitorAggregationFlags := viper.GetStringSlice(MonitorAggregationFlags)
    var ctMonitorReportFlags uint16
    for i := 0; i < len(monitorAggregationFlags); i++ {
        value := strings.ToLower(monitorAggregationFlags[i])
        flag, exists := TCPFlags[value]
        if !exists {
            log.Fatalf("Unable to parse TCP flag %q for %s!",
                value, MonitorAggregationFlags)
        }
        ctMonitorReportFlags |= flag
    }
    c.MonitorAggregationFlags = ctMonitorReportFlags

    // Map options
    if m := command.GetStringMapString(viper.GetViper(), FixedIdentityMapping); err != nil {
        log.Fatalf("unable to parse %s: %s", FixedIdentityMapping, err)
    } else if len(m) != 0 {
        c.FixedIdentityMapping = m
    }

    c.ConntrackGCInterval = viper.GetDuration(ConntrackGCInterval)

    if m, err := command.GetStringMapStringE(viper.GetViper(), KVStoreOpt); err != nil {
        log.Fatalf("unable to parse %s: %s", KVStoreOpt, err)
    } else {
        c.KVStoreOpt = m
    }

    if m, err := command.GetStringMapStringE(viper.GetViper(), LogOpt); err != nil {
        log.Fatalf("unable to parse %s: %s", LogOpt, err)
    } else {
        c.LogOpt = m
    }

    if m, err := command.GetStringMapStringE(viper.GetViper(), APIRateLimitName); err != nil {
        log.Fatalf("unable to parse %s: %s", APIRateLimitName, err)
    } else {
        c.APIRateLimit = m
    }

    for _, option := range viper.GetStringSlice(EndpointStatus) {
        c.EndpointStatus[option] = struct{}{}
    }

    if c.MonitorQueueSize == 0 {
        c.MonitorQueueSize = getDefaultMonitorQueueSize(runtime.NumCPU())
    }

    // Metrics Setup
    defaultMetrics := metrics.DefaultMetrics()
    flagMetrics := append(viper.GetStringSlice(Metrics), c.additionalMetrics()...)
    for _, metric := range flagMetrics {
        switch metric[0] {
        case '+':
            defaultMetrics[metric[1:]] = struct{}{}
        case '-':
            delete(defaultMetrics, metric[1:])
        }
    }
    var collectors []prometheus.Collector
    metricsSlice := common.MapStringStructToSlice(defaultMetrics)
    c.MetricsConfig, collectors = metrics.CreateConfiguration(metricsSlice)
    metrics.MustRegister(collectors...)

    if err := c.parseExcludedLocalAddresses(viper.GetStringSlice(ExcludeLocalAddress)); err != nil {
        log.WithError(err).Fatalf("Unable to parse excluded local addresses")
    }

    c.IdentityAllocationMode = viper.GetString(IdentityAllocationMode)
    switch c.IdentityAllocationMode {
    // This is here for tests. Some call Populate without the normal init
    case "":
        c.IdentityAllocationMode = IdentityAllocationModeKVstore

    case IdentityAllocationModeKVstore, IdentityAllocationModeCRD:
        // c.IdentityAllocationMode is set above

    default:
        log.Fatalf("Invalid identity allocation mode %q. It must be one of %s or %s", c.IdentityAllocationMode, IdentityAllocationModeKVstore, IdentityAllocationModeCRD)
    }
    if c.KVStore == "" {
        if c.IdentityAllocationMode != IdentityAllocationModeCRD {
            log.Warningf("Running Cilium with %q=%q requires identity allocation via CRDs. Changing %s to %q", KVStore, c.KVStore, IdentityAllocationMode, IdentityAllocationModeCRD)
            c.IdentityAllocationMode = IdentityAllocationModeCRD
        }
        if c.DisableCiliumEndpointCRD {
            log.Warningf("Running Cilium with %q=%q requires endpoint CRDs. Changing %s to %t", KVStore, c.KVStore, DisableCiliumEndpointCRDName, false)
            c.DisableCiliumEndpointCRD = false
        }
        if c.K8sEventHandover {
            log.Warningf("Running Cilium with %q=%q requires KVStore capability. Changing %s to %t", KVStore, c.KVStore, K8sEventHandover, false)
            c.K8sEventHandover = false
        }
    }

    switch c.IPAM {
    case ipamOption.IPAMKubernetes, ipamOption.IPAMClusterPool, ipamOption.IPAMClusterPoolV2:
        if c.EnableIPv4 {
            c.K8sRequireIPv4PodCIDR = true
        }

        if c.EnableIPv6 {
            c.K8sRequireIPv6PodCIDR = true
        }
    }

    c.KubeProxyReplacementHealthzBindAddr = viper.GetString(KubeProxyReplacementHealthzBindAddr)

    // Hubble options.
    c.EnableHubble = viper.GetBool(EnableHubble)
    c.HubbleSocketPath = viper.GetString(HubbleSocketPath)
    c.HubbleListenAddress = viper.GetString(HubbleListenAddress)
    c.HubbleTLSDisabled = viper.GetBool(HubbleTLSDisabled)
    c.HubbleTLSCertFile = viper.GetString(HubbleTLSCertFile)
    c.HubbleTLSKeyFile = viper.GetString(HubbleTLSKeyFile)
    c.HubbleTLSClientCAFiles = viper.GetStringSlice(HubbleTLSClientCAFiles)
    c.HubbleEventBufferCapacity = viper.GetInt(HubbleEventBufferCapacity)
    c.HubbleEventQueueSize = viper.GetInt(HubbleEventQueueSize)
    if c.HubbleEventQueueSize == 0 {
        c.HubbleEventQueueSize = getDefaultMonitorQueueSize(runtime.NumCPU())
    }
    c.HubbleMetricsServer = viper.GetString(HubbleMetricsServer)
    c.HubbleMetrics = viper.GetStringSlice(HubbleMetrics)
    c.HubbleExportFilePath = viper.GetString(HubbleExportFilePath)
    c.HubbleExportFileMaxSizeMB = viper.GetInt(HubbleExportFileMaxSizeMB)
    c.HubbleExportFileMaxBackups = viper.GetInt(HubbleExportFileMaxBackups)
    c.HubbleExportFileCompress = viper.GetBool(HubbleExportFileCompress)
    c.EnableHubbleRecorderAPI = viper.GetBool(EnableHubbleRecorderAPI)
    c.HubbleRecorderStoragePath = viper.GetString(HubbleRecorderStoragePath)
    c.HubbleRecorderSinkQueueSize = viper.GetInt(HubbleRecorderSinkQueueSize)
    c.DisableIptablesFeederRules = viper.GetStringSlice(DisableIptablesFeederRules)
    c.EnableCiliumEndpointSlice = viper.GetBool(EnableCiliumEndpointSlice)

    // Hidden options
    c.CompilerFlags = viper.GetStringSlice(CompilerFlags)
    c.ConfigFile = viper.GetString(ConfigFile)
    c.HTTP403Message = viper.GetString(HTTP403Message)
    c.K8sNamespace = viper.GetString(K8sNamespaceName)
    c.AgentNotReadyNodeTaintKey = viper.GetString(AgentNotReadyNodeTaintKeyName)
    c.MaxControllerInterval = viper.GetInt(MaxCtrlIntervalName)
    c.PolicyQueueSize = sanitizeIntParam(PolicyQueueSize, defaults.PolicyQueueSize)
    c.EndpointQueueSize = sanitizeIntParam(EndpointQueueSize, defaults.EndpointQueueSize)
    c.EndpointGCInterval = viper.GetDuration(EndpointGCInterval)
    c.SelectiveRegeneration = viper.GetBool(SelectiveRegeneration)
    c.DisableCNPStatusUpdates = viper.GetBool(DisableCNPStatusUpdates)
    c.EnableICMPRules = viper.GetBool(EnableICMPRules)
    c.BypassIPAvailabilityUponRestore = viper.GetBool(BypassIPAvailabilityUponRestore)
    c.EnableK8sTerminatingEndpoint = viper.GetBool(EnableK8sTerminatingEndpoint)

    // Disable Envoy version check if L7 proxy is disabled.
    c.DisableEnvoyVersionCheck = viper.GetBool(DisableEnvoyVersionCheck)
    if !c.EnableL7Proxy {
        c.DisableEnvoyVersionCheck = true
    }

    // VTEP integration enable option
    c.EnableVTEP = viper.GetBool(EnableVTEP)

    // Enable BGP control plane features
    c.EnableBGPControlPlane = viper.GetBool(EnableBGPControlPlane)

    // Envoy secrets namespace to watch
    c.EnvoySecretNamespace = viper.GetString(IngressSecretsNamespace)
}

func (c *DaemonConfig) Validate() error {

    return nil
}

func (c *DaemonConfig) SetDevices(devices []string) {
    c.devicesMu.Lock()
    c.devices = devices
    c.devicesMu.Unlock()
}

func (c *DaemonConfig) AppendDevice(dev string) {
    c.devicesMu.Lock()
    c.devices = append(c.devices, dev)
    c.devicesMu.Unlock()
}

func (c *DaemonConfig) GetDevices() []string {
    c.devicesMu.RLock()
    defer c.devicesMu.RUnlock()
    return c.devices
}

// GetNodeConfigPath /var/run/cilium/state/globals/node_config.h
func (c *DaemonConfig) GetNodeConfigPath() string {
    return filepath.Join(c.GetGlobalsDir(), nodeConfigFile)
}

func (c *DaemonConfig) GetGlobalsDir() string {
    return filepath.Join(c.StateDir, "globals")
}

func BindEnv(optName string) {
    registerOpt(optName)
    viper.BindEnv(optName, getEnvName(optName))
}

func registerOpt(optName string) {
    _, ok := RegisteredOptions[optName]
    if ok || optName == "" {
        panic(fmt.Errorf("option already registered: %s", optName))
    }
    RegisteredOptions[optName] = struct{}{}
}

func getEnvName(option string) string {
    under := strings.Replace(option, "-", "_", -1)
    upper := strings.ToUpper(under)
    return ciliumEnvPrefix + upper
}

func InitConfig(cmd *cobra.Command, programName, configName string) func() {
    return func() {
        Config.ConfigFile = viper.GetString(ConfigFile) // enable ability to specify config file via flag
        Config.ConfigDir = viper.GetString(ConfigDir)
        viper.SetEnvPrefix("cilium")

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
            log.WithError(err).Debug("Skipped reading configuration file")
        }
    }
}
