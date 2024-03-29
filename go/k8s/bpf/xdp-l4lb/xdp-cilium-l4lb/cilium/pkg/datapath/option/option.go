package option

// Available options for datapath mode.
const (
	// DatapathModeVeth specifies veth datapath mode (i.e. containers are
	// attached to a network via veth pairs).
	DatapathModeVeth = "veth"

	// DatapathModeIpvlan specifies ipvlan datapath mode.
	DatapathModeIpvlan = "ipvlan"

	// DatapathModeLBOnly specifies lb-only datapath mode.
	DatapathModeLBOnly = "lb-only"
)
