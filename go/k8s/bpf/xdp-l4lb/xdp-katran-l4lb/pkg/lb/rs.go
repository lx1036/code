package lb

import (
	"context"
	"net"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-katran-l4lb/pkg/rpc"

	"github.com/sirupsen/logrus"
)

type IPAddress string

func (address *IPAddress) isIPv4() bool {
	ip := net.ParseIP(string(*address))
	if ip == nil {
		return false
	}

	return ip.To4() != nil
}

func (address *IPAddress) isIPv6() bool {
	ip := net.ParseIP(string(*address))
	if ip == nil {
		return false
	}

	return ip.To16() != nil
}

type RealMeta struct {
	// vip's number
	num uint32
	/**
	 * one real could be used by multiple vips
	 * we will delete real (recycle it's num),
	 * only when refcount would be equal to zero
	 */
	refCount uint32
	flags    uint8
}

// RealPos This struct show on which position real w/ specified opaque id should be located on ch ring.
type RealPos struct {
	real     uint32
	position uint32
}

type RealsIdCallback interface {
	RealsIdCallback()

	onRealAdded(real *IPAddress, id uint32)

	onRealDeleted(real *IPAddress, id uint32)
}

type RealDefinition struct {
	daddr struct {
		daddr   uint32
		v6daddr uint32
	}
	flags uint8
}

func parseAddrToBe(addr *IPAddress, bigEndian bool) RealDefinition {
	translatedAddr := RealDefinition{
		daddr: struct {
			daddr   uint32
			v6daddr uint32
		}{},
		flags: 0,
	}
	if addr.isIPv4() {
		translatedAddr.flags = 0
		if bigEndian {
			translatedAddr.daddr.daddr = addr
		} else {

		}
	} else {

	}

	return translatedAddr
}

// information about new real
type NewReal struct {
	address IPAddress
	weight  uint32
	flags   uint8
}

func translateRealObject(real *rpc.Real) *NewReal {
	return &NewReal{
		address: IPAddress(real.Address),
		weight:  uint32(real.Weight),
		flags:   uint8(real.Flags),
	}
}

type UpdateReal struct {
	action      ModifyAction
	updatedReal Endpoint
}

func (lb *OpenLb) AddRealForVip(ctx context.Context, vip *rpc.RealForVip) bool {
	//func (lb *OpenLb) AddRealForVip(real *NewReal, vip *VipKey) bool {
	if lb.config.disableForwarding {
		logrus.Warn("AddRealForVip called on non-forwarding instance")
		return false
	}

	vipKey := translateVipObject(vip.GetVip())
	rs := translateRealObject(vip.GetReal())

	return lb.modifyRealsForVip(ADD, rs, vipKey)
}

func (lb *OpenLb) DelRealForVip(ctx context.Context, vip *rpc.RealForVip) bool {
	//func (lb *OpenLb) DelRealForVip(real *NewReal, vip *VipKey) bool {
	if lb.config.disableForwarding {
		logrus.Warn("AddRealForVip called on non-forwarding instance")
		return false
	}

	return lb.modifyRealsForVip(DEL, reals, vip)
}

func (lb *OpenLb) modifyRealsForVip(action ModifyAction, reals []*NewReal, vipKey *VipKey) bool {
	if lb.config.disableForwarding {
		logrus.Warn("modifyRealsForVip called on non-forwarding instance")
		return false
	}

	var updateReal UpdateReal
	var updateReals []UpdateReal
	updateReal.action = action
	vip, ok := lb.vips[*vipKey]
	if !ok {
		logrus.Errorf("trying to modify reals for non existing vip: %s", vipKey.string())
		return false
	}

	currentReals := vip.getReals()
	for _, rs := range reals {
		if lb.validateAddress(string(rs.address), false) == INVALID {
			logrus.Errorf("Invalid real's address: %s", rs.address)
			continue
		}

		if action == DEL {
			realMeta, ok := lb.reals[&rs.address]
			if !ok {
				logrus.Errorf("trying to delete non-existing real: %s", rs.address)
				continue
			}

			// this real doesn't belong to this vip
			if _, ok = currentReals[rs]; !ok {
				logrus.Errorf("trying to delete non-existing real for the VIP: %s", vipKey.string())
				continue
			}

			updateReal.updatedReal.num = realMeta.num
			lb.decreaseRefCountForReal(&rs.address)
		} else {
			realMeta, ok := lb.reals[&rs.address]
			if !ok {

			} else {

			}

			updateReal.updatedReal.weight = rs.weight
			updateReal.updatedReal.hash = rs.address
		}

		updateReals = append(updateReals, updateReal)
	}

	chPositions := vip.batchRealsUpdate(updateReals)
	lb.programHashRing(chPositions, vip.vipNum)
	return true
}

func (lb *OpenLb) programHashRing(chPositions []RealPos, vipNum uint32) {
	if len(chPositions) == 0 {
		return
	}

	updateSize := len(chPositions)
	keys := make([]uint32, updateSize)
	values := make([]uint32, updateSize)
	if !lb.config.testing {
		for i := 0; i < updateSize; i++ {
			keys[i] = vipNum*lb.config.chRingSize + chPositions[i].position
			values[i] = chPositions[i].real
		}

		count, err := lb.chRingsMap.BatchUpdate(keys, values, nil)
		if err != nil {
			lb.stats.bpfFailedCalls++
			logrus.Errorf("can't update ch ring error: %v", err)
		}
		if count != updateSize {
			lb.stats.bpfFailedCalls++
			logrus.Errorf("can't update ch ring error: expect %d, actual %d", updateSize, count)
		}
	}
}

func (lb *OpenLb) ModifyRealsForVip(ctx context.Context, vip *rpc.ModifiedRealsForVip) (*rpc.Bool, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetRealFlags(ctx context.Context, r *rpc.Real) (*rpc.Flags, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetRealsForVip(ctx context.Context, vip *rpc.Vip) (*rpc.Reals, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) ModifyReal(ctx context.Context, meta *rpc.RealMeta) (*rpc.Bool, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) decreaseRefCountForReal(real *IPAddress) {
	realMeta, ok := lb.reals[real]
	if !ok {
		return
	}

	realMeta.refCount--
	if realMeta.refCount == 0 {
		num := realMeta.num
		// no more vips using this real
		lb.realNums = append(lb.realNums, num)
		delete(lb.numToReals, num)
		delete(lb.reals, real)

		if lb.realsIdCallback != nil {
			lb.realsIdCallback.onRealDeleted(real, num)
		}
	}
}

func (lb *OpenLb) increaseRefCountForReal(real *IPAddress, flags uint8) uint32 {
	realMeta, ok := lb.reals[real]
	if ok {
		// to keep IPv4/IPv6 specific flag
		flags &= ^V6DADDR
		realMeta.refCount++
		return realMeta.num
	}

	if len(lb.realNums) == 0 {
		return lb.config.maxReals
	}

	rnum := lb.realNums[0]
	lb.numToReals[rnum] = *real
	rmeta := RealMeta{
		num:      rnum,
		refCount: 1,
		flags:    flags,
	}
	lb.reals[real] = &rmeta
	if !lb.config.testing {
		lb.updateRealsMap(real, rnum, flags)
	}

	if lb.realsIdCallback != nil {
		lb.realsIdCallback.onRealAdded(real, rnum)
	}

	return rnum
}

func (lb *OpenLb) updateRealsMap(real *IPAddress, num uint32, flags uint8) bool {
	// to keep IPv4/IPv6 specific flag
	realDefinition := parseAddrToBe(real)
	flags &= ^V6DADDR
	realDefinition.flags |= flags
	err := lb.realsMap.Update(&num, &realDefinition, 0)
	if err != nil {
		logrus.Errorf("can't add new real, error: %v", err)
		lb.stats.bpfFailedCalls++
		return false
	}

	return true
}
