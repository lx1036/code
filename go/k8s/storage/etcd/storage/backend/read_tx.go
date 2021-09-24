package backend

type ReadTx interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}
