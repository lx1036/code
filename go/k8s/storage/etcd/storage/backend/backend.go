package backend

type Backend interface {
	BatchTx() BatchTx
}
