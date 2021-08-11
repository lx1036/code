package meta

import (
	"fmt"
	"io/ioutil"
	"path"
)

// INFO: 开启 loop 定时存储 msg
func (partition *metaPartition) startSchedule(curIndex uint64) {
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
	}

	go func(stopC chan bool) {
		var messages []*storeMsg
		readyChan := make(chan struct{}, 1)
		for {
			if len(messages) > 0 {
				readyChan <- struct{}{}
			}

			select {
			case <-stopC:

				return
			case <-readyChan:
				msg := findMaxApplyIndexMsg(messages, curIndex)
				if msg != nil {
					go storeMsgFunc(msg)
				}
				messages = messages[:0]
			case msg := <-partition.storeChan: // INFO: storeMsg channel buffer size = 5
				switch msg.command {
				case opFSMStoreTick:
					messages = append(messages, msg)
				}
			}
		}
	}(partition.stopC)
}

// INFO: save inode/dentry/apply/sign snapshot file
func (partition *metaPartition) store(msg *storeMsg) error {
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
