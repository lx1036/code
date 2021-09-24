package backend

type ReadTx interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type baseReadTx struct {
}

type readTx struct {
	baseReadTx
}

type concurrentReadTx struct {
	baseReadTx
}
