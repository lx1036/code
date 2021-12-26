package meta

import (
	"encoding/json"
	"fmt"
	"net"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

func (m *metadataManager) HandleMetadataOperation(conn net.Conn, p *proto.Packet, remoteAddr string) error {
	var err error
	switch p.Opcode {
	case proto.OpCreateMetaPartition:
		err = m.opCreateMetaPartition()

	case proto.OpMetaCreateInode:
		err = m.opCreateInode(conn, p, remoteAddr)
	//case proto.OpMetaLinkInode:
	//	err = m.opMetaLinkInode(conn, p, remoteAddr)
	//case proto.OpMetaUnlinkInode:
	//	err = m.opMetaUnlinkInode(conn, p, remoteAddr)
	//case proto.OpMetaInodeGet:
	//	err = m.opMetaInodeGet(conn, p, remoteAddr)
	//case proto.OpMetaEvictInode:
	//	err = m.opMetaEvictInode(conn, p, remoteAddr)
	//case proto.OpMetaSetattr:
	//	err = m.opSetAttr(conn, p, remoteAddr)
	//case proto.OpMetaCreateDentry:
	//	err = m.opCreateDentry(conn, p, remoteAddr)
	//case proto.OpMetaDeleteDentry:
	//	err = m.opDeleteDentry(conn, p, remoteAddr)
	//case proto.OpMetaUpdateDentry:
	//	err = m.opUpdateDentry(conn, p, remoteAddr)
	//case proto.OpMetaReadDir:
	//	err = m.opReadDir(conn, p, remoteAddr)
	//case proto.OpCreateMetaPartition:
	//	err = m.opCreateMetaPartition(conn, p, remoteAddr)
	//case proto.OpMetaNodeHeartbeat:
	//	err = m.opMasterHeartbeat(conn, p, remoteAddr)
	//case proto.OpMetaLookup:
	//	err = m.opMetaLookup(conn, p, remoteAddr)
	//case proto.OpMetaLookupName:
	//	err = m.opMetaLookupName(conn, p, remoteAddr)
	//case proto.OpDeleteMetaPartition:
	//	err = m.opDeleteMetaPartition(conn, p, remoteAddr)
	//case proto.OpUpdateMetaPartition:
	//	err = m.opUpdateMetaPartition(conn, p, remoteAddr)
	//case proto.OpLoadMetaPartition:
	//	err = m.opLoadMetaPartition(conn, p, remoteAddr)
	//case proto.OpDecommissionMetaPartition:
	//	err = m.opDecommissionMetaPartition(conn, p, remoteAddr)
	//case proto.OpAddMetaPartitionRaftMember:
	//	err = m.opAddMetaPartitionRaftMember(conn, p, remoteAddr)
	//case proto.OpRemoveMetaPartitionRaftMember:
	//	err = m.opRemoveMetaPartitionRaftMember(conn, p, remoteAddr)
	//case proto.OpMetaPartitionTryToLeader:
	//	err = m.opMetaPartitionTryToLeader(conn, p, remoteAddr)
	//case proto.OpMetaBatchInodeGet:
	//	err = m.opMetaBatchInodeGet(conn, p, remoteAddr)
	default:
		err = fmt.Errorf("%s unknown Opcode: %d, reqId: %d", remoteAddr,
			p.Opcode, p.GetReqID())
	}
	if err != nil {
		err = fmt.Errorf("%s [%s] req: %d - %v", remoteAddr, p.GetOpMsg(), p.GetReqID(), err)
	}

	return err
}

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
