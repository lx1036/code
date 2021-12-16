package partition

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/btree"
)

const (
	SnapshotDir     = "snapshot"
	snapshotDirTmp  = ".snapshot"
	snapshotBackup  = ".snapshot_backup"
	InodeFile       = "inode"
	DentryFile      = "dentry"
	ApplyIDFile     = "apply"
	SnapshotSign    = "sign"
	MetadataFile    = "meta"
	metadataFileTmp = ".meta"
)

const (
	StateStandby uint32 = iota
	StateStart
	StateRunning
	StateShutdown
	StateStopped
)

const (
	intervalToPersistData = time.Minute * 5
)

// INFO: storeMsg 包含两个 btree，这里如何持久化到磁盘是重点!!!
type storeMsg struct {
	command    uint32
	applyIndex uint64
	inodeTree  *BTree
	dentryTree *BTree
}

// INFO: 开启/关闭 定时 snapshot，且只有 leader 才会 snapshot，如果 leader change，需要关闭定时 snapshot
func (partition *MetaPartitionFSM) startSchedule(curIndex uint64) {
	scheduleState := StateStopped
	timer := time.NewTimer(time.Hour * 24 * 365)
	timer.Stop()

	storeMsgFunc := func(msg *storeMsg) {
		if err := partition.store(msg); err == nil {
			// INFO: 已经持久化了，可以 truncate raft log
			if partition.raftPartition != nil {
				partition.raftPartition.Truncate(curIndex)
			}
			curIndex = msg.applyIndex
		} else {
			// INFO: retry store again
			partition.storeChan <- msg
		}

		if _, ok := partition.IsLeader(); ok {
			timer.Reset(intervalToPersistData)
		}
		scheduleState = StateStopped
	}

	go func(stopC chan bool) {
		var messages []*storeMsg
		readyChan := make(chan struct{}, 1)
		for {
			if len(messages) > 0 {
				if scheduleState == StateStopped {
					scheduleState = StateRunning
					readyChan <- struct{}{}
				}
			}

			select {
			case <-stopC:
				return
			case <-timer.C:
				if partition.applyID <= curIndex { // 开启定时 snapshot
					timer.Reset(intervalToPersistData)
					continue
				}
				partition.Put(opFSMStoreTick, nil)
			case <-readyChan:
				// INFO: 第一个 msg 会走 storeMsgFunc，然后后面的 msg 存储在 []messages，如果没有新的，
				//  则 intervalToPersistData 之后，会继续下一个循环走 storeMsgFunc，这时上一个 storeMsgFunc 已经走完了，
				//  这样可以避免并发走 store(msg *storeMsg)
				msg := findMaxApplyIndexMsg(messages, curIndex)
				if msg != nil {
					go storeMsgFunc(msg)
				}
				messages = messages[:0]
			case msg := <-partition.storeChan: // INFO: storeMsg channel buffer size = 5
				switch msg.command {
				case startStoreTick: // 开启定时 snapshot
					timer.Reset(intervalToPersistData)
				case stopStoreTick: // 关闭定时 snapshot
					timer.Stop()
				case opFSMStoreTick:
					messages = append(messages, msg)
				}
			}
		}
	}(partition.stopC)
}

// INFO: 找出比当前 index 大的且最大的那个 storeMsg
func findMaxApplyIndexMsg(messages []*storeMsg, index uint64) *storeMsg {
	var maxMessage *storeMsg
	maxApplyIndex := uint64(0)
	for _, message := range messages {
		if message.applyIndex <= index {
			continue
		}

		if message.applyIndex > maxApplyIndex {
			maxApplyIndex = message.applyIndex
			maxMessage = message
		}
	}

	return maxMessage
}

// INFO: save inode/dentry/apply/sign snapshot file
func (partition *MetaPartitionFSM) store(msg *storeMsg) error {
	// INFO: save msg into data/metanode/partition/partition_${id}/snapshot/inode
	inodeCRC, err := partition.storeInode(msg)
	if err != nil {
		return err
	}
	// INFO: save msg into data/metanode/partition/partition_${id}/snapshot/dentry
	dentryCRC, err := partition.storeDentry(msg)
	if err != nil {
		return err
	}
	// INFO: save msg into data/metanode/partition/partition_${id}/snapshot/apply
	err = partition.storeApplyID(msg)
	if err != nil {
		return err
	}
	// INFO: save crc into data/metanode/partition/partition_${id}/snapshot/sign
	if err = ioutil.WriteFile(path.Join(partition.config.RootDir, SnapshotDir, SnapshotSign),
		[]byte(fmt.Sprintf("%d %d", inodeCRC, dentryCRC)), 0775); err != nil {
		return err
	}

	return nil
}

// INFO: 加载 data/metanode/partition_${id}/meta 文件
func (partition *MetaPartitionFSM) loadMetadata() error {
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
func (partition *MetaPartitionFSM) loadInode() (err error) {
	inodeFile := path.Join(partition.config.RootDir, SnapshotDir, InodeFile)
	content, err := ioutil.ReadFile(inodeFile)
	if err != nil {
		return err
	}

	inode := NewInode(0, 0)

}

// INFO: 这个函数很有参考意义，解决问题：如何持久化一个 BTree 到一个文件
func (partition *MetaPartitionFSM) storeInode(msg *storeMsg) (uint32, error) {
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
func (partition *MetaPartitionFSM) loadDentry() error {

}

// INFO: 迭代BTree，序列化每一个 btree.Item，然后持久化到一个文件内
func (partition *MetaPartitionFSM) storeDentry(msg *storeMsg) (uint32, error) {
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

func (partition *MetaPartitionFSM) storeApplyID(msg *storeMsg) error {
	filename := path.Join(partition.config.RootDir, SnapshotDir, ApplyIDFile)
	// INFO: 注意这里使用的是 atomic.LoadUint64()，非常重要，学习下 atomic ！！！
	return ioutil.WriteFile(filename, []byte(fmt.Sprintf("%d|%d", msg.applyIndex, atomic.LoadUint64(&partition.config.Cursor))), 0775)
}

func (partition *MetaPartitionFSM) loadApplyID() error {
	filename := path.Join(partition.config.RootDir, SnapshotDir, ApplyIDFile)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if len(content) == 0 {
		return fmt.Errorf("[loadApplyID]applyID is empty")
	}

	var cursor uint64
	if strings.Contains(string(content), "|") {
		_, err = fmt.Sscanf(string(content), "%d|%d", &partition.applyID, &cursor)
	} else {
		_, err = fmt.Sscanf(string(content), "%d", &partition.applyID)
	}
	if err != nil {
		return err
	}

	if cursor > atomic.LoadUint64(&partition.config.Cursor) {
		atomic.StoreUint64(&partition.config.Cursor, cursor)
	}

	return nil
}
