package client

const (
	DefaultPageSize        = 1 << 17 // 128KB
	DefaultBlksize         = 1 << 20 // 1M
	DefaultBufSize         = 5 << 30 //5GB
	DefaultBufDirtyMax     = 6 << 20 // 6MB
	DefaultPartBlocks      = 5
	DefaultFlushInterval   = 5
	DefaultFlushWait       = 30
	DefaultMaxNameLen      = uint32(256)
	DefaultBlkExpiration   = 10
	DefaultReadAheadSize   = 100 << 20 // 100MB
	DefaultReadThreshold   = 20 << 20  // 20MB
	DefautlBufFreeTime     = 300
	DefaultMinBufferBlocks = 15
	DefaultMaxReleaseCount = 400
)
