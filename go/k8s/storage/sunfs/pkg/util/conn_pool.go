package util

import (
	"net"
	"sync"
	"time"
)

const (
	ConnectIdleTime = 30
)

type Object struct {
	conn *net.TCPConn
	idle int64
}

type Pool struct {
	objects chan *Object
	mincap  int
	maxcap  int
	target  string
	timeout int64
}

func (p *Pool) autoRelease() {
	connectLen := len(p.objects)
	for i := 0; i < connectLen; i++ {
		select {
		case o := <-p.objects:
			if time.Now().UnixNano()-int64(o.idle) > p.timeout {
				o.conn.Close()
			} else {
				p.PutConnectObjectToPool(o)
			}
		default:
			return
		}
	}
}

func (p *Pool) initAllConnect() {
	for i := 0; i < p.mincap; i++ {
		c, err := net.Dial("tcp", p.target)
		if err == nil {
			conn := c.(*net.TCPConn)
			conn.SetKeepAlive(true)
			conn.SetNoDelay(true)
			o := &Object{conn: conn, idle: time.Now().UnixNano()}
			p.PutConnectObjectToPool(o)
		}
	}
}

func (p *Pool) PutConnectObjectToPool(o *Object) {
	select {
	case p.objects <- o:
		return
	default:
		if o.conn != nil {
			o.conn.Close()
		}
		return
	}
}

func (p *Pool) GetConnectFromPool() (c *net.TCPConn, err error) {
	var (
		o *Object
	)
	for i := 0; i < len(p.objects); i++ {
		select {
		case o = <-p.objects:
			if time.Now().UnixNano()-int64(o.idle) > p.timeout {
				o.conn.Close()
				o = nil
				break
			}
			return o.conn, nil
		default:
			return p.NewConnect(p.target)
		}
	}

	return p.NewConnect(p.target)
}

func (p *Pool) NewConnect(target string) (c *net.TCPConn, err error) {
	var connect net.Conn
	connect, err = net.Dial("tcp", p.target)
	if err == nil {
		conn := connect.(*net.TCPConn)
		conn.SetKeepAlive(true)
		conn.SetNoDelay(true)
		c = conn
	}
	return
}

func NewPool(min, max int, timeout int64, target string) (p *Pool) {
	p = new(Pool)
	p.mincap = min
	p.maxcap = max
	p.target = target
	p.objects = make(chan *Object, max)
	p.timeout = timeout
	p.initAllConnect()
	return p
}

type ConnectPool struct {
	sync.RWMutex
	pools   map[string]*Pool
	mincap  int
	maxcap  int
	timeout int64
}

func (cp *ConnectPool) autoRelease() {
	for {
		pools := make([]*Pool, 0)
		cp.RLock()
		for _, pool := range cp.pools {
			pools = append(pools, pool)
		}
		cp.RUnlock()
		for _, pool := range pools {
			pool.autoRelease()
		}
		time.Sleep(time.Second)
	}
}

func (cp *ConnectPool) GetConnect(targetAddr string) (c *net.TCPConn, err error) {
	cp.RLock()
	defer cp.RUnlock()
	pool, ok := cp.pools[targetAddr]
	if !ok {
		cp.Lock()
		pool = NewPool(cp.mincap, cp.maxcap, cp.timeout, targetAddr)
		cp.pools[targetAddr] = pool
		cp.Unlock()
	}

	return pool.GetConnectFromPool()
}

func (cp *ConnectPool) PutConnect(c *net.TCPConn, forceClose bool) {
	if c == nil {
		return
	}
	if forceClose {
		c.Close()
		return
	}
	addr := c.RemoteAddr().String()
	cp.RLock()
	defer cp.RUnlock()
	pool, ok := cp.pools[addr]
	if !ok {
		c.Close()
		return
	}
	object := &Object{conn: c, idle: time.Now().UnixNano()}
	pool.PutConnectObjectToPool(object)

	return
}

func NewConnectPool() *ConnectPool {
	connectPool := &ConnectPool{
		pools:   make(map[string]*Pool),
		mincap:  5,
		maxcap:  80,
		timeout: int64(time.Second * ConnectIdleTime),
	}

	go connectPool.autoRelease()

	return connectPool
}
