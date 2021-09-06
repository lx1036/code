package util

import (
	"net"
	"time"
)

type ConnTimeout struct {
	addr      string
	conn      net.Conn
	readTime  time.Duration
	writeTime time.Duration
}

func NewConnTimeout(conn net.Conn) *ConnTimeout {
	if conn == nil {
		return nil
	}

	conn.(*net.TCPConn).SetNoDelay(true)
	conn.(*net.TCPConn).SetLinger(0)
	conn.(*net.TCPConn).SetKeepAlive(true)

	return &ConnTimeout{conn: conn, addr: conn.RemoteAddr().String()}
}

func (c *ConnTimeout) Close() error {
	return c.conn.Close()
}

func (c *ConnTimeout) Read(p []byte) (n int, err error) {
	panic("implement me")
}

func (c *ConnTimeout) RemoteAddr() string {
	return c.addr
}
