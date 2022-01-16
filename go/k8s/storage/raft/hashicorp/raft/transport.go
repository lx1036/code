package raft

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/go-msgpack/codec"
	"k8s.io/klog/v2"
)

var (
	// ErrTransportShutdown is returned when operations on a transport are
	// invoked after it's been terminated.
	ErrTransportShutdown = errors.New("transport shutdown")

	// ErrPipelineShutdown is returned when the pipeline is closed.
	ErrPipelineShutdown = errors.New("append pipeline closed")
)

// RPCResponse captures both a response and a potential error.
type RPCResponse struct {
	Response interface{}
	Error    error
}

// RPC has a command, and provides a response mechanism.
type RPC struct {
	Command  interface{}
	Reader   io.Reader // Set only for InstallSnapshot
	RespChan chan<- RPCResponse
}

// Respond is used to respond with a response, error or both
func (r *RPC) Respond(resp interface{}, err error) {
	r.RespChan <- RPCResponse{resp, err}
}

// Transport provides an interface for network transports
// to allow Raft to communicate with other nodes.
type Transport interface {
	// Consumer returns a channel that can be used to
	// consume and respond to RPC requests.
	Consumer() <-chan RPC

	// LocalAddr is used to return our local address to distinguish from our peers.
	LocalAddr() ServerAddress

	// AppendEntriesPipeline returns an interface that can be used to pipeline
	// AppendEntries requests.
	AppendEntriesPipeline(id ServerID, target ServerAddress) (AppendPipeline, error)

	// AppendEntries sends the appropriate RPC to the target node.
	AppendEntries(id ServerID, target ServerAddress, args *AppendEntriesRequest, resp *AppendEntriesResponse) error

	// RequestVote sends the appropriate RPC to the target node.
	RequestVote(id ServerID, target ServerAddress, args *RequestVoteRequest, resp *RequestVoteResponse) error

	// InstallSnapshot is used to push a snapshot down to a follower. The data is read from
	// the ReadCloser and streamed to the client.
	InstallSnapshot(id ServerID, target ServerAddress, args *InstallSnapshotRequest, resp *InstallSnapshotResponse, data io.Reader) error

	// EncodePeer is used to serialize a peer's address.
	EncodePeer(id ServerID, addr ServerAddress) []byte

	// DecodePeer is used to deserialize a peer's address.
	DecodePeer([]byte) ServerAddress

	// SetHeartbeatHandler is used to setup a heartbeat handler
	// as a fast-pass. This is to avoid head-of-line blocking from
	// disk IO. If a Transport does not support this, it can simply
	// ignore the call, and push the heartbeat onto the Consumer channel.
	SetHeartbeatHandler(cb func(rpc RPC))

	// TimeoutNow is used to start a leadership transfer to the target node.
	TimeoutNow(id ServerID, target ServerAddress, args *TimeoutNowRequest, resp *TimeoutNowResponse) error
}

// AppendPipeline is used for pipelining AppendEntries requests. It is used
// to increase the replication throughput by masking latency and better
// utilizing bandwidth.
type AppendPipeline interface {
	// AppendEntries is used to add another request to the pipeline.
	// The send may block which is an effective form of back-pressure.
	AppendEntries(args *AppendEntriesRequest, resp *AppendEntriesResponse) (AppendFuture, error)

	// Consumer returns a channel that can be used to consume
	// response futures when they are ready.
	Consumer() <-chan AppendFuture

	// Close closes the pipeline and cancels all inflight RPCs
	Close() error
}

// AppendFuture is used to return information about a pipelined AppendEntries request.
type AppendFuture interface {
	Future

	// Start returns the time that the append request was started.
	// It is always OK to call this method.
	Start() time.Time

	// Request holds the parameters of the AppendEntries call.
	// It is always OK to call this method.
	Request() *AppendEntriesRequest

	// Response holds the results of the AppendEntries call.
	// This method must only be called after the Error
	// method returns, and will only be valid on success.
	Response() *AppendEntriesResponse
}

// StreamLayer is used with the NetworkTransport to provide
// the low level stream abstraction.
type StreamLayer interface {
	net.Listener

	// Dial is used to create a new outgoing connection
	Dial(address ServerAddress, timeout time.Duration) (net.Conn, error)
}

const (
	// DefaultTimeoutScale is the default TimeoutScale in a NetworkTransport.
	DefaultTimeoutScale = 256 * 1024 // 256KB
)

/*

NetworkTransport provides a network based transport that can be
used to communicate with Raft on remote machines. It requires
an underlying stream layer to provide a stream abstraction, which can
be simple TCP, TLS, etc.

This transport is very simple and lightweight. Each RPC request is
framed by sending a byte that indicates the message type, followed
by the MsgPack encoded request.

The response is an error string followed by the response object,
both are encoded using MsgPack.

InstallSnapshot is special, in that after the RPC request we stream
the entire state. That socket is not re-used as the connection state
is not known if there is an error.

*/
type NetworkTransport struct {
	connPoolLock sync.Mutex
	connPool     map[ServerAddress][]*netConn
	maxPool      int
	listener     *net.TCPListener
	timeout      time.Duration
	TimeoutScale int

	consumeCh chan RPC

	shutdown   bool
	shutdownCh chan struct{}
}

type netConn struct {
	target ServerAddress
	conn   net.Conn
	writer *bufio.Writer
	dec    *codec.Decoder
	enc    *codec.Encoder
}

func (n *netConn) Release() error {
	return n.conn.Close()
}

func NewNetworkTransport(bindAddr string, maxPool int, timeout time.Duration) (*NetworkTransport, error) {

	listener, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	transport := &NetworkTransport{
		connPool: make(map[ServerAddress][]*netConn),
		maxPool:  maxPool,
		listener: listener.(*net.TCPListener),
		timeout:  timeout,

		consumeCh: make(chan RPC),

		TimeoutScale: DefaultTimeoutScale,

		shutdownCh: make(chan struct{}),
	}

	go transport.listen()

	return transport, nil
}

// Consumer implements the Transport interface.
func (transport *NetworkTransport) Consumer() <-chan RPC {
	return transport.consumeCh
}

func (transport *NetworkTransport) listen() {
	const baseDelay = 5 * time.Millisecond
	const maxDelay = 1 * time.Second

	var loopDelay time.Duration
	for {
		conn, err := transport.listener.Accept()
		if err != nil {
			if loopDelay == 0 {
				loopDelay = baseDelay
			} else {
				loopDelay *= 2
			}
			if loopDelay > maxDelay {
				loopDelay = maxDelay
			}

			select {
			case <-transport.shutdownCh:
				return
			case <-time.After(loopDelay):
				klog.Errorf(fmt.Sprintf("failed to accept connection err:%v", err))
				continue
			}
		}

		go transport.handleConn(conn)
	}
}

const (
	// connReceiveBufferSize is the size of the buffer we will use for reading RPC requests into
	// on followers
	connReceiveBufferSize = 256 * 1024 // 256KB

	// connSendBufferSize is the size of the buffer we will use for sending RPC request data from
	// the leader to followers.
	connSendBufferSize = 256 * 1024 // 256KB
)

func (transport *NetworkTransport) handleConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReaderSize(conn, connReceiveBufferSize)
	w := bufio.NewWriter(conn)
	dec := codec.NewDecoder(r, &codec.MsgpackHandle{})
	enc := codec.NewEncoder(w, &codec.MsgpackHandle{})

	for {
		if err := transport.handleCommand(r, dec, enc); err != nil {
			if err != io.EOF {
				klog.Errorf(fmt.Sprintf("failed to decode incoming command err:%v", err))
			}
			return
		}
		if err := w.Flush(); err != nil {
			klog.Errorf(fmt.Sprintf("failed to flush response err:%v", err))
			return
		}
	}
}

const (
	rpcAppendEntries uint8 = iota
	rpcRequestVote
	rpcInstallSnapshot
	rpcTimeoutNow
)

// handleCommand is used to decode and dispatch a single command.
func (transport *NetworkTransport) handleCommand(r *bufio.Reader, dec *codec.Decoder, enc *codec.Encoder) error {

	// Get the rpc type
	rpcType, err := r.ReadByte()
	if err != nil {
		return err
	}

	// Create the RPC object
	respCh := make(chan RPCResponse, 1)
	rpc := RPC{
		RespChan: respCh,
	}

	switch rpcType {
	case rpcRequestVote:
		var req RequestVoteRequest
		if err := dec.Decode(&req); err != nil {
			return err
		}
		rpc.Command = &req

	default:
		return fmt.Errorf("unknown rpc type %d", rpcType)
	}

	// Dispatch the RPC
	select {
	case transport.consumeCh <- rpc:
	case <-transport.shutdownCh:
		return ErrTransportShutdown
	}

	select {
	case resp := <-respCh:
		klog.Info(resp)
	case <-transport.shutdownCh:
		return ErrTransportShutdown
	}

	return nil
}

// RequestVote implements the Transport interface.
func (transport *NetworkTransport) RequestVote(id ServerID, target ServerAddress, request *RequestVoteRequest, resp *RequestVoteResponse) error {
	return transport.genericRPC(id, target, rpcRequestVote, request, resp)
}

// genericRPC handles a simple request/response RPC.
func (transport *NetworkTransport) genericRPC(id ServerID, target ServerAddress, rpcType uint8, request interface{}, resp interface{}) error {
	conn, err := transport.getConn(target)
	if err != nil {
		return err
	}

	// Set a deadline
	if transport.timeout > 0 {
		conn.conn.SetDeadline(time.Now().Add(transport.timeout))
	}

	// Send the RPC
	if err = sendRPC(conn, rpcType, request); err != nil {
		return err
	}

	// Decode the response
	canReturn, err := decodeResponse(conn, resp)
	if canReturn {
		transport.addConn(conn)
	}

	return err
}

// getConn is used to get a connection from the pool.
func (transport *NetworkTransport) getConn(target ServerAddress) (*netConn, error) {
	// Check for a pooled conn
	if conn := transport.getPooledConn(target); conn != nil {
		return conn, nil
	}

	// Dial a new connection
	conn, err := net.DialTimeout("tcp", string(target), transport.timeout)
	if err != nil {
		return nil, err
	}
	netConn := &netConn{
		target: target,
		conn:   conn,
		dec:    codec.NewDecoder(bufio.NewReader(conn), &codec.MsgpackHandle{}),
		writer: bufio.NewWriterSize(conn, connSendBufferSize),
	}

	netConn.enc = codec.NewEncoder(netConn.writer, &codec.MsgpackHandle{})

	return netConn, nil
}

// getPooledConn is used to grab a pooled connection.
func (transport *NetworkTransport) getPooledConn(target ServerAddress) *netConn {
	transport.connPoolLock.Lock()
	defer transport.connPoolLock.Unlock()

	conns, ok := transport.connPool[target]
	if !ok || len(conns) == 0 {
		return nil
	}

	var conn *netConn
	num := len(conns)
	conn, conns[num-1] = conns[num-1], nil
	transport.connPool[target] = conns[:num-1]
	return conn
}

// returnConn returns a connection back to the pool.
func (transport *NetworkTransport) addConn(conn *netConn) {
	transport.connPoolLock.Lock()
	defer transport.connPoolLock.Unlock()

	key := conn.target
	conns, _ := transport.connPool[key]

	if len(conns) < transport.maxPool {
		transport.connPool[key] = append(conns, conn)
	} else {
		conn.Release()
	}
}

// sendRPC is used to encode and send the RPC.
func sendRPC(conn *netConn, rpcType uint8, request interface{}) error {
	// Write the request type
	if err := conn.writer.WriteByte(rpcType); err != nil {
		conn.Release()
		return err
	}

	// Send the request
	if err := conn.enc.Encode(request); err != nil {
		conn.Release()
		return err
	}

	// Flush
	if err := conn.writer.Flush(); err != nil {
		conn.Release()
		return err
	}
	return nil
}

// decodeResponse is used to decode an RPC response and reports whether
// the connection can be reused.
func decodeResponse(conn *netConn, resp interface{}) (bool, error) {
	return true, nil
}
