package proto

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

const (
	ConfAddNode    ConfChangeType = 0
	ConfRemoveNode ConfChangeType = 1
	ConfUpdateNode ConfChangeType = 2

	EntryNormal     EntryType = 0
	EntryConfChange EntryType = 1

	PeerNormal  PeerType = 0
	PeerArbiter PeerType = 1
)

// Entry is the repl log entry.
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

type SnapshotMeta struct {
	Index uint64
	Term  uint64
	Peers []Peer
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

type HeartbeatContext []uint64

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
