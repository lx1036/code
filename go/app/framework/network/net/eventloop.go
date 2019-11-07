package net

import "k8s-lx1036/app/framework/network/netpoll"

type loop struct {
	idx         int                 // loop index in the server loops list
	svr         *server             // server in loop
	packet      []byte              // read packet buffer
	poller      *netpoll.Poller     // epoll or kqueue
	connections map[int]*Connection // loop connections fd -> conn
}
