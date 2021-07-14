package metadata

import (
	"io"
	"net"

	"k8s-lx1036/k8s/storage/sunfs/pkg/util/proto"

	"k8s.io/klog/v2"
)

// StartTcpService binds and listens to the specified port.
func (m *MetaNode) startServer() (err error) {
	// initialize and start the server.
	m.httpStopC = make(chan uint8)
	ln, err := net.Listen("tcp", ":"+m.listen)
	if err != nil {
		return
	}
	go func(stopC chan uint8) {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			select {
			case <-stopC:
				return
			default:
			}
			if err != nil {
				continue
			}
			go m.serveConn(conn, stopC)
		}
	}(m.httpStopC)
	klog.Infof("start server over...")
	return
}

// Read data from the specified tcp connection until the connection is closed by the remote or the tcp service is down.
func (m *MetaNode) serveConn(conn net.Conn, stopC chan uint8) {
	defer conn.Close()
	c := conn.(*net.TCPConn)
	c.SetKeepAlive(true)
	c.SetNoDelay(true)
	remoteAddr := conn.RemoteAddr().String()
	for {
		select {
		case <-stopC:
			return
		default:
		}
		p := &proto.Packet{}
		if err := p.ReadFromConn(conn, proto.NoReadDeadlineTime); err != nil {
			if err != io.EOF {
				klog.Errorf("serve MetaNode remote[%v] %v error: %v", remoteAddr, p.GetUniqueLogId(), err)
			}
			return
		}

		// Start a goroutine for packet handling. Do not block connection read goroutine.
		go func() {
			if err := m.handlePacket(conn, p, remoteAddr); err != nil {
				klog.Errorf("serve operatorPkg: %v", err)
				return
			}
		}()
	}
}

func (m *MetaNode) handlePacket(conn net.Conn, p *proto.Packet, remoteAddr string) (err error) {
	// Handle request
	err = m.metadataManager.HandleMetadataOperation(conn, p, remoteAddr)
	return
}
