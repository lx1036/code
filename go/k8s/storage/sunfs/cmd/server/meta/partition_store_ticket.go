package meta

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

//
func (partition *metaPartition) store(msg *storeMsg) error {

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
