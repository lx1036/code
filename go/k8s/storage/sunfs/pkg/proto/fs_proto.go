package proto

// Dentry defines the dentry struct.
type Dentry struct {
	Name  string `json:"name"`
	Inode uint64 `json:"ino"`
	Type  uint32 `json:"type"`
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
