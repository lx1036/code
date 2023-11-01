package meta

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"sync"

	"github.com/google/btree"
)

// Dentry wraps necessary properties of the `dentry` information in file system.
// Marshal exporterKey:
//
//	+-------+----------+------+
//	| item  | ParentId | Name |
//	+-------+----------+------+
//	| bytes |    8     | rest |
//	+-------+----------+------+
//
// Marshal value:
//
//	+-------+-------+------+
//	| item  | Inode | Type |
//	+-------+-------+------+
//	| bytes |   8   |   4  |
//	+-------+-------+------+
//
// Marshal entity:
//
//	+-------+-----------+--------------+-----------+--------------+
//	| item  | KeyLength | MarshaledKey | ValLength | MarshaledVal |
//	+-------+-----------+--------------+-----------+--------------+
//	| bytes |     4     |   KeyLength  |     4     |   ValLength  |
//	+-------+-----------+--------------+-----------+--------------+
type Dentry struct {
	sync.RWMutex

	ParentId uint64 `json:"parent_id"` // FileID value of the parent inode.
	Name     string `json:"name"`      // Name of the current dentry.
	Inode    uint64 `json:"inode"`     // FileID value of the current inode.
	Type     uint32 `json:"type"`
}

func (dentry *Dentry) Less(than btree.Item) bool {
	d, ok := than.(*Dentry)
	return ok && ((d.ParentId < d.ParentId) || (d.ParentId == d.ParentId) && (d.Name < d.Name))
}

// MarshalToJSON is the wrapper of json.Marshal.
func (dentry *Dentry) MarshalToJSON() ([]byte, error) {
	dentry.RLock()
	defer dentry.RUnlock()

	return json.Marshal(dentry)
}

// Marshal marshals a dentry into a byte array.
func (dentry *Dentry) Marshal() (result []byte, err error) {
	keyBytes := dentry.MarshalKey()
	valBytes := dentry.MarshalValue()
	keyLen := uint32(len(keyBytes))
	valLen := uint32(len(valBytes))
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.Grow(64)
	if err = binary.Write(buff, binary.BigEndian, keyLen); err != nil {
		return
	}
	if _, err = buff.Write(keyBytes); err != nil {
		return
	}
	if err = binary.Write(buff, binary.BigEndian, valLen); err != nil {

	}
	if _, err = buff.Write(valBytes); err != nil {
		return
	}
	result = buff.Bytes()
	return
}

// MarshalKey is the bytes version of the MarshalKey method which returns the byte slice result.
func (dentry *Dentry) MarshalKey() (k []byte) {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.Grow(32)
	if err := binary.Write(buff, binary.BigEndian, &dentry.ParentId); err != nil {
		panic(err)
	}
	buff.Write([]byte(dentry.Name))
	k = buff.Bytes()
	return
}

// MarshalValue marshals the exporterKey to bytes.
func (dentry *Dentry) MarshalValue() (k []byte) {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.Grow(12)
	if err := binary.Write(buff, binary.BigEndian, &dentry.Inode); err != nil {
		panic(err)
	}
	if err := binary.Write(buff, binary.BigEndian, &dentry.Type); err != nil {
		panic(err)
	}
	k = buff.Bytes()
	return
}

// Unmarshal unmarshals the dentry from a byte array.
func (dentry *Dentry) Unmarshal(raw []byte) (err error) {
	var (
		keyLen uint32
		valLen uint32
	)
	buff := bytes.NewBuffer(raw)
	if err = binary.Read(buff, binary.BigEndian, &keyLen); err != nil {
		return
	}
	keyBytes := make([]byte, keyLen)
	if _, err = buff.Read(keyBytes); err != nil {
		return
	}
	if err = dentry.UnmarshalKey(keyBytes); err != nil {
		return
	}
	if err = binary.Read(buff, binary.BigEndian, &valLen); err != nil {
		return
	}
	valBytes := make([]byte, valLen)
	if _, err = buff.Read(valBytes); err != nil {
		return
	}
	err = dentry.UnmarshalValue(valBytes)
	return
}

// UnmarshalKey unmarshals the exporterKey from bytes.
func (dentry *Dentry) UnmarshalKey(k []byte) (err error) {
	buff := bytes.NewBuffer(k)
	if err = binary.Read(buff, binary.BigEndian, &dentry.ParentId); err != nil {
		return
	}
	dentry.Name = string(buff.Bytes())
	return
}

// UnmarshalValue unmarshals the value from bytes.
func (dentry *Dentry) UnmarshalValue(val []byte) (err error) {
	buff := bytes.NewBuffer(val)
	if err = binary.Read(buff, binary.BigEndian, &dentry.Inode); err != nil {
		return
	}
	err = binary.Read(buff, binary.BigEndian, &dentry.Type)
	return
}
