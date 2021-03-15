package proto

import (
	"encoding/binary"
	"sort"
)

const (
	version1        byte   = 1
	peer_size       uint64 = 11
	entry_header    uint64 = 17
	snapmeta_header uint64 = 20
	message_header  uint64 = 68
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
