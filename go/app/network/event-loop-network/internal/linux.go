package internal

import "syscall"

// Poll ...
type Poll struct {
	fd    int // epoll fd
	wfd   int // wake fd
	notes noteQueue
}

// AddRead ...
func (poll *Poll) AddRead(fd int) {
	if err := syscall.EpollCtl(poll.fd, syscall.EPOLL_CTL_ADD, fd,
		&syscall.EpollEvent{Fd: int32(fd),
			Events: syscall.EPOLLIN,
		},
	); err != nil {
		panic(err)
	}
}

func (poll *Poll) Trigger(errClosing interface{}) {

}

func (poll *Poll) Close() {

}

// OpenPoll ...
func OpenPoll() *Poll {
	l := new(Poll)
	p, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	l.fd = p
	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if e0 != 0 {
		syscall.Close(p)
		panic(err)
	}
	l.wfd = int(r0)
	l.AddRead(l.wfd)
	return l
}
