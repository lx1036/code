package meta

import (
	"fmt"
	"sync/atomic"
	"syscall"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

// INFO: `ll /mnt/hellofs`
func (mw *MetaWrapper) ReadDir_ll(parentID uint64) ([]proto.Dentry, error) {
	parentMetaPartition := mw.getPartitionByInodeID(parentID)
	if parentMetaPartition == nil {
		return nil, syscall.ENOENT
	}

	status, children, err := mw.readdir(parentMetaPartition, parentID)
	if err != nil || status != statusOK {
		return nil, statusToErrno(status)
	}
	return children, nil
}

// Create_ll INFO: 在meta cluster 的 partition 中，创建 inode/dentry 对象，其实这个函数重要一个逻辑是：分配一个 inodeID
func (mw *MetaWrapper) Create_ll(parentInodeID uint64, name string, mode, uid, gid uint32,
	target []byte) (*proto.InodeInfo, error) {
	parentMetaPartition := mw.getPartitionByInodeID(parentInodeID)
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
		rwPartition = rwPartitions[index]
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
			mw.inodeUnlink(rwPartition, info.Inode)
			mw.inodeEvict(rwPartition, info.Inode)
			return nil, statusToErrno(status)
		}
	}

	return info, nil
}

func (mw *MetaWrapper) InodeGet_ll(inodeID uint64) (*proto.InodeInfo, error) {
	// 本地记录了 inodeID 和 partition 对应信息
	partition := mw.getPartitionByInodeID(inodeID)
	if partition == nil {
		klog.Errorf(fmt.Sprintf("[InodeGet_ll]no partition for inodeID %d", inodeID))
		return nil, syscall.ENOENT
	}

	status, info, err := mw.inodeGet(partition, inodeID)
	if err != nil || status != statusOK {
		return nil, statusToErrno(status)
	}

	return info, nil
}
