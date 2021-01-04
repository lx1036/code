package calico

import (
	"math/big"
	"net"

	"github.com/projectcalico/libcalico-go/lib/backend/model"
)

func OrdinalToIP(b *model.AllocationBlock, ord int) net.IP {
	ip := b.CIDR.IP
	var intVal *big.Int
	if ip.To4() != nil {
		intVal = big.NewInt(0).SetBytes(ip.To4())
	} else {
		intVal = big.NewInt(0).SetBytes(ip.To16())
	}
	sum := big.NewInt(0).Add(intVal, big.NewInt(int64(ord)))
	return net.IP(sum.Bytes())
}
