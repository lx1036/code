package meta

import (
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"path"

	"github.com/google/btree"
)

const (
	SnapshotDir     = "snapshot"
	snapshotDirTmp  = ".snapshot"
	snapshotBackup  = ".snapshot_backup"
	InodeFile       = "inode"
	DentryFile      = "dentry"
	applyIDFile     = "apply"
	SnapshotSign    = ".sign"
	MetadataFile    = "meta"
	metadataFileTmp = ".meta"
)

// INFO: storeMsg 包含两个 btree，这里如何持久化到磁盘是重点!!!
type storeMsg struct {
	command    uint32
	applyIndex uint64
	inodeTree  *btree.BTree
	dentryTree *btree.BTree
}

// INFO: 加载 data/metanode/partition_${id}/meta 文件
func (partition *metaPartition) loadMetadata() error {
	metaFile := path.Join(partition.config.RootDir, MetadataFile)
	content, err := ioutil.ReadFile(metaFile)
	if err != nil {
		return err
	}

	mConf := &MetaPartitionConfig{}
	if err = json.Unmarshal(content, mConf); err != nil {
		return err
	}

	partition.config = mConf
	partition.config.Cursor = mConf.Start
	return nil
}

// INFO: 加载 data/metanode/partition/partition_${id}/snapshot/inode 文件
func (partition *metaPartition) loadInode() (err error) {
	inodeFile := path.Join(partition.config.RootDir, SnapshotDir, InodeFile)
	content, err := ioutil.ReadFile(inodeFile)
	if err != nil {
		return err
	}

	inode := NewInode(0, 0)

}

// INFO: 这个函数很有参考意义，解决问题：如何持久化一个 BTree 到一个文件
func (partition *metaPartition) storeInode(msg *storeMsg) (uint32, error) {
	filename := path.Join(partition.config.RootDir, SnapshotDir, InodeFile)
	inodeFile, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = inodeFile.Sync() // 持久化到磁盘
		err = inodeFile.Close()
		if err != nil {
			klog.Error(err)
		}
	}()

	lenBuf := make([]byte, 4)
	sign := crc32.NewIEEE()
	msg.inodeTree.Ascend(func(i btree.Item) bool {
		inode := i.(*Inode)
		data, err := inode.MarshalToJSON() // 这里使用 json 序列化形式
		//data, err := inode.Marshal()
		if err != nil {
			return false
		}
		// INFO: 对于每一个Item，持久化数据格式 len + data(字节长度 + 数据)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
		if _, err := inodeFile.Write(lenBuf); err != nil {
			return false
		}
		if _, err := sign.Write(lenBuf); err != nil {
			return false
		}
		if _, err := inodeFile.Write(data); err != nil {
			return false
		}
		if _, err := sign.Write(data); err != nil {
			return false
		}

		return true
	})

	// 内容的hash值
	return sign.Sum32(), nil
}

// INFO: 加载 data/metanode/partition/partition_${id}/snapshot/dentry 文件
func (partition *metaPartition) loadDentry() error {

}

func (partition *metaPartition) storeDentry(msg *storeMsg) (uint32, error) {
	filename := path.Join(partition.config.RootDir, SnapshotDir, DentryFile)
	dentryFile, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = dentryFile.Sync() // 持久化到磁盘
		err = dentryFile.Close()
		if err != nil {
			klog.Error(err)
		}
	}()

	lenBuf := make([]byte, 4)
	sign := crc32.NewIEEE()
	msg.dentryTree.Ascend(func(i btree.Item) bool {
		inode := i.(*Dentry)
		data, err := inode.MarshalToJSON() // 这里使用 json 序列化形式
		//data, err := inode.Marshal()
		if err != nil {
			return false
		}
		// INFO: 对于每一个Item，持久化数据格式 len + data(字节长度 + 数据)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
		if _, err := dentryFile.Write(lenBuf); err != nil {
			return false
		}
		if _, err := sign.Write(lenBuf); err != nil {
			return false
		}
		if _, err := dentryFile.Write(data); err != nil {
			return false
		}
		if _, err := sign.Write(data); err != nil {
			return false
		}

		return true
	})

	// 内容的hash值
	return sign.Sum32(), nil
}
