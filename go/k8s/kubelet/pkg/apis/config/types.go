package config

// hairpinMode specifies how the Kubelet should configure the container
// bridge for hairpin packets.
// Setting this flag allows endpoints in a Service to loadbalance back to
// themselves if they should try to access their own Service. Values:
//   "promiscuous-bridge": make the container bridge promiscuous.
//   "hairpin-veth":       set the hairpin flag on container veth interfaces.
//   "none":               do nothing.
// Generally, one must set --hairpin-mode=hairpin-veth to achieve hairpin NAT,
// because promiscuous-bridge assumes the existence of a container bridge named cbr0.

// HairpinMode denotes how the kubelet should configure networking to handle
// hairpin packets.
type HairpinMode string

// Enum settings for different ways to handle hairpin packets.
const (
	// Set the hairpin flag on the veth of containers in the respective
	// container runtime.
	HairpinVeth = "hairpin-veth"
	// Make the container bridge promiscuous. This will force it to accept
	// hairpin packets, even if the flag isn't set on ports of the bridge.
	PromiscuousBridge = "promiscuous-bridge"
	// Neither of the above. If the kubelet is started in this hairpin mode
	// and kube-proxy is running in iptables mode, hairpin packets will be
	// dropped by the container bridge.
	HairpinNone = "none"
)
