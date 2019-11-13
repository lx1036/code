package net

type eventLoopGroup struct {
	nextLoopIndex int
	eventLoops    []*loop
	size          int
}
