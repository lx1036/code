package lb

import (
	"context"
	"fmt"
	"sort"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-katran-l4lb/pkg/rpc"

	"github.com/sirupsen/logrus"
)

type ModifyAction int

const (
	ADD ModifyAction = iota
	DEL
)

// result of vip's lookup
type vipMeta struct {
	flags  uint32
	vipNum uint32
}

type VipKey struct {
	address string
	port    uint16
	proto   string
}

func (key *VipKey) equal(other *VipKey) bool {
	return key.proto == other.proto && key.address == other.address && key.port == other.port
}
func (key *VipKey) string() string {
	return fmt.Sprintf("%s://%s:%d", key.proto, key.address, key.port)
}

// VipDefinition vip's definition for lookup also used for hc_keys
type VipDefinition struct {
	vip struct {
		vip   string
		vipv6 string
	}
	port  uint16
	proto string
}

type Vip struct {
	/**
	 * map of reals (theirs opaque id). the value is a real's related
	 * metadata (weight and per real hash value).
	 */
	reals map[uint32]*VipRealMeta

	/**
	 * number which uniquely identifies this vip
	 * (also used as an index inside forwarding table)
	 */
	vipNum uint32

	/**
	 * ch ring which is used for this vip. we are going to use it
	 * for delta computation (between old and new ch rings)
	 */
	chRing []int

	/**
	 * hash function to generate hash ring
	 */
	chash ConsistentHash

	/**
	 * size of ch ring
	 */
	chRingSize uint32
}

func (vip *Vip) getReals() map[uint32]VipRealMeta {
	realNums := make(map[uint32]VipRealMeta)
	for key, meta := range vip.reals {
		realNums[key] = meta
	}
	return realNums
}

func (vip *Vip) batchRealsUpdate(updateReals []UpdateReal) []RealPos {
	endpoints := vip.getEndpoints(updateReals)
	return vip.calculateHashRing(endpoints)
}

func (vip *Vip) calculateHashRing(endpoints []Endpoint) []RealPos {
	var delta []RealPos
	if len(endpoints) != 0 {
		newChRing := vip.chash.generateHashRing(endpoints, vip.chRingSize)
		// compare new and old ch rings. send back only delta between em.
		for i := 0; i < int(vip.chRingSize); i++ {
			if newChRing[i] != vip.chRing[i] {
				newPos := RealPos{
					real:     uint32(newChRing[i]),
					position: uint32(i),
				}
				delta = append(delta, newPos)
				vip.chRing[i] = newChRing[i]
			}
		}
	}

	return delta
}

func (vip *Vip) getEndpoints(updateReals []UpdateReal) []Endpoint {
	var endpoints []Endpoint
	realsChanged := false
	for _, updateReal := range updateReals {
		if updateReal.action == DEL {
			delete(vip.reals, updateReal.updatedReal.num)
			realsChanged = true
		} else {
			curWeight := vip.reals[updateReal.updatedReal.num].weight
			if curWeight != updateReal.updatedReal.weight {
				vip.reals[updateReal.updatedReal.num].weight = updateReal.updatedReal.weight
				vip.reals[updateReal.updatedReal.num].hash = updateReal.updatedReal.hash
				realsChanged = true
			}
		}
	}

	if realsChanged {
		for num, vipRealMeta := range vip.reals {
			// skipping 0 weight
			if vipRealMeta.weight != 0 {
				endpoints = append(endpoints, Endpoint{
					num:    num,
					weight: vipRealMeta.weight,
					hash:   vipRealMeta.hash,
				})
			}
		}

		sort.Slice(endpoints, func(i, j int) bool {
			return endpoints[i].hash < endpoints[j].hash
		})
	}

	return endpoints
}

func translateVipObject(v *rpc.Vip) *VipKey {
	vip := &VipKey{
		address: v.GetAddress(),
		port:    uint16(v.GetPort()),
		proto:   v.GetProtocol(),
	}
	return vip
}

type VipRealMeta struct {
	weight uint32
	hash   uint64
}

func translateVipMetaObject(meta *rpc.VipMeta) *VipKey {
	v := meta.GetVip()
	vip := &VipKey{
		address: v.GetAddress(),
		port:    uint16(v.GetPort()),
		proto:   v.GetProtocol(),
	}
	return vip
}

func (lb *OpenLb) updateVipMap(action ModifyAction, vip *VipKey, meta *vipMeta) bool {
	key := lb.vipKeyToVipDefinition(vip)
	if action == ADD {
		err := lb.vipMap.Update(&key, meta, 0)
		if err != nil {
			logrus.Errorf("can't add new element into vip_map, error: %v", err)
			lb.stats.bpfFailedCalls++
			return false
		}
	} else {
		err := lb.vipMap.Delete(&key)
		if err != nil {
			logrus.Errorf("can't delete element into vip_map, error: %v", err)
			lb.stats.bpfFailedCalls++
			return false
		}
	}

	return true
}

func (lb *OpenLb) vipKeyToVipDefinition(vip *VipKey) VipDefinition {
	vipDefinition := VipDefinition{
		vip: struct {
			vip   string
			vipv6 string
		}{},
		port:  vip.port,
		proto: vip.proto,
	}

	if isIPv4(vip.address) {
		vipDefinition.vip.vip = vip.address
	} else if isIPv6(vip.address) {
		vipDefinition.vip.vipv6 = vip.address
	}

	return vipDefinition
}

func (lb *OpenLb) AddVip(ctx context.Context, meta *rpc.VipMeta) (*rpc.Bool, error) {
	if lb.config.disableForwarding {
		msg := "Ignoring addVip call on non-forwarding instance"
		logrus.Warn(msg)
		return &rpc.Bool{Success: false}, fmt.Errorf(msg)
	}

	flags := meta.Flags
	vip := translateVipMetaObject(meta)

	if lb.validateAddress(vip.address, false) == INVALID {
		msg := fmt.Sprintf("Invalid Vip address: %s", vip.address)
		logrus.Error(msg)
		return &rpc.Bool{Success: false}, fmt.Errorf(msg)
	}

	logrus.Infof("adding new vip: %s:%s:%d", vip.proto, vip.address, vip.port)

	if len(lb.vipNums) == 0 {
		msg := "exhausted vip's space"
		logrus.Info(msg)
		return &rpc.Bool{Success: false}, fmt.Errorf(msg)
	}

	vipNum := lb.vipNums[0]
	if !lb.config.testing {
		meta := vipMeta{
			flags:  uint32(flags),
			vipNum: vipNum,
		}
		if lb.updateVipMap(ADD, vip, &meta) {
			// TODO add into vips[]
			//lb.vips[]
		}
	}

	return &rpc.Bool{Success: true}, nil
}

func (lb *OpenLb) DelVip(ctx context.Context, rpcVip *rpc.Vip) (*rpc.Bool, error) {
	if lb.config.disableForwarding {
		msg := "ignoring addVip call on non-forwarding instance"
		logrus.Warn(msg)
		return &rpc.Bool{Success: false}, fmt.Errorf(msg)
	}

	vipKey := translateVipObject(rpcVip)
	logrus.Infof("deleting vip: %s:%s:%d", vipKey.proto, vipKey.address, vipKey.port)

	vip, ok := lb.vips[*vipKey]
	if !ok {
		msg := "trying to delete non-existing vip"
		logrus.Error(msg)
		return &rpc.Bool{Success: false}, fmt.Errorf(msg)
	}

	reals := vip.getReals()
	// decreasing ref count for reals. delete em if it became 0
	for num, _ := range reals {
		addr := lb.numToReals[num]
		lb.decreaseRefCountForReal(&addr)
	}

	vipNumKey := 0
	for key, num := range lb.vipNums {
		if num == vip.vipNum {
			vipNumKey = key
		}
	}
	lb.vipNums = append(lb.vipNums[0:vipNumKey], lb.vipNums[vipNumKey+1:]...)
	delete(lb.vips, *vipKey)

	return &rpc.Bool{Success: true}, nil
}

func (lb *OpenLb) GetVipFlags(ctx context.Context, vip *rpc.Vip) (*rpc.Flags, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetAllVips(ctx context.Context, empty *rpc.Empty) (*rpc.Vips, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) ModifyVip(ctx context.Context, meta *rpc.VipMeta) (*rpc.Bool, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetStatsForVip(ctx context.Context, vip *rpc.Vip) (*rpc.Stats, error) {
	//TODO implement me
	panic("implement me")
}
