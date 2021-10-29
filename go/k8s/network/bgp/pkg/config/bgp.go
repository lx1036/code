package config

// struct for container bgp:config.
// Configuration parameters relating to the global BGP router.
type GlobalConfig struct {
	// original -> bgp:as
	// bgp:as's original type is inet:as-number.
	// Local autonomous system number of the router.  Uses
	// the 32-bit as-number type from the model in RFC 6991.
	As uint32 `mapstructure:"as" json:"as,omitempty"`
	// original -> bgp:router-id
	// bgp:router-id's original type is inet:ipv4-address.
	// Router id of the router, expressed as an
	// 32-bit value, IPv4 address.
	RouterId string `mapstructure:"router-id" json:"router-id,omitempty"`
	// original -> gobgp:port
	Port int32 `mapstructure:"port" json:"port,omitempty"`
	// original -> gobgp:local-address
	LocalAddressList []string `mapstructure:"local-address-list" json:"local-address-list,omitempty"`
}

// struct for container bgp:global.
// Global configuration for the BGP router.
type Global struct {
	// original -> bgp:global-config
	// Configuration parameters relating to the global BGP router.
	Config GlobalConfig `mapstructure:"config" json:"config,omitempty"`
}

// struct for container bgp:bgp.
// Top-level configuration and state for the BGP router.
type Bgp struct {
	// original -> bgp:global
	// Global configuration for the BGP router.
	Global Global `mapstructure:"global" json:"global,omitempty"`
}
