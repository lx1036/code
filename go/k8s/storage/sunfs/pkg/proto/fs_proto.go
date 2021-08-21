package proto

import "os"

// INFO: Meta Partition

// UpdateMetaPartitionRequest defines the request to update a meta partition.
type UpdateMetaPartitionRequest struct {
	PartitionID uint64
	VolName     string
	Start       uint64
	End         uint64
}

// UpdateMetaPartitionResponse defines the response to the request of updating the meta partition.
type UpdateMetaPartitionResponse struct {
	PartitionID uint64
	VolName     string
	End         uint64
	Status      uint8
	Result      string
}

// INFO: Inode

const (
	RootInode = uint64(1)
)

// CreateInodeRequest defines the request to create an inode.
type CreateInodeRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	Mode        uint32 `json:"mode"`
	Uid         uint32 `json:"uid"`
	Gid         uint32 `json:"gid"`
	Target      []byte `json:"tgt"`
	PInode      uint64 `json:"pino"`
}

// CreateInodeResponse defines the response to the request of creating an inode.
type CreateInodeResponse struct {
	Info *InodeInfo `json:"info"`
}

// InodeGetRequest defines the request to get the inode.
type InodeGetRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	Inode       uint64 `json:"ino"`
}

// InodeGetResponse defines the response to the InodeGetRequest.
type InodeGetResponse struct {
	Info *InodeInfo `json:"info"`
}

// UnlinkInodeRequest defines the request to unlink an inode.
type UnlinkInodeRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	Inode       uint64 `json:"ino"`
}

// BatchInodeGetRequest defines the request to get the inode in batch.
type BatchInodeGetRequest struct {
	VolName     string   `json:"vol"`
	PartitionID uint64   `json:"pid"`
	Inodes      []uint64 `json:"inos"`
}

// LinkInodeRequest defines the request to link an inode.
type LinkInodeRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	Inode       uint64 `json:"ino"`
}

// EvictInodeRequest defines the request to evict an inode.
type EvictInodeRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	Inode       uint64 `json:"ino"`
}

type InodeInfo struct {
	Inode      uint64 `json:"ino"`
	Mode       uint32 `json:"mode"`
	Nlink      uint32 `json:"nlink"`
	Size       uint64 `json:"sz"`
	Uid        uint32 `json:"uid"`
	Gid        uint32 `json:"gid"`
	Generation uint64 `json:"gen"`
	ModifyTime int64  `json:"mt"`
	CreateTime int64  `json:"ct"`
	AccessTime int64  `json:"at"`
	Target     []byte `json:"tgt"`
	PInode     uint64 `json:"pino"`
}

// OsMode returns os.FileMode.
func OsFileMode(mode uint32) os.FileMode {
	return os.FileMode(mode)
}

func IsDir(mode uint32) bool {
	return OsFileMode(mode).IsDir()
}

// INFO: Dentry

// Dentry defines the dentry struct.
type Dentry struct {
	Name  string `json:"name"`
	Inode uint64 `json:"ino"`
	Type  uint32 `json:"type"`
}

// CreateDentryRequest defines the request to create a dentry.
type CreateDentryRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	ParentID    uint64 `json:"pino"`
	Inode       uint64 `json:"ino"`
	Name        string `json:"name"`
	Mode        uint32 `json:"mode"`
}

// UpdateDentryRequest defines the request to update a dentry.
type UpdateDentryRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	ParentID    uint64 `json:"pino"`
	Name        string `json:"name"`
	Inode       uint64 `json:"ino"` // new inode number
}

// DeleteDentryRequest define the request tp delete a dentry.
type DeleteDentryRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	ParentID    uint64 `json:"pino"`
	Name        string `json:"name"`
}

// ReadDirRequest defines the request to read dir.
type ReadDirRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	ParentID    uint64 `json:"pino"`
}

// ReadDirResponse defines the response to the request of reading dir.
type ReadDirResponse struct {
	Children []Dentry `json:"children"`
}

// LookupRequest defines the request for lookup.
type LookupRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	ParentID    uint64 `json:"pino"`
	Name        string `json:"name"`
}

// LookupNameRequest defines the request for lookupName
type LookupNameRequest struct {
	VolName     string `json:"vol"`
	PartitionID uint64 `json:"pid"`
	ParentID    uint64 `json:"pino"`
	ID          uint64 `json:"id"`
}
