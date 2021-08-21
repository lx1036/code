package meta

import (
	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
	"sync/atomic"
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

func (mw *MetaWrapper) Create_ll(parentInodeID uint64, name string, mode, uid, gid uint32,
	target []byte) (*proto.InodeInfo, error) {

	parentMetaPartition := mw.getPartitionByInode(parentInodeID)
	if parentMetaPartition == nil {

		return nil, syscall.ENOENT
	}

	rwPartitions := mw.getRWPartitions()
	length := len(rwPartitions)
	epoch := atomic.AddUint64(&mw.epoch, 1)
	found := false
	var info *proto.InodeInfo
	var err error
	var status int
	var rwPartition *MetaPartition
	for i := 0; i < length; i++ {
		index := (int(epoch) + i) % length
		rwPartition := rwPartitions[index]
		status, info, err = mw.inodeCreate(rwPartition, mode, uid, gid, target, parentInodeID)
		if err == nil && status == statusOK {
			found = true
			break
		}
	}

	if !found {
		return nil, syscall.ENOMEM
	}

	// create dentry
	status, err = mw.dentryCreate(parentMetaPartition, parentInodeID, name, info.Inode, mode)
	if err != nil || status != statusOK {
		if status == statusExist {
			return nil, syscall.EEXIST
		} else {
			mw.iunlink(rwPartition, info.Inode)
			mw.ievict(rwPartition, info.Inode)
			return nil, statusToErrno(status)
		}
	}

	return info, nil
}
