package meta

import (
	"encoding/json"
	"sync"

	"github.com/google/btree"
)

// Dentry wraps necessary properties of the `dentry` information in file system.
// Marshal exporterKey:
//  +-------+----------+------+
//  | item  | ParentId | Name |
//  +-------+----------+------+
//  | bytes |    8     | rest |
//  +-------+----------+------+
// Marshal value:
//  +-------+-------+------+
//  | item  | Inode | Type |
//  +-------+-------+------+
//  | bytes |   8   |   4  |
//  +-------+-------+------+
// Marshal entity:
//  +-------+-----------+--------------+-----------+--------------+
//  | item  | KeyLength | MarshaledKey | ValLength | MarshaledVal |
//  +-------+-----------+--------------+-----------+--------------+
//  | bytes |     4     |   KeyLength  |     4     |   ValLength  |
//  +-------+-----------+--------------+-----------+--------------+
type Dentry struct {
	sync.RWMutex

	ParentId uint64 `json:"parent_id"` // FileID value of the parent inode.
	Name     string `json:"name"`      // Name of the current dentry.
	Inode    uint64 `json:"inode"`     // FileID value of the current inode.
	Type     uint32 `json:"type"`
}

func (dentry *Dentry) Less(than btree.Item) bool {
	d, ok := than.(*Dentry)
	return ok && ((dentry.ParentId < d.ParentId) || (dentry.ParentId == d.ParentId) && (dentry.Name < d.Name))
}

// MarshalToJSON is the wrapper of json.Marshal.
func (dentry *Dentry) MarshalToJSON() ([]byte, error) {
	dentry.RLock()
	defer dentry.RUnlock()

	return json.Marshal(dentry)
}
