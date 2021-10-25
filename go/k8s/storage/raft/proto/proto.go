package proto

import (
	"encoding/binary"
	"fmt"

	"k8s-lx1036/k8s/storage/raft/util"
)

type (
	MsgType        byte
	EntryType      byte
	ConfChangeType byte
	PeerType       byte
)

const (
	ReqMsgAppend MsgType = iota
	ReqMsgVote
	ReqMsgHeartBeat
	ReqMsgSnapShot
	ReqMsgElectAck
	RespMsgAppend
	RespMsgVote
	RespMsgHeartBeat
	RespMsgSnapShot
	RespMsgElectAck
	LocalMsgHup
	LocalMsgProp
	LeaseMsgOffline
	LeaseMsgTimeout
	ReqCheckQuorum
	RespCheckQuorum
)

func (t MsgType) String() string {
	switch t {
	case 0:
		return "ReqMsgAppend"
	case 1:
		return "ReqMsgVote"
	case 2:
		return "ReqMsgHeartBeat"
	case 3:
		return "ReqMsgSnapShot"
	case 4:
		return "ReqMsgElectAck"
	case 5:
		return "RespMsgAppend"
	case 6:
		return "RespMsgVote"
	case 7:
		return "RespMsgHeartBeat"
	case 8:
		return "RespMsgSnapShot"
	case 9:
		return "RespMsgElectAck"
	case 10:
		return "LocalMsgHup"
	case 11:
		return "LocalMsgProp"
	case 12:
		return "LeaseMsgOffline"
	case 13:
		return "LeaseMsgTimeout"
	case 14:
		return "ReqCheckQuorum"
	case 15:
		return "RespCheckQuorum"
	}
	return "unkown"
}

const (
	ConfAddNode    ConfChangeType = 0
	ConfRemoveNode ConfChangeType = 1
	ConfUpdateNode ConfChangeType = 2

	EntryNormal     EntryType = 0
	EntryConfChange EntryType = 1

	PeerNormal  PeerType = 0
	PeerArbiter PeerType = 1
)

// Entry is the replication log entry.
type Entry struct {
	Type  EntryType
	Term  uint64
	Index uint64
	Data  []byte
}

// Entry codec
func (e *Entry) Size() uint64 {
	return entry_header + uint64(len(e.Data))
}

func (e *Entry) Decode(datas []byte) {
	e.Type = EntryType(datas[0])
	e.Term = binary.BigEndian.Uint64(datas[1:])
	e.Index = binary.BigEndian.Uint64(datas[9:])
	if uint64(len(datas)) > entry_header {
		e.Data = datas[entry_header:]
	}
}

// Message is the transport message.
type Message struct {
	Type         MsgType
	ForceVote    bool
	Reject       bool
	RejectIndex  uint64
	ID           uint64
	From         uint64
	To           uint64
	Term         uint64
	LogTerm      uint64
	Index        uint64
	Commit       uint64
	SnapshotMeta SnapshotMeta
	Entries      []*Entry
	Context      []byte
	Snapshot     Snapshot // No need for codec
}

func (m *Message) IsHeartbeatMsg() bool {
	return m.Type == ReqMsgHeartBeat || m.Type == RespMsgHeartBeat
}

func (m *Message) Decode(r *util.BufferReader) error {
	var (
		datas []byte
		err   error
	)
	if datas, err = r.ReadFull(4); err != nil {
		return err
	}
	if datas, err = r.ReadFull(int(binary.BigEndian.Uint32(datas))); err != nil {
		return err
	}

	ver := datas[0]
	if ver == version1 {
		m.Type = MsgType(datas[1])
		m.ForceVote = (datas[2] == 1)
		m.Reject = (datas[3] == 1)
		m.RejectIndex = binary.BigEndian.Uint64(datas[4:])
		m.ID = binary.BigEndian.Uint64(datas[12:])
		m.From = binary.BigEndian.Uint64(datas[20:])
		m.To = binary.BigEndian.Uint64(datas[28:])
		m.Term = binary.BigEndian.Uint64(datas[36:])
		m.LogTerm = binary.BigEndian.Uint64(datas[44:])
		m.Index = binary.BigEndian.Uint64(datas[52:])
		m.Commit = binary.BigEndian.Uint64(datas[60:])
		if m.Type == ReqMsgSnapShot {
			m.SnapshotMeta.Decode(datas[message_header:])
		} else {
			size := binary.BigEndian.Uint32(datas[message_header:])
			start := message_header + 4
			if size > 0 {
				for i := uint32(0); i < size; i++ {
					esize := binary.BigEndian.Uint32(datas[start:])
					start = start + 4
					end := start + uint64(esize)
					entry := new(Entry)
					entry.Decode(datas[start:end])
					m.Entries = append(m.Entries, entry)
					start = end
				}
			}
			if start < uint64(len(datas)) {
				m.Context = datas[start:]
			}
		}
	}
	return nil
}

func (m *Message) ToString() (mesg string) {
	return fmt.Sprintf("Mesg:[%v] type(%v) ForceVote(%v) Reject(%v) RejectIndex(%v) "+
		"From(%v) To(%v) Term(%v) LogTrem(%v) Index(%v) Commit(%v)", m.ID, m.Type.String(), m.ForceVote,
		m.Reject, m.RejectIndex, m.From, m.To, m.Term, m.LogTerm, m.Index, m.Commit)
}

type SnapshotMeta struct {
	Index uint64
	Term  uint64
	Peers []Peer
}

func (m *SnapshotMeta) Decode(datas []byte) {
	m.Index = binary.BigEndian.Uint64(datas)
	m.Term = binary.BigEndian.Uint64(datas[8:])
	size := binary.BigEndian.Uint32(datas[16:])
	m.Peers = make([]Peer, size)
	start := snapmeta_header
	for i := uint32(0); i < size; i++ {
		m.Peers[i].Decode(datas[start:])
		start = start + peer_size
	}
}

// The Snapshot interface is supplied by the application to access the snapshot data of application.
type Snapshot interface {
	SnapIterator
	ApplyIndex() uint64
	Close()
}

type SnapIterator interface {
	// if error=io.EOF represent snapshot terminated.
	Next() ([]byte, error)
}

type Peer struct {
	Type     PeerType
	Priority uint16
	ID       uint64 // NodeID
	PeerID   uint64 // Replica ID, unique over all raft groups and all replicas in the same group
}

func (p *Peer) Encode(datas []byte) {
	datas[0] = byte(p.Type)
	binary.BigEndian.PutUint16(datas[1:], p.Priority)
	binary.BigEndian.PutUint64(datas[3:], p.ID)
}

func (p *Peer) Decode(datas []byte) {
	p.Type = PeerType(datas[0])
	p.Priority = binary.BigEndian.Uint16(datas[1:])
	p.ID = binary.BigEndian.Uint64(datas[3:])
}

// HardState is the repl state,must persist to the storage.
type HardState struct {
	Term   uint64
	Commit uint64
	Vote   uint64
}

type ConfChange struct {
	Type    ConfChangeType
	Peer    Peer
	Context []byte
}

func (c *ConfChange) Encode() []byte {
	datas := make([]byte, 1+peer_size+uint64(len(c.Context)))
	datas[0] = byte(c.Type)
	c.Peer.Encode(datas[1:])
	if len(c.Context) > 0 {
		copy(datas[peer_size+1:], c.Context)
	}

	return datas
}
