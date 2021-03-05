package util

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
	TB
	PB
)

const (
	TaskWorkerInterval = 1
	BlockCount         = 1024
	BlockSize          = 65536 * 2
	PerBlockCrcSize    = 4
	PacketHeaderSize   = 31
	BlockHeaderSize    = 4096
)

const (
	DefaultTinySizeLimit = 1 * MB
)
