package endpoint

// Endpoint represents a container or similar which can be individually
// addresses on L3 with its own IP addresses.
//
// The representation of the Endpoint which is serialized to disk for restore
// purposes is the serializableEndpoint type in this package.
type Endpoint struct {
	isHost bool
}

func (e *Endpoint) IsHost() bool {
	return e.isHost
}
