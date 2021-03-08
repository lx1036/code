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

func (p *Pool) PutConnectObjectToPool(o *Object) {
	select {
	case p.objects <- o: // 如果buffer chan objects阻塞了执行default语句
		return
	default:
		if o.conn != nil {
			o.conn.Close()
		}
		return
	}
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
