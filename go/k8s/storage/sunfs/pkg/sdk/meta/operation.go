package meta

import (
	"fmt"
	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
	"syscall"

	"k8s.io/klog/v2"
)

// INFO: VFS(虚拟文件系统) = 超级块super_block + inode + dentry + vfsmount
//  inode: 记录文件系统对象中的一般元数据，比如 kind/size/created/modified/sharing&permissions
//  dentry: 记录文件系统对象在全局文件系统树中的位置
//  **[Linux 虚拟文件系统四大对象：超级块、inode、dentry、file之间关系](https://www.eet-china.com/mp/a38145.html)**

func (mw *MetaWrapper) readdir(mp *MetaPartition, parentID uint64) (status int,
	children []proto.Dentry, err error) {

	req := &proto.ReadDirRequest{
		VolName:     mw.volname,
		PartitionID: mp.PartitionID,
		ParentID:    parentID,
	}

	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaReadDir
	err = packet.MarshalData(req)
	if err != nil {
		klog.Errorf("readdir: req(%v) err(%v)", *req, err)
		return
	}

	//
	packet, err = mw.sendToMetaPartition(mp, packet)
	if err != nil {
		klog.Errorf("readdir: packet(%v) mp(%v) req(%v) err(%v)", packet, mp, *req, err)
		return
	}

	status = parseStatus(packet.ResultCode)
	if status != statusOK {
		children = make([]proto.Dentry, 0)
		klog.Errorf("readdir: packet(%v) mp(%v) req(%v) result(%v)", packet, mp, *req, packet.GetResultMsg())
		return
	}

	resp := new(proto.ReadDirResponse)
	err = packet.UnmarshalData(resp)
	if err != nil {
		klog.Errorf("readdir: packet(%v) mp(%v) err(%v) PacketData(%v)", packet, mp, err, string(packet.Data))
		return
	}

	klog.Infof("readdir: packet(%v) mp(%v) req(%v)", packet, mp, *req)
	return statusOK, resp.Children, nil
}

// INFO: 在partition里新建inode
func (mw *MetaWrapper) inodeCreate(rwPartition *MetaPartition, mode, uid, gid uint32, target []byte,
	pino uint64) (status int, info *proto.InodeInfo, err error) {
	req := &proto.CreateInodeRequest{
		VolName:     mw.volname,
		PartitionID: rwPartition.PartitionID,
		Mode:        mode,
		Uid:         uid,
		Gid:         gid,
		Target:      target,
		PInode:      pino,
	}
	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaCreateInode // Create Inode
	err = packet.MarshalData(req)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[inodeCreate]packet MarshalData err(%v)", err))
		return
	}

	response, err := mw.sendToMetaPartition(rwPartition, packet)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[inodeCreate]sendToMetaPartition: err(%v)", err))
		return
	}

	status = parseStatus(response.ResultCode)
	if status != statusOK {
		klog.Errorf(fmt.Sprintf("[inodeCreate]parseStatus no ok"))
		return
	}

	resp := new(proto.CreateInodeResponse)
	err = packet.UnmarshalData(resp)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[inodeCreate]packet UnmarshalData err(%v)", err))
		return
	}
	if resp.Info == nil {
		err = fmt.Errorf("[inodeCreate]CreateInodeResponse info is nil")
		klog.Error(err)
		return
	}

	return statusOK, resp.Info, nil
}

// INFO: 在 parent partition 里新建 dentry
func (mw *MetaWrapper) dentryCreate(parentMetaPartition *MetaPartition, parentInodeID uint64, name string,
	inode uint64, mode uint32) (int, error) {
	if parentInodeID == inode {
		return statusExist, nil
	}

	req := &proto.CreateDentryRequest{
		VolName:     mw.volname,
		PartitionID: parentMetaPartition.PartitionID,
		ParentID:    parentInodeID,
		Inode:       inode,
		Name:        name,
		Mode:        mode,
	}
	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaCreateDentry // Create Dentry
	err := packet.MarshalData(req)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[dentryCreate]CreateDentryRequest MarshalData err %v", err))
		return 0, err
	}

	response, err := mw.sendToMetaPartition(parentMetaPartition, packet)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[inodeCreate]sendToMetaPartition: err(%v)", err))
		return 0, err
	}

	status := parseStatus(response.ResultCode)
	if status != statusOK {
		msg := fmt.Sprintf("[inodeCreate]parseStatus no ok")
		klog.Errorf(msg)
		return 0, fmt.Errorf(msg)
	}

	return status, nil
}

// Proto ResultCode to status
func parseStatus(result uint8) (status int) {
	switch result {
	case proto.OpOk:
		status = statusOK
	case proto.OpExistErr:
		status = statusExist
	case proto.OpNotExistErr:
		status = statusNoent
	case proto.OpInodeFullErr:
		status = statusFull
	case proto.OpAgain:
		status = statusAgain
	case proto.OpArgMismatchErr:
		status = statusInval
	case proto.OpNotPerm:
		status = statusNotPerm
	default:
		status = statusError
	}
	return
}

func statusToErrno(status int) error {
	switch status {
	case statusOK:
		// return error anyway
		return syscall.EAGAIN
	case statusExist:
		return syscall.EEXIST
	case statusNoent:
		return syscall.ENOENT
	case statusFull:
		return syscall.ENOMEM
	case statusAgain:
		return syscall.EAGAIN
	case statusInval:
		return syscall.EINVAL
	case statusNotPerm:
		return syscall.EPERM
	case statusError:
		return syscall.EPERM
	default:
	}
	return syscall.EIO
}
