package backend

type Backend interface {
	BatchTx() BatchTx
	// ReadTx 读事务/并发读
	ReadTx() ReadTx
	// ConcurrentReadTx non-blocking read transaction
	ConcurrentReadTx() ReadTx

	Snapshot() Snapshot

	ForceCommit()
	Close() error
}

type backend struct {
	batchTx *batchTxBuffered
	readTx  *readTx
}

type BackendConfig struct {
}

func (b *backend) BatchTx() BatchTx {
	return b.batchTx
}

func (b *backend) ReadTx() ReadTx {
	return b.readTx
}
