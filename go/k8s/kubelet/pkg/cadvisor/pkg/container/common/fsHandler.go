package common

type FsHandler interface {
	Start()
	Usage() FsUsage
	Stop()
}

type FsUsage struct {
	BaseUsageBytes  uint64
	TotalUsageBytes uint64
	InodeUsage      uint64
}
