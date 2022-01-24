package raft

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type MemoryTransport struct {
	sync.RWMutex

	localAddr ServerAddress
	timeout   time.Duration

	peers map[ServerAddress]*MemoryTransport

	consumerCh chan RPC
}

func NewMemoryTransport(addr ServerAddress) *MemoryTransport {
	return &MemoryTransport{
		localAddr:  addr,
		timeout:    1 * time.Second,
		peers:      make(map[ServerAddress]*MemoryTransport),
		consumerCh: make(chan RPC, 16),
	}
}

func (transport *MemoryTransport) Consumer() <-chan RPC {
	return transport.consumerCh
}

func (transport *MemoryTransport) LocalAddr() ServerAddress {
	return transport.localAddr
}

func (transport *MemoryTransport) AppendEntriesPipeline(id ServerID, target ServerAddress) (AppendPipeline, error) {
	panic("implement me")
}

func (transport *MemoryTransport) AppendEntries(id ServerID, target ServerAddress, request *AppendEntriesRequest, resp *AppendEntriesResponse) error {
	rpcResp, err := transport.sendRPC(target, request, nil, transport.timeout)
	if err != nil {
		return err
	}

	// Copy the result back
	out := rpcResp.Response.(*AppendEntriesResponse)
	*resp = *out
	return nil
}

func (transport *MemoryTransport) RequestVote(id ServerID, target ServerAddress, request *RequestVoteRequest, resp *RequestVoteResponse) error {
	rpcResp, err := transport.sendRPC(target, request, nil, transport.timeout)
	if err != nil {
		return err
	}

	// Copy the result back
	out := rpcResp.Response.(*RequestVoteResponse)
	*resp = *out
	return nil
}

func (transport *MemoryTransport) InstallSnapshot(id ServerID, target ServerAddress, request *InstallSnapshotRequest, resp *InstallSnapshotResponse, data io.Reader) error {
	panic("implement me")
}

func (transport *MemoryTransport) TimeoutNow(id ServerID, target ServerAddress, request *TimeoutNowRequest, resp *TimeoutNowResponse) error {
	panic("implement me")
}

// Connect is used to connect this transport to another transport for
// a given peer name. This allows for local routing.
func (transport *MemoryTransport) Connect(peer ServerAddress, t Transport) {
	trans := t.(*MemoryTransport)
	transport.Lock()
	defer transport.Unlock()
	transport.peers[peer] = trans
}

func (transport *MemoryTransport) Disconnect(peer ServerAddress) {
	panic("implement me")
}

func (transport *MemoryTransport) DisconnectAll() {
	panic("implement me")
}

func (transport *MemoryTransport) sendRPC(target ServerAddress, request interface{}, reader io.Reader, timeout time.Duration) (rpcResp RPCResponse, err error) {
	transport.RLock()
	trans, ok := transport.peers[target]
	transport.RUnlock()
	if !ok {
		err = fmt.Errorf("failed to connect to peer: %v", target)
		return
	}

	// Send the RPC
	respCh := make(chan RPCResponse, 1)
	req := RPC{
		Command:  request,
		Reader:   reader,
		RespChan: respCh,
	}
	select {
	case trans.consumerCh <- req:
	case <-time.After(timeout):
		err = fmt.Errorf("send timed out")
		return
	}

	// Wait for a response
	select {
	case rpcResp = <-respCh:
		if rpcResp.Error != nil {
			err = rpcResp.Error
		}
	case <-time.After(timeout):
		err = fmt.Errorf("command timed out")
	}
	return
}

func (transport *MemoryTransport) EncodePeer(id ServerID, addr ServerAddress) []byte {
	return []byte(addr)
}

func (transport *MemoryTransport) DecodePeer(addr []byte) ServerAddress {
	return ServerAddress(addr)
}

func (transport *MemoryTransport) SetHeartbeatHandler(cb func(rpc RPC)) {}

func (transport *MemoryTransport) Close() error {
	return nil
}
