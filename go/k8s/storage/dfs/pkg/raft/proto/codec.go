package proto

import (
	"encoding/binary"
	"sort"
)

func EncodeHBConext(ctx HeartbeatContext) (buf []byte) {
	sort.Slice(ctx, func(i, j int) bool {
		return ctx[i] < ctx[j]
	})

	scratch := make([]byte, binary.MaxVarintLen64)
	prev := uint64(0)
	for _, id := range ctx {
		n := binary.PutUvarint(scratch, id-prev)
		buf = append(buf, scratch[:n]...)
		prev = id
	}
	return
}
