package meta

import (
	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"

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
