package proto

import "fmt"

// Peer defines the peer of the node id and address.
type Peer struct {
	ID   uint64 `json:"id"`
	Addr string `json:"addr"`
}

// CreateMetaPartitionRequest defines the request to create a meta partition.
type CreateMetaPartitionRequest struct {
	MetaId      string
	VolName     string
	Start       uint64
	End         uint64
	PartitionID uint64
	Members     []Peer
}

func (cr *CreateMetaPartitionRequest) ToString() string {
	return fmt.Sprintf("MetaId[%v] VolName[%v] Start[%v] End[%v] Members:%+v", cr.MetaId, cr.VolName, cr.Start, cr.End, cr.Members)
}

// CreateMetaPartitionResponse defines the response to the request of creating a meta partition.
type CreateMetaPartitionResponse struct {
	VolName     string
	PartitionID uint64
	Status      uint8
	Result      string
}
