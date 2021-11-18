package table

import (
	"bytes"
	"github.com/osrg/gobgp/pkg/packet/bgp"
)

type cage struct {
	attrsBytes []byte
	paths      []*Path
}

func newCage(b []byte, path *Path) *cage {
	return &cage{
		attrsBytes: b,
		paths:      []*Path{path},
	}
}

type packer struct {
	eof    bool
	family bgp.RouteFamily
	total  uint32
}

type packerMP struct {
	packer
	paths       []*Path
	withdrawals []*Path
}

func (p *packerMP) add(path *Path) {
	p.packer.total++
	
	if path.IsEOR() {
		p.packer.eof = true
		return
	}
	
	if path.IsWithdraw {
		p.withdrawals = append(p.withdrawals, path)
		return
	}
	
	p.paths = append(p.paths, path)
}

func createMPReachMessage(path *Path) *bgp.BGPMessage {
	oattrs := path.GetPathAttrs()
	attrs := make([]bgp.PathAttributeInterface, 0, len(oattrs))
	for _, a := range oattrs {
		if a.GetType() == bgp.BGP_ATTR_TYPE_MP_REACH_NLRI {
			attrs = append(attrs, bgp.NewPathAttributeMpReachNLRI(path.GetNexthop().String(), []bgp.AddrPrefixInterface{path.GetNlri()}))
		} else {
			attrs = append(attrs, a)
		}
	}
	return bgp.NewBGPUpdateMessage(nil, attrs, nil)
}

func (p *packerMP) pack(options ...*bgp.MarshallingOption) []*bgp.BGPMessage {
	msgs := make([]*bgp.BGPMessage, 0, p.packer.total)
	
	for _, path := range p.withdrawals {
		nlris := []bgp.AddrPrefixInterface{path.GetNlri()}
		msgs = append(msgs, bgp.NewBGPUpdateMessage(nil, []bgp.PathAttributeInterface{bgp.NewPathAttributeMpUnreachNLRI(nlris)}, nil))
	}
	
	for _, path := range p.paths {
		msgs = append(msgs, createMPReachMessage(path))
	}
	
	if p.eof {
		msgs = append(msgs, bgp.NewEndOfRib(p.family))
	}
	return msgs
}

func newPackerMP(f bgp.RouteFamily) *packerMP {
	return &packerMP{
		packer: packer{
			family: f,
		},
		withdrawals: make([]*Path, 0),
		paths:       make([]*Path, 0),
	}
}

type packerV4 struct {
	packer
	hashmap     map[uint32][]*cage
	mpPaths     []*Path
	withdrawals []*Path
}

func (p *packerV4) add(path *Path) {
	p.packer.total++
	
	if path.IsEOR() {
		p.packer.eof = true
		return
	}
	
	if path.IsWithdraw {
		p.withdrawals = append(p.withdrawals, path)
		return
	}
	
	if path.GetNexthop().To4() == nil {
		// RFC 5549
		p.mpPaths = append(p.mpPaths, path)
		return
	}
	
	key := path.GetHash()
	attrsB := bytes.NewBuffer(make([]byte, 0))
	for _, v := range path.GetPathAttrs() {
		b, _ := v.Serialize()
		attrsB.Write(b)
	}
	
	if cages, y := p.hashmap[key]; y {
		added := false
		for _, c := range cages {
			if bytes.Equal(c.attrsBytes, attrsB.Bytes()) {
				c.paths = append(c.paths, path)
				added = true
				break
			}
		}
		if !added {
			p.hashmap[key] = append(p.hashmap[key], newCage(attrsB.Bytes(), path))
		}
	} else {
		p.hashmap[key] = []*cage{newCage(attrsB.Bytes(), path)}
	}
}

func (p *packerV4) pack(options ...*bgp.MarshallingOption) []*bgp.BGPMessage {
	split := func(max int, paths []*Path) ([]*bgp.IPAddrPrefix, []*Path) {
		nlris := make([]*bgp.IPAddrPrefix, 0, max)
		i := 0
		if max > len(paths) {
			max = len(paths)
		}
		for ; i < max; i++ {
			nlris = append(nlris, paths[i].GetNlri().(*bgp.IPAddrPrefix))
		}
		return nlris, paths[i:]
	}
	addpathNLRILen := 0
	if bgp.IsAddPathEnabled(false, p.packer.family, options) {
		addpathNLRILen = 4
	}
	// Header + Update (WithdrawnRoutesLen +
	// TotalPathAttributeLen + attributes + maxlen of NLRI).
	// the max size of NLRI is 5bytes (plus 4bytes with addpath enabled)
	maxNLRIs := func(attrsLen int) int {
		return (bgp.BGP_MAX_MESSAGE_LENGTH - (19 + 2 + 2 + attrsLen)) / (5 + addpathNLRILen)
	}
	
	loop := func(attrsLen int, paths []*Path, cb func([]*bgp.IPAddrPrefix)) {
		max := maxNLRIs(attrsLen)
		var nlris []*bgp.IPAddrPrefix
		for {
			nlris, paths = split(max, paths)
			if len(nlris) == 0 {
				break
			}
			cb(nlris)
		}
	}
	
	msgs := make([]*bgp.BGPMessage, 0, p.packer.total)
	
	loop(0, p.withdrawals, func(nlris []*bgp.IPAddrPrefix) {
		msgs = append(msgs, bgp.NewBGPUpdateMessage(nlris, nil, nil))
	})
	
	for _, cages := range p.hashmap {
		for _, c := range cages {
			paths := c.paths
			
			attrs := paths[0].GetPathAttrs()
			// we can apply a fix here when gobgp receives from MP peer
			// and propagtes to non-MP peer
			// we should make sure that next-hop exists in pathattrs
			// while we build the update message
			// we do not want to modify the `path` though
			if paths[0].getPathAttr(bgp.BGP_ATTR_TYPE_NEXT_HOP) == nil {
				attrs = append(attrs, bgp.NewPathAttributeNextHop(paths[0].GetNexthop().String()))
			}
			// if we have ever reach here
			// there is no point keeping MP_REACH_NLRI in the announcement
			attrs_without_mp := make([]bgp.PathAttributeInterface, 0, len(attrs))
			for _, attr := range attrs {
				if attr.GetType() != bgp.BGP_ATTR_TYPE_MP_REACH_NLRI {
					attrs_without_mp = append(attrs_without_mp, attr)
				}
			}
			attrsLen := 0
			for _, a := range attrs_without_mp {
				attrsLen += a.Len()
			}
			
			loop(attrsLen, paths, func(nlris []*bgp.IPAddrPrefix) {
				msgs = append(msgs, bgp.NewBGPUpdateMessage(nil, attrs_without_mp, nlris))
			})
		}
	}
	
	for _, path := range p.mpPaths {
		msgs = append(msgs, createMPReachMessage(path))
	}
	
	if p.eof {
		msgs = append(msgs, bgp.NewEndOfRib(p.family))
	}
	return msgs
}

func newPackerV4(f bgp.RouteFamily) *packerV4 {
	return &packerV4{
		packer: packer{
			family: f,
		},
		hashmap:     make(map[uint32][]*cage),
		withdrawals: make([]*Path, 0),
		mpPaths:     make([]*Path, 0),
	}
}

func newPacker(f bgp.RouteFamily) packerInterface {
	switch f {
	case bgp.RF_IPv4_UC:
		return newPackerV4(bgp.RF_IPv4_UC)
	default:
		return newPackerMP(f)
	}
}

func UpdatePathAttrs2ByteAs(msg *bgp.BGPUpdate) error {
	ps := msg.PathAttributes
	msg.PathAttributes = make([]bgp.PathAttributeInterface, len(ps))
	copy(msg.PathAttributes, ps)
	var asAttr *bgp.PathAttributeAsPath
	idx := 0
	for i, attr := range msg.PathAttributes {
		if a, ok := attr.(*bgp.PathAttributeAsPath); ok {
			asAttr = a
			idx = i
			break
		}
	}
	
	if asAttr == nil {
		return nil
	}
	
	as4Params := make([]*bgp.As4PathParam, 0, len(asAttr.Value))
	as2Params := make([]bgp.AsPathParamInterface, 0, len(asAttr.Value))
	mkAs4 := false
	for _, param := range asAttr.Value {
		segType := param.GetType()
		asList := param.GetAS()
		as2Path := make([]uint16, 0, len(asList))
		for _, as := range asList {
			if as > (1<<16)-1 {
				mkAs4 = true
				as2Path = append(as2Path, bgp.AS_TRANS)
			} else {
				as2Path = append(as2Path, uint16(as))
			}
		}
		as2Params = append(as2Params, bgp.NewAsPathParam(segType, as2Path))
		
		// RFC 6793 4.2.2 Generating Updates
		//
		// Whenever the AS path information contains the AS_CONFED_SEQUENCE or
		// AS_CONFED_SET path segment, the NEW BGP speaker MUST exclude such
		// path segments from the AS4_PATH attribute being constructed.
		switch segType {
		case bgp.BGP_ASPATH_ATTR_TYPE_CONFED_SEQ, bgp.BGP_ASPATH_ATTR_TYPE_CONFED_SET:
			// pass
		default:
			if as4param, ok := param.(*bgp.As4PathParam); ok {
				as4Params = append(as4Params, as4param)
			}
		}
	}
	msg.PathAttributes[idx] = bgp.NewPathAttributeAsPath(as2Params)
	if mkAs4 {
		msg.PathAttributes = append(msg.PathAttributes, bgp.NewPathAttributeAs4Path(as4Params))
	}
	return nil
}

func UpdatePathAggregator2ByteAs(msg *bgp.BGPUpdate) {
	as := uint32(0)
	var addr string
	for i, attr := range msg.PathAttributes {
		switch agg := attr.(type) {
		case *bgp.PathAttributeAggregator:
			addr = agg.Value.Address.String()
			if agg.Value.AS > (1<<16)-1 {
				as = agg.Value.AS
				msg.PathAttributes[i] = bgp.NewPathAttributeAggregator(uint16(bgp.AS_TRANS), addr)
			} else {
				msg.PathAttributes[i] = bgp.NewPathAttributeAggregator(uint16(agg.Value.AS), addr)
			}
		}
	}
	if as != 0 {
		msg.PathAttributes = append(msg.PathAttributes, bgp.NewPathAttributeAs4Aggregator(as, addr))
	}
}


type packerInterface interface {
	add(*Path)
	pack(options ...*bgp.MarshallingOption) []*bgp.BGPMessage
}

func CreateUpdateMsgFromPaths(pathList []*Path, options ...*bgp.MarshallingOption) []*bgp.BGPMessage {
	msgs := make([]*bgp.BGPMessage, 0, len(pathList))
	
	m := make(map[bgp.RouteFamily]packerInterface)
	for _, path := range pathList {
		f := path.GetRouteFamily()
		if _, y := m[f]; !y {
			m[f] = newPacker(f)
		}
		m[f].add(path)
	}
	
	for _, p := range m {
		msgs = append(msgs, p.pack(options...)...)
	}
	return msgs
}
