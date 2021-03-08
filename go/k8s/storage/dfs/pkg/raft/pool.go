package raft

import "sync"

var pool = newPoolFactory()

type poolFactory struct {
	applyPool    *sync.Pool
	proposalPool *sync.Pool
}

func (f *poolFactory) getProposal() *proposal {
	p := f.proposalPool.Get().(*proposal)
	p.data = nil
	p.future = nil
	return p
}

func newPoolFactory() *poolFactory {
	return &poolFactory{
		applyPool: &sync.Pool{
			New: func() interface{} {
				return new(apply)
			},
		},

		proposalPool: &sync.Pool{
			New: func() interface{} {
				return new(proposal)
			},
		},
	}
}
