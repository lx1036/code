package proto

// Peer defines the peer of the node id and address.
type Peer struct {
	ID   uint64 `json:"id"`
	Addr string `json:"addr"`
}
