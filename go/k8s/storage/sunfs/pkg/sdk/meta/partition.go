package meta

import "github.com/google/btree"

type MetaPartition struct {
	PartitionID uint64
	Start       uint64
	End         uint64
	Members     []string
	LeaderAddr  string
	Status      int8
}

func (mw *MetaPartition) Less(than btree.Item) bool {
	that := than.(*MetaPartition)
	return mw.Start < that.Start
}

func (mw *MetaWrapper) getPartitionByInode(inode uint64) *MetaPartition {
	var metaPartition *MetaPartition
	mw.RLock()
	defer mw.RUnlock()

	pivot := &MetaPartition{Start: inode}
	mw.ranges.DescendLessOrEqual(pivot, func(item btree.Item) bool {
		metaPartition = item.(*MetaPartition)
		if inode > metaPartition.End || inode < metaPartition.Start {
			metaPartition = nil
		}
		// Iterate one item is enough
		return false
	})

	return metaPartition
}
