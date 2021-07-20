package meta

import (
	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
	"syscall"
)

// INFO: `ll /mnt/hellofs`
func (mw *MetaWrapper) ReadDir_ll(parentID uint64) ([]proto.Dentry, error) {
	parentMetaPartition := mw.getPartitionByInode(parentID)
	if parentMetaPartition == nil {
		return nil, syscall.ENOENT
	}

	status, children, err := mw.readdir(parentMetaPartition, parentID)
	if err != nil || status != statusOK {
		return nil, statusToErrno(status)
	}
	return children, nil
}
