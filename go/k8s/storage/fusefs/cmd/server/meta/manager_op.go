package meta

import (
	"encoding/json"
	"net"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

// Handle OpCreate inode.
func (m *metadataManager) opCreateInode(conn net.Conn, p *proto.Packet, remoteAddr string) error {
	req := &proto.CreateInodeRequest{}

	if err := json.Unmarshal(p.Data, req); err != nil {
		//p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		m.respondToClient(conn, p)
		return err
	}
	partition, err := m.getPartition(req.PartitionID)
	if err != nil {
		//p.PacketErrorWithBody(proto.OpNotExistErr, []byte(err.Error()))
		m.respondToClient(conn, p)
		return err
	}
	// TODO: 如果不是leader，可以 proxy request to leader
	/*if !m.serveProxy(conn, mp, p) {
		return err
	}*/
	err = partition.CreateInode(req, p)
	// reply the operation result to the client through TCP
	m.respondToClient(conn, p)
	klog.Infof("%s [opCreateInode] req: %d - %v, resp: %v, body: %s", remoteAddr, p.GetReqID(), req, p.GetResultMsg(), p.Data)

	return nil
}
