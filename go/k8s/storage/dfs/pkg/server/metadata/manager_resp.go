package metadata

import (
	"fmt"
	"net"

	"k8s-lx1036/k8s/storage/dfs/pkg/util/proto"

	"k8s.io/klog/v2"
)

// Reply data through tcp connection to the client.
func (m *metadataManager) respondToClient(conn net.Conn, p *proto.Packet) (err error) {
	// Handle panic
	defer func() {
		if r := recover(); r != nil {
			switch data := r.(type) {
			case error:
				err = data
			default:
				err = fmt.Errorf(data.(string))
			}
		}
	}()

	// process data and send reply though specified tcp connection.
	err = p.WriteToConn(conn)
	if err != nil {
		klog.Errorf("response to client[%v], request[%s], response packet[%s]", err, p.GetOpMsg(), p.GetResultMsg())
	}
	return
}
