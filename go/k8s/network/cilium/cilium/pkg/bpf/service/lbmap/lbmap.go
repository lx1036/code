package lbmap

import "github.com/cilium/cilium/pkg/u8proto"

// LBBPFMap is an implementation of the LBMap interface.
type LBBPFMap struct{}

func (*LBBPFMap) UpdateOrInsertService() {

	svcKey = NewService4Key(svcIP, svcPort, u8proto.ANY, svcScope, 0)

}
