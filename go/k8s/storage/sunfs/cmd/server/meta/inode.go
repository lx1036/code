package meta

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/google/btree"
)

// TODO: 为了debug，后面加上 json 序列化格式

// Inode wraps necessary properties of `Inode` information in the file system.
// Marshal exporterKey:
//  +-------+-------+
//  | item  | Inode |
//  +-------+-------+
//  | bytes |   8   |
//  +-------+-------+
// Marshal value:
//  +-------+------+------+-----+----+----+----+
//  | item  | Type | Size | Gen | CT | AT | MT |
//  +-------+------+------+-----+----+----+----+
//  | bytes |  4   |  8   |  8  | 8  | 8  | 8  |
//  +-------+------+------+-----+----+----+----+
// Marshal entity:
//  +-------+-----------+--------------+-----------+--------------+
//  | item  | KeyLength | MarshaledKey | ValLength | MarshaledVal |
//  +-------+-----------+--------------+-----------+--------------+
//  | bytes |     4     |   KeyLength  |     4     |   ValLength  |
//  +-------+-----------+--------------+-----------+--------------+
type Inode struct {
	sync.RWMutex
	Inode uint64 `json:"inode"` // Inode ID // 8

	// INFO: Marshal Value
	Type       uint32 `json:"type"`        // 4
	Uid        uint32 `json:"uid"`         // 4
	Gid        uint32 `json:"gid"`         // 4
	Size       uint64 `json:"size"`        // 8
	Generation uint64 `json:"generation"`  // 8
	CreateTime int64  `json:"create_time"` // 16
	AccessTime int64  `json:"access_time"` // 16
	ModifyTime int64  `json:"modify_time"` // 16
	LinkTarget []byte `json:"link_target"` // SymLink target name
	NLink      uint32 `json:"n_link"`      // NodeLink counts // 4
	Flag       int32  `json:"flag"`        // 8
	PInode     uint64 `json:"p_inode"`     // 8
	Reserved   uint64 `json:"reserved"`    // reserved space // 8
}

func (i *Inode) Less(than btree.Item) bool {
	inode, ok := than.(*Inode)
	return ok && i.Inode < inode.Inode
}

// MarshalToJSON is the wrapper of json.Marshal.
func (i *Inode) MarshalToJSON() ([]byte, error) {
	i.RLock()
	defer i.RUnlock()
	return json.Marshal(i)
}

// Marshal INFO: 序列化格式还是：字节长度 + 数据
func (i *Inode) Marshal() ([]byte, error) {
	keyBytes := i.MarshalKey()
	valBytes := i.MarshalValue()
	keyLen := uint32(len(keyBytes))
	valLen := uint32(len(valBytes))

	buff := bytes.NewBuffer(make([]byte, 0, 128))
	buff.Grow(128)
	var err error
	if err = binary.Write(buff, binary.BigEndian, keyLen); err != nil {
		return nil, err
	}
	if _, err = buff.Write(keyBytes); err != nil {
		return nil, err
	}
	if err = binary.Write(buff, binary.BigEndian, valLen); err != nil {
		return nil, err
	}
	if _, err = buff.Write(valBytes); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (i *Inode) MarshalKey() []byte {
	buffer := bytes.NewBuffer(make([]byte, 0, 8)) // uint64 是 8 字节长度
	if err := binary.Write(buffer, binary.BigEndian, &i.Inode); err != nil {
		return nil
	}

	return buffer.Bytes()
}

func (i *Inode) MarshalValue() []byte {
	buffer := bytes.NewBuffer(make([]byte, 0, 128))
	buffer.Grow(64)
	var err error

	// INFO: 加读锁
	i.RLock()
	// INFO: 为何这里都是指针，这里Write()函数里已经做了处理，指针和形参都一样。
	//  以及为啥这里要使用 binary.Write 这种方式保存???
	if err = binary.Write(buffer, binary.BigEndian, &i.Type); err != nil {
		return nil
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.Uid); err != nil {
		return nil
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.Gid); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.Size); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.Generation); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.CreateTime); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.AccessTime); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.ModifyTime); err != nil {
		panic(err)
	}
	// write SymLink
	symSize := uint32(len(i.LinkTarget))
	if err = binary.Write(buffer, binary.BigEndian, &symSize); err != nil {
		panic(err)
	}
	if _, err = buffer.Write(i.LinkTarget); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.NLink); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.Flag); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.PInode); err != nil {
		panic(err)
	}
	if err = binary.Write(buffer, binary.BigEndian, &i.Reserved); err != nil {
		panic(err)
	}

	i.RUnlock()

	return buffer.Bytes()
}

// NewInode returns a new Inode instance with specified Inode ID, name and type.
// The AccessTime and ModifyTime will be set to the current time.
func NewInode(inode uint64, t uint32) *Inode {
	ts := time.Now().Unix()
	i := &Inode{
		Inode:      inode,
		Type:       t,
		Generation: 1,
		CreateTime: ts,
		AccessTime: ts,
		ModifyTime: ts,
		NLink:      1,
	}

	if os.FileMode(t).IsDir() {
		i.NLink = 2
	}

	return i
}
