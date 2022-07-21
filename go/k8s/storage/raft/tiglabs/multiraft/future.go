package multiraft

type respErr struct {
	errCh chan error
}

func (e *respErr) init() {
	e.errCh = make(chan error, 1)
}

func (e *respErr) respond(err error) {
	e.errCh <- err
	close(e.errCh)
}

// Future the future
type Future struct {
	respErr
	respCh chan interface{}
}

func (f *Future) respond(resp interface{}, err error) {
	if err == nil {
		f.respCh <- resp
		close(f.respCh)
	} else {
		f.respErr.respond(err)
	}
}

// AsyncResponse export channels
func (f *Future) AsyncResponse() (respCh <-chan interface{}, errCh <-chan error) {
	return f.respCh, f.errCh
}

func newFuture() *Future {
	f := &Future{
		respCh: make(chan interface{}, 1),
	}
	f.init()
	return f
}
