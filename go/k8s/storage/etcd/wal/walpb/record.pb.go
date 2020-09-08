package walpb

import (
	proto "github.com/golang/protobuf/proto"

)

type Record struct {
	Type             int64  `protobuf:"varint,1,opt,name=type" json:"type"`
	Crc              uint32 `protobuf:"varint,2,opt,name=crc" json:"crc"`
	Data             []byte `protobuf:"bytes,3,opt,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}
func (m *Record) Reset()                    { *m = Record{} }
func (m *Record) String() string            { return proto.CompactTextString(m) }
func (*Record) ProtoMessage()               {}
func (*Record) Descriptor() ([]byte, []int) { return fileDescriptorRecord, []int{0} }

type Snapshot struct {
	Index            uint64 `protobuf:"varint,1,opt,name=index" json:"index"`
	Term             uint64 `protobuf:"varint,2,opt,name=term" json:"term"`
	XXX_unrecognized []byte `json:"-"`
}
func (m *Snapshot) Reset()                    { *m = Snapshot{} }
func (m *Snapshot) String() string            { return proto.CompactTextString(m) }
func (*Snapshot) ProtoMessage()               {}
func (*Snapshot) Descriptor() ([]byte, []int) { return fileDescriptorRecord, []int{1} }

var fileDescriptorRecord = []byte{
	// 186 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x29, 0x4a, 0x4d, 0xce,
	0x2f, 0x4a, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2d, 0x4f, 0xcc, 0x29, 0x48, 0x92,
	0x12, 0x49, 0xcf, 0x4f, 0xcf, 0x07, 0x8b, 0xe8, 0x83, 0x58, 0x10, 0x49, 0x25, 0x3f, 0x2e, 0xb6,
	0x20, 0xb0, 0x62, 0x21, 0x09, 0x2e, 0x96, 0x92, 0xca, 0x82, 0x54, 0x09, 0x46, 0x05, 0x46, 0x0d,
	0x66, 0x27, 0x96, 0x13, 0xf7, 0xe4, 0x19, 0x82, 0xc0, 0x22, 0x42, 0x62, 0x5c, 0xcc, 0xc9, 0x45,
	0xc9, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0xbc, 0x50, 0x09, 0x90, 0x80, 0x90, 0x10, 0x17, 0x4b, 0x4a,
	0x62, 0x49, 0xa2, 0x04, 0xb3, 0x02, 0xa3, 0x06, 0x4f, 0x10, 0x98, 0xad, 0xe4, 0xc0, 0xc5, 0x11,
	0x9c, 0x97, 0x58, 0x50, 0x9c, 0x91, 0x5f, 0x22, 0x24, 0xc5, 0xc5, 0x9a, 0x99, 0x97, 0x92, 0x5a,
	0x01, 0x36, 0x92, 0x05, 0xaa, 0x13, 0x22, 0x04, 0xb6, 0x2d, 0xb5, 0x28, 0x17, 0x6c, 0x28, 0x0b,
	0xdc, 0xb6, 0xd4, 0xa2, 0x5c, 0x27, 0x91, 0x13, 0x0f, 0xe5, 0x18, 0x4e, 0x3c, 0x92, 0x63, 0xbc,
	0xf0, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x19, 0x8f, 0xe5, 0x18, 0x00, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x7f, 0x5e, 0x5c, 0x46, 0xd3, 0x00, 0x00, 0x00,
}

func init() {
	proto.RegisterType((*Record)(nil), "walpb.Record")
	proto.RegisterType((*Snapshot)(nil), "walpb.Snapshot")
	proto.RegisterFile("record.proto", fileDescriptorRecord)
}


func sovRecord(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozRecord(x uint64) (n int) {
	return sovRecord(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func encodeVarintRecord(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Record) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovRecord(uint64(m.Type))
	n += 1 + sovRecord(uint64(m.Crc))
	if m.Data != nil {
		l = len(m.Data)
		n += 1 + l + sovRecord(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}
func (m *Record) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}
func (m *Record) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintRecord(dAtA, i, uint64(m.Type))
	dAtA[i] = 0x10
	i++
	i = encodeVarintRecord(dAtA, i, uint64(m.Crc))
	if m.Data != nil {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintRecord(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *Snapshot) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}
func (m *Snapshot) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovRecord(uint64(m.Index))
	n += 1 + sovRecord(uint64(m.Term))
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}
func (m *Snapshot) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintRecord(dAtA, i, uint64(m.Index))
	dAtA[i] = 0x10
	i++
	i = encodeVarintRecord(dAtA, i, uint64(m.Term))
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}
