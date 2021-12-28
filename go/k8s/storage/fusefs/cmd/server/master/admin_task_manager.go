package master

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"
)

//const
const (
	// the maximum number of tasks that can be handled each time
	MaxTaskNum         = 30
	TaskWorkerInterval = time.Second * time.Duration(2)
)

// AdminTaskManager sends administration commands to the metaNode.
type AdminTaskManager struct {
	sync.RWMutex

	clusterID  string
	targetAddr string
	TaskMap    map[string]*proto.AdminTask
	exitCh     chan struct{}
	connPool   *util.ConnectPool
}

func newAdminTaskManager(targetAddr, clusterID string) *AdminTaskManager {
	sender := &AdminTaskManager{
		targetAddr: targetAddr,
		clusterID:  clusterID,
		TaskMap:    make(map[string]*proto.AdminTask),
		exitCh:     make(chan struct{}, 1),
		connPool:   util.NewConnectPool(),
	}
	go sender.process()

	return sender
}

func (sender *AdminTaskManager) process() {

}

func (sender *AdminTaskManager) getConn() (*net.TCPConn, error) {
	connect, err := net.Dial("tcp", sender.targetAddr)
	if err != nil {
		return nil, err
	}

	conn := connect.(*net.TCPConn)
	_ = conn.SetKeepAlive(true)
	return conn, nil
}

func (sender *AdminTaskManager) syncSendAdminTask(task *proto.AdminTask) (packet *proto.Packet, err error) {
	conn, err := sender.getConn()
	if err != nil {
		return nil, err
	}

	packet, _ = sender.buildPacket(task)
	if err = packet.WriteToConn(conn); err != nil {
		return nil, err
	}
	if err = packet.ReadFromConn(conn, proto.SyncSendTaskDeadlineTime); err != nil {
		return nil, err
	}
	if packet.ResultCode != proto.OpOk {
		return nil, fmt.Errorf("result code[%v],msg[%v]", packet.ResultCode, string(packet.Data))
	}

	return packet, nil
}

func (sender *AdminTaskManager) buildPacket(task *proto.AdminTask) (packet *proto.Packet, err error) {
	packet = proto.NewPacket()
	packet.Opcode = task.OpCode
	packet.ReqID = proto.GenerateRequestID()
	packet.PartitionID = task.PartitionID
	body, _ := json.Marshal(task)
	packet.Size = uint32(len(body))
	packet.Data = body
	return packet, nil
}
