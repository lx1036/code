package meta

import "github.com/google/btree"

type MetaPartition struct {
	PartitionID uint64   `json:"PartitionID"`
	Start       uint64   `json:"Start"`
	End         uint64   `json:"End"`
	Members     []string `json:"Members"`
	LeaderAddr  string   `json:"LeaderAddr"`
	Status      int8     `json:"Status"`
}

func (mw *MetaPartition) Less(than btree.Item) bool {
	that := than.(*MetaPartition)
	return mw.Start < that.Start
}

// INFO: 根据 inodeID 获取 meta partition，这里有个设计点，partition 是根据 inodeID 范围划分的，
//  比如 range=1000, 则0-999 inodeID 是 partitionID 1；1000-1999 inodeID 是 partitionID 2
func (mw *MetaWrapper) getPartitionByInodeID(inodeID uint64) *MetaPartition {
	var metaPartition *MetaPartition
	mw.RLock()
	defer mw.RUnlock()

	pivot := &MetaPartition{Start: inodeID}
	mw.ranges.DescendLessOrEqual(pivot, func(item btree.Item) bool {
		metaPartition = item.(*MetaPartition)
		if inodeID > metaPartition.End || inodeID < metaPartition.Start {
			metaPartition = nil
		}
		// Iterate one item is enough
		return false
	})

	return metaPartition
}

func (mw *MetaWrapper) replaceOrInsertPartition(mp *MetaPartition) {
	mw.Lock()
	defer mw.Unlock()

	found, ok := mw.partitions[mp.PartitionID]
	if ok {
		mw.deletePartition(found)
	}

	mw.addPartition(mp)
	return
}

func (mw *MetaWrapper) deletePartition(mp *MetaPartition) {
	delete(mw.partitions, mp.PartitionID)
	mw.ranges.Delete(mp)
}

func (mw *MetaWrapper) addPartition(mp *MetaPartition) {
	mw.partitions[mp.PartitionID] = mp
	mw.ranges.ReplaceOrInsert(mp)
}

// INFO: 如果 rwPartitions 为空，则直接用 partitions
func (mw *MetaWrapper) getRWPartitions() []*MetaPartition {
	mw.Lock()
	defer mw.Unlock()

	rwPartitions := mw.rwPartitions
	if len(rwPartitions) == 0 {
		rwPartitions = make([]*MetaPartition, 0)
		for _, mp := range mw.partitions {
			rwPartitions = append(rwPartitions, mp)
		}
	}

	return rwPartitions
}
