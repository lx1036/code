package proto

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/pkg/buf"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"
)

var (
	Buffers = buf.NewBufferPool()
)

// Operations
const (
	ProtoMagic uint8 = 0xFF

	// Operations: Client -> MetaNode.
	OpMetaCreateInode      uint8 = 0x20
	OpMetaUnlinkInode      uint8 = 0x21
	OpMetaBatchUnlinkInode uint8 = 0x92
	OpMetaCreateDentry     uint8 = 0x22
	OpMetaDeleteDentry     uint8 = 0x23
	OpMetaOpen             uint8 = 0x24
	OpMetaLookup           uint8 = 0x25
	OpMetaReadDir          uint8 = 0x26
	OpMetaInodeGet         uint8 = 0x27
	OpMetaBatchInodeGet    uint8 = 0x28
	OpMetaUpdateDentry     uint8 = 0x2C
	OpMetaTruncate         uint8 = 0x2D
	OpMetaLinkInode        uint8 = 0x2E
	OpMetaEvictInode       uint8 = 0x2F
	OpMetaSetattr          uint8 = 0x30
	OpMetaReleaseOpen      uint8 = 0x31
	OpMetaLookupName       uint8 = 0x32

	//Operations: MetaNode Leader -> MetaNode Follower
	OpMetaFreeInodesOnRaftFollower uint8 = 0x3F

	// Operations: Master -> MetaNode
	OpCreateMetaPartition           uint8 = 0x40
	OpMetaNodeHeartbeat             uint8 = 0x41
	OpDeleteMetaPartition           uint8 = 0x42
	OpUpdateMetaPartition           uint8 = 0x43
	OpLoadMetaPartition             uint8 = 0x44
	OpDecommissionMetaPartition     uint8 = 0x45
	OpAddMetaPartitionRaftMember    uint8 = 0x46
	OpRemoveMetaPartitionRaftMember uint8 = 0x47
	OpMetaPartitionTryToLeader      uint8 = 0x48

	// Commons
	OpArgMismatchErr uint8 = 0xF1
	OpNotExistErr    uint8 = 0xF2
	OpDiskErr        uint8 = 0xF3
	OpErr            uint8 = 0xF4
	OpAgain          uint8 = 0xF5
	OpExistErr       uint8 = 0xF6
	OpInodeFullErr   uint8 = 0xF7
	OpTryOtherAddr   uint8 = 0xF8
	OpNotPerm        uint8 = 0xF9
	OpNotEmtpy       uint8 = 0xFA
	OpOk             uint8 = 0xFB
)

const (
	WriteDeadlineTime        = 15
	ReadDeadlineTime         = 15
	SyncSendTaskDeadlineTime = 20
	NoReadDeadlineTime       = -1
)

// Packet defines the packet structure.
type Packet struct {
	Magic       uint8
	Opcode      uint8
	ResultCode  uint8
	CRC         uint32
	Size        uint32
	ArgLen      uint32
	PartitionID uint64
	ReqID       int64
	Arg         []byte // for create or append ops, the data contains the address
	Data        []byte
	StartT      int64
}

func NewPacket() *Packet {
	return &Packet{
		Magic:  ProtoMagic,
		StartT: time.Now().UnixNano(),
	}
}

// MarshalData marshals the packet data.
func (p *Packet) MarshalData(v interface{}) error {
	data, err := json.Marshal(v)
	if err == nil {
		p.Data = data
		p.Size = uint32(len(p.Data))
	}
	return err
}

// WriteToConn writes through the given connection.
func (p *Packet) WriteToConn(c net.Conn) (err error) {
	c.SetWriteDeadline(time.Now().Add(WriteDeadlineTime * time.Second))
	header, err := Buffers.Get(util.PacketHeaderSize)
	if err != nil {
		header = make([]byte, util.PacketHeaderSize)
	}
	defer Buffers.Put(header)

	p.MarshalHeader(header)
	if _, err = c.Write(header); err == nil {
		if _, err = c.Write(p.Arg[:int(p.ArgLen)]); err == nil {
			if p.Data != nil && p.Size != 0 {
				_, err = c.Write(p.Data[:p.Size])
			}
		}
	}

	return
}

// MarshalHeader marshals the packet header.
func (p *Packet) MarshalHeader(out []byte) {
	out[0] = p.Magic
	out[1] = p.Opcode
	out[2] = p.ResultCode
	binary.BigEndian.PutUint32(out[3:7], p.CRC)
	binary.BigEndian.PutUint32(out[7:11], p.Size)
	binary.BigEndian.PutUint32(out[11:15], p.ArgLen)
	binary.BigEndian.PutUint64(out[15:23], p.PartitionID)
	binary.BigEndian.PutUint64(out[23:util.PacketHeaderSize], uint64(p.ReqID))
	return
}

// UnmarshalHeader unmarshals the packet header.
func (p *Packet) UnmarshalHeader(in []byte) error {
	p.Magic = in[0]
	if p.Magic != ProtoMagic {
		return errors.New("Bad Magic " + strconv.Itoa(int(p.Magic)))
	}

	p.Opcode = in[1]
	p.ResultCode = in[2]
	p.CRC = binary.BigEndian.Uint32(in[3:7])
	p.Size = binary.BigEndian.Uint32(in[7:11])
	p.ArgLen = binary.BigEndian.Uint32(in[11:15])
	p.PartitionID = binary.BigEndian.Uint64(in[15:23])
	p.ReqID = int64(binary.BigEndian.Uint64(in[23:util.PacketHeaderSize]))
	return nil
}

// ReadFromConn reads the data from the given connection.
func (p *Packet) ReadFromConn(c net.Conn, timeoutSec int) (err error) {
	if timeoutSec != NoReadDeadlineTime {
		c.SetReadDeadline(time.Now().Add(time.Second * time.Duration(timeoutSec)))
	} else {
		c.SetReadDeadline(time.Time{})
	}
	header, err := Buffers.Get(util.PacketHeaderSize)
	if err != nil {
		header = make([]byte, util.PacketHeaderSize)
	}
	defer Buffers.Put(header)
	if _, err = io.ReadFull(c, header); err != nil {
		return
	}
	if err = p.UnmarshalHeader(header); err != nil {
		return
	}

	if p.ArgLen > 0 {
		p.Arg = make([]byte, int(p.ArgLen))
		if _, err = io.ReadFull(c, p.Arg[:int(p.ArgLen)]); err != nil {
			return err
		}
	}

	if p.Size < 0 {
		return
	}
	size := p.Size
	p.Data = make([]byte, size)
	_, err = io.ReadFull(c, p.Data[:size])
	return err
}

// UnmarshalData unmarshals the packet data.
func (p *Packet) UnmarshalData(v interface{}) error {
	return json.Unmarshal(p.Data, v)
}

// GetResultMsg returns the result message.
func (p *Packet) GetResultMsg() (m string) {
	if p == nil {
		return ""
	}
	switch p.ResultCode {
	case OpDiskErr:
		m = "DiskErr"
	case OpErr:
		m = "Err"
	case OpAgain:
		m = "Again"
	case OpOk:
		m = "Ok"
	case OpExistErr:
		m = "ExistErr"
	case OpInodeFullErr:
		m = "InodeFullErr"
	case OpArgMismatchErr:
		m = "ArgUnmatchErr"
	case OpNotExistErr:
		m = "NotExistErr"
	case OpTryOtherAddr:
		m = "TryOtherAddr"
	case OpNotPerm:
		m = "NotPerm"
	case OpNotEmtpy:
		m = "DirNotEmpty"
	default:
		return fmt.Sprintf("Unknown ResultCode(%v)", p.ResultCode)
	}
	return
}

// GetOpMsg returns the operation type.
func (p *Packet) GetOpMsg() (m string) {
	switch p.Opcode {
	case OpMetaCreateInode:
		m = "OpMetaCreateInode"
	case OpMetaUnlinkInode:
		m = "OpMetaUnlinkInode"
	case OpMetaCreateDentry:
		m = "OpMetaCreateDentry"
	case OpMetaDeleteDentry:
		m = "OpMetaDeleteDentry"
	case OpMetaOpen:
		m = "OpMetaOpen"
	case OpMetaReleaseOpen:
		m = "OpMetaReleaseOpen"
	case OpMetaLookup:
		m = "OpMetaLookup"
	case OpMetaReadDir:
		m = "OpMetaReadDir"
	case OpMetaInodeGet:
		m = "OpMetaInodeGet"
	case OpMetaBatchInodeGet:
		m = "OpMetaBatchInodeGet"
	case OpMetaUpdateDentry:
		m = "OpMetaUpdateDentry"
	case OpMetaTruncate:
		m = "OpMetaTruncate"
	case OpMetaLinkInode:
		m = "OpMetaLinkInode"
	case OpMetaEvictInode:
		m = "OpMetaEvictInode"
	case OpMetaSetattr:
		m = "OpMetaSetattr"
	case OpCreateMetaPartition:
		m = "OpCreateMetaPartition"
	case OpMetaNodeHeartbeat:
		m = "OpMetaNodeHeartbeat"
	case OpDeleteMetaPartition:
		m = "OpDeleteMetaPartition"
	case OpUpdateMetaPartition:
		m = "OpUpdateMetaPartition"
	case OpLoadMetaPartition:
		m = "OpLoadMetaPartition"
	case OpDecommissionMetaPartition:
		m = "OpDecommissionMetaPartition"
	case OpAddMetaPartitionRaftMember:
		m = "OpAddMetaPartitionRaftMember"
	case OpRemoveMetaPartitionRaftMember:
		m = "OpRemoveMetaPartitionRaftMember"
	case OpMetaPartitionTryToLeader:
		m = "OpMetaPartitionTryToLeader"
	case OpMetaLookupName:
		m = "OpMetaLookupName"
	}
	return
}

// GetUniqueLogId returns the unique log ID.
func (p *Packet) GetUniqueLogId() (m string) {
	defer func() {
		m = m + fmt.Sprintf("_ResultMesg(%v)", p.GetResultMsg())
	}()
	m = fmt.Sprintf("Req(%v)_Partition(%v)_Size(%v)_Opcode(%v)_CRC(%v)",
		p.ReqID, p.PartitionID, p.Size, p.GetOpMsg(), p.CRC)

	return
}

func (p *Packet) GetReqID() int64 {
	return p.ReqID
}

// PacketErrorWithBody sets the packet with error code whose body is filled with the given data.
func (p *Packet) PacketErrorWithBody(code uint8, reply []byte) {
	p.Size = uint32(len(reply))
	p.Data = make([]byte, p.Size)
	copy(p.Data[:p.Size], reply)
	p.ResultCode = code
	p.ArgLen = 0
}

var (
	GRequestID = int64(1)
)

// GenerateRequestID generates the request ID.
func GenerateRequestID() int64 {
	return atomic.AddInt64(&GRequestID, 1)
}

// NewPacketReqID returns a new packet with ReqID assigned.
func NewPacketReqID() *Packet {
	p := NewPacket()
	p.ReqID = GenerateRequestID()
	return p
}
