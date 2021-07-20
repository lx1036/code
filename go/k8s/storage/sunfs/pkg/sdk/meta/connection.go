package meta

import (
	"errors"
	"fmt"
	"net"

	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
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

	// INFO: meta cluster 的地址
	addr = mp.LeaderAddr
	if addr == "" {
		return nil, errors.New(fmt.Sprintf("sendToMetaPartition failed: leader addr empty, req(%v) mp(%v)", req, mp))
	}
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
