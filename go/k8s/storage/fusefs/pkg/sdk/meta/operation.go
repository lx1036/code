package meta

import (
	"fmt"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"syscall"

	"k8s.io/klog/v2"
)

// INFO: VFS(虚拟文件系统) = 超级块super_block + inode + dentry + vfsmount
//  inode 索引节点: 记录文件系统对象中的一般元数据，比如 kind/size/created/modified/sharing&permissions/指向存储该内容的磁盘区块的指针/文件分类
//  dentry 目录项: 记录文件系统对象在全局文件系统树中的位置，例如：open一个文件/home/xxx/yyy.txt，那么/、home、xxx、yyy.txt都是一个目录项，
//  VFS在查找的时候，根据一层一层的目录项找到对应的每个目录项的inode，那么沿着目录项进行操作就可以找到最终的文件。
//  **[Linux 虚拟文件系统四大对象：超级块、inode、dentry、file之间关系](https://www.eet-china.com/mp/a38145.html)**

// INFO: VFS 就是一层接口，定义了一些接口函数
//  VFS是一种软件机制，只存在于内存中，每次系统初始化期间Linux都会先在内存中构造一棵VFS的目录树（也就是源码中的namespace）。
//  VFS主要的作用是对上层应用屏蔽底层不同的调用方法，提供一套统一的调用接口，二是便于对不同的文件系统进行组织管理。
//  VFS提供了一个抽象层，将POSIX API接口与不同存储设备的具体接口实现进行了分离，使得底层的文件系统类型、设备类型对上层应用程序透明。
//  例如read，write，那么映射到VFS中就是sys_read，sys_write，那么VFS可以根据你操作的是哪个“实际文件系统”(哪个分区)来进行不同的实际的操作！这个技术也是很熟悉的“钩子结构”技术来处理的。
//  其实就是VFS中提供一个抽象的struct结构体，然后对于每一个具体的文件系统要把自己的字段和函数填充进去，这样就解决了异构问题（内核很多子系统都大量使用了这种机制）

func (mw *MetaWrapper) readdir(mp *MetaPartition, parentID uint64) (status int,
	children []proto.Dentry, err error) {

	req := &proto.ReadDirRequest{
		VolName:     mw.volname,
		PartitionID: mp.PartitionID,
		ParentID:    parentID,
	}

	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaReadDir // Read Dir
	err = packet.MarshalData(req)
	if err != nil {
		klog.Errorf("readdir: req(%v) err(%v)", *req, err)
		return
	}

	// INFO: 调用 meta cluster Read Dir api
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

// INFO: api调用meta cluster，在partition里新建inode
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

	// INFO: 调用 meta cluster Create Inode api
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

func (mw *MetaWrapper) inodeGet(partition *MetaPartition, inodeID uint64) (status int, info *proto.InodeInfo, err error) {
	req := &proto.InodeGetRequest{
		VolName:     mw.volname,
		PartitionID: partition.PartitionID,
		Inode:       inodeID,
	}
	packet := proto.NewPacketReqID()
	packet.Opcode = proto.OpMetaInodeGet // Get Inode
	err = packet.MarshalData(req)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[inodeGet]MarshalData err %v", err))
		return
	}

	// INFO: 调用 meta cluster Get Inode api
	packet, err = mw.sendToMetaPartition(partition, packet)
	if err != nil {
		return
	}

	status = parseStatus(packet.ResultCode)
	if status != statusOK {
		return
	}

	resp := new(proto.InodeGetResponse)
	err = packet.UnmarshalData(resp)
	if err != nil || resp.Info == nil {
		return
	}

	return statusOK, resp.Info, nil
}

// INFO: api调用meta cluster，在 parent partition 里新建 dentry
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

	// INFO: 调用 meta cluster Create Dentry api
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
