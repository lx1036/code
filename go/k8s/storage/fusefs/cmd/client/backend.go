package client

type Backend interface {
	Read(file string, offset int64, data []byte) (int, error)
}
