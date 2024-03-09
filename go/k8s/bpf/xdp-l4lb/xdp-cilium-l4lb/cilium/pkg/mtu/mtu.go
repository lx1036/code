package mtu

// Configuration is an MTU configuration as returned by NewConfiguration
type Configuration struct {
	// standardMTU is the regular MTU used for configuring devices and
	// routes where packets are expected to be delivered outside the node.
	//
	// Note that this is a singleton for the process including this
	// package. This means, for instance, that when using this from the
	// ``pkg/plugins/*`` sources, it will not respect the settings
	// configured inside the ``daemon/``.
	standardMTU int

	// tunnelMTU is the MTU used for configuring a tunnel mesh for
	// inter-node connectivity.
	//
	// Similar to StandardMTU, this is a singleton for the process.
	tunnelMTU int

	// preEncrypMTU is the MTU used for configurations of a encryption route.
	// If tunneling is enabled the tunnelMTU is used which will include
	// additional encryption overhead if needed.
	preEncryptMTU int

	// postEncryptMTU is the MTU used for configurations of a encryption
	// route _after_ encryption tags have been addded. These will be used
	// in the encryption routing table. The MTU accounts for the tunnel
	// overhead, if any, but assumes packets are already encrypted.
	postEncryptMTU int

	encapEnabled     bool
	encryptEnabled   bool
	wireguardEnabled bool
}
