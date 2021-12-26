package meta

import (
	"fmt"
	"io"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"net"

	"k8s.io/klog/v2"
)

// INFO: tcp 监听在 9021 port，没有走 http 协议
func (m *MetaNode) startTCPServer() (err error) {
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

func (m *MetaNode) stopTCPServer() {
	if m.httpStopC != nil {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf(fmt.Sprintf("action[StopTcpServer],err:%v", r))
			}
		}()
		close(m.httpStopC)
	}
}

// INFO: 从 tcp connection 读数据
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
			if err := m.metadataManager.HandleMetadataOperation(conn, p, remoteAddr); err != nil {
				klog.Errorf("serve operatorPkg: %v", err)
				return
			}
		}()
	}
}
