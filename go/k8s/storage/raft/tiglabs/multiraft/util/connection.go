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
	if c.readTime.Nanoseconds() > 0 {
		err = c.conn.SetReadDeadline(time.Now().Add(c.readTime))
		if err != nil {
			return
		}
	}

	n, err = c.conn.Read(p)
	return
}

func (c *ConnTimeout) Write(p []byte) (n int, err error) {
	if c.writeTime.Nanoseconds() > 0 {
		err = c.conn.SetWriteDeadline(time.Now().Add(c.writeTime))
		if err != nil {
			return
		}
	}

	n, err = c.conn.Write(p)
	return
}

func (c *ConnTimeout) RemoteAddr() string {
	return c.addr
}
