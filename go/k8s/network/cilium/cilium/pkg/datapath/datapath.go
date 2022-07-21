package datapath

type Datapath interface {
	ConfigWriter

	// LocalNodeAddressing must return the node addressing implementation
	// of the local node
	LocalNodeAddressing() NodeAddressing

	// Node must return the handler for node events
	Node() NodeHandler

	// Loader must return the implementation of the loader, which is responsible
	// for loading, reloading, and compiling datapath programs.
	Loader() Loader
}
