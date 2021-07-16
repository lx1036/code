package meta

import (
	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
	"net"
)

type MetaConn struct {
	conn *net.TCPConn
	id   uint64 //PartitionID
	addr string //MetaNode addr
}

func (mw *MetaWrapper) sendToMetaPartition(mp *MetaPartition, req *proto.Packet) (*proto.Packet, error) {
	var (
		resp           *proto.Packet
		err            error
		addr           string
		metaConnection *MetaConn
	)

	metaConnection, err = mw.getConn(mp.PartitionID, addr)
	if err != nil {
		return nil, err
	}

	resp, err = metaConnection.send(req)
	mw.putConn(metaConnection, err)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (mw *MetaWrapper) getConn(partitionID uint64, addr string) (*MetaConn, error) {
	conn, err := mw.conns.GetConnect(addr)
	if err != nil {
		return nil, err
	}

	mc := &MetaConn{conn: conn, id: partitionID, addr: addr}
	return mc, nil
}

func (mc *MetaConn) send(req *proto.Packet) (resp *proto.Packet, err error) {
	err = req.WriteToConn(mc.conn)
	if err != nil {
		return nil, err
	}

	resp = proto.NewPacket()
	err = resp.ReadFromConn(mc.conn, proto.ReadDeadlineTime)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (mw *MetaWrapper) putConn(mc *MetaConn, err error) {
	if err != nil {
		mw.conns.PutConnect(mc.conn, true)
	} else {
		mw.conns.PutConnect(mc.conn, false)
	}
}
