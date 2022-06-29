package datapath

type Datapath interface {
	// LocalNodeAddressing must return the node addressing implementation
	// of the local node
	LocalNodeAddressing() NodeAddressing
}
