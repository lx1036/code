package defaults

const (
    // DefaultIPv4Prefix is the prefix for all the IPv4 addresses.
    // %d is substituted with the last byte of first global IPv4 address
    // configured on the system.
    DefaultIPv4Prefix = "10.%d.0.1"

    // DefaultIPv4PrefixLen is the length used to allocate container IPv4 addresses from.
    DefaultIPv4PrefixLen = 16

    // HostDevice is the name of the device that connects the cilium IP
    // space with the host's networking model
    HostDevice = "cilium_host"

    // SecondHostDevice is the name of the second interface of the host veth pair.
    SecondHostDevice = "cilium_net"

    // CiliumK8sAnnotationPrefix is the prefix key for the annotations used in kubernetes.
    CiliumK8sAnnotationPrefix = "cilium.io/"

    // AgentNotReadyNodeTaint is a node taint which prevents pods from being
    // scheduled. Once cilium is setup it is removed from the node. Mostly
    // used in cloud providers to prevent existing CNI plugins from managing
    // pods.
    AgentNotReadyNodeTaint = "node." + CiliumK8sAnnotationPrefix + "agent-not-ready"
)
