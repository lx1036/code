package meta

import (
	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
	"syscall"

	"k8s.io/klog/v2"
)

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
