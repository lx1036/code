package raft

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	pb "k8s-lx1036/k8s/storage/raft/hashicorp/raft/rpc"

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

// WithClose is an interface that a transport may provide which
// allows a transport to be shut down cleanly when a Raft instance
// shuts down.
//
// It is defined separately from Transport as unfortunately it wasn't in the
// original interface specification.
type WithClose interface {
	// Close permanently closes a transport, stopping
	// any associated goroutines and freeing other resources.
	Close() error
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

// StreamLayer is used with the TCPTransport to provide
// the low level stream abstraction.
type StreamLayer interface {
	net.Listener

	// Dial is used to create a new outgoing connection
	Dial(address ServerAddress, timeout time.Duration) (net.Conn, error)
}

const (
	// DefaultTimeoutScale is the default TimeoutScale in a TCPTransport.
	DefaultTimeoutScale = 256 * 1024 // 256KB
)

/*

TCPTransport provides a network based transport that can be
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
type TCPTransport struct {
	connPoolLock sync.Mutex
	connPool     map[ServerAddress][]*netConn
	maxPool      int
	listener     *net.TCPListener
	timeout      time.Duration
	TimeoutScale int

	consumerCh      chan RPC
	heartbeatFn     func(RPC)
	heartbeatFnLock sync.Mutex

	// streamCtx is used to cancel existing connection handlers.
	streamCtx     context.Context
	streamCancel  context.CancelFunc
	streamCtxLock sync.RWMutex

	shutdownLock sync.Mutex
	shutdown     bool
	shutdownCh   chan struct{}
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

func NewTCPTransport(bindAddr string, maxPool int, timeout time.Duration) (*TCPTransport, error) {
	listener, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	transport := &TCPTransport{
		connPool: make(map[ServerAddress][]*netConn),
		maxPool:  maxPool,
		listener: listener.(*net.TCPListener),
		timeout:  timeout,

		consumerCh: make(chan RPC),

		TimeoutScale: DefaultTimeoutScale,

		shutdownCh: make(chan struct{}),
	}

	// Create the connection context and then start our listener.
	transport.setupStreamContext()
	go transport.listen()

	return transport, nil
}

// setupStreamContext is used to create a new stream context. This should be
// called with the stream lock held.
func (transport *TCPTransport) setupStreamContext() {
	ctx, cancel := context.WithCancel(context.Background())
	transport.streamCtx = ctx
	transport.streamCancel = cancel
}

// getStreamContext is used retrieve the current stream context.
func (transport *TCPTransport) getStreamContext() context.Context {
	transport.streamCtxLock.RLock()
	defer transport.streamCtxLock.RUnlock()
	return transport.streamCtx
}

// Consumer implements the Transport interface.
func (transport *TCPTransport) Consumer() <-chan RPC {
	return transport.consumerCh
}

func (transport *TCPTransport) listen() {
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

		loopDelay = 0
		go transport.handleConn(transport.getStreamContext(), conn)
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

// handleConn is used to handle an inbound connection for its lifespan. The
// handler will exit when the passed context is cancelled or the connection is
// closed.
func (transport *TCPTransport) handleConn(connCtx context.Context, conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReaderSize(conn, connReceiveBufferSize)
	w := bufio.NewWriter(conn)
	dec := codec.NewDecoder(r, &codec.MsgpackHandle{})
	enc := codec.NewEncoder(w, &codec.MsgpackHandle{})

	for {
		select {
		case <-connCtx.Done():
			klog.Infof("stream layer is closed")
			return
		case <-transport.shutdownCh:
			return
		default:
		}

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
func (transport *TCPTransport) handleCommand(r *bufio.Reader, dec *codec.Decoder, enc *codec.Encoder) error {
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

	isHeartbeat := false
	switch rpcType {
	case rpcRequestVote:
		var req pb.RequestVoteRequest
		if err := dec.Decode(&req); err != nil {
			return err
		}
		rpc.Command = &req

	case rpcAppendEntries:
		var req AppendEntriesRequest
		if err := dec.Decode(&req); err != nil {
			return err
		}
		rpc.Command = &req

		// Check if this is a heartbeat
		if isHeartbeatRequest(&req) {
			isHeartbeat = true
		}

	default:
		return fmt.Errorf("unknown rpc type %d", rpcType)
	}

	// INFO: Check for heartbeat fast-path, skip Dispatch the RPC to consumerCh
	if isHeartbeat {
		transport.heartbeatFnLock.Lock()
		fn := transport.heartbeatFn
		transport.heartbeatFnLock.Unlock()
		if fn != nil {
			fn(rpc)
			goto RESP
		}
	}

	// Dispatch the RPC to handle the request -> response
	select {
	case transport.consumerCh <- rpc:
	case <-transport.shutdownCh:
		return ErrTransportShutdown
	}

	// Wait for response
RESP:
	select {
	case resp := <-respCh:
		// Send the error first
		respErr := ""
		if resp.Error != nil {
			respErr = resp.Error.Error()
		}
		if err := enc.Encode(respErr); err != nil {
			return err
		}

		// Send the response
		if err := enc.Encode(resp.Response); err != nil {
			return err
		}

	case <-transport.shutdownCh:
		return ErrTransportShutdown
	}

	return nil
}

// SetHeartbeatHandler is used to setup a heartbeat handler
// as a fast-pass. This is to avoid head-of-line blocking from
// disk IO.
func (transport *TCPTransport) SetHeartbeatHandler(cb func(rpc RPC)) {
	transport.heartbeatFnLock.Lock()
	defer transport.heartbeatFnLock.Unlock()
	transport.heartbeatFn = cb
}

// RequestVote implements the Transport interface.
func (transport *TCPTransport) RequestVote(id ServerID, target ServerAddress, request *pb.RequestVoteRequest, resp *RequestVoteResponse) error {
	return transport.genericRPC(id, target, rpcRequestVote, request, resp)
}

func (transport *TCPTransport) AppendEntries(id ServerID, target ServerAddress, request *AppendEntriesRequest, resp *AppendEntriesResponse) error {
	return transport.genericRPC(id, target, rpcAppendEntries, request, resp)
}

// genericRPC handles a simple request/response RPC.
func (transport *TCPTransport) genericRPC(id ServerID, target ServerAddress, rpcType uint8, request interface{}, resp interface{}) error {
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

	// Wait for and Decode the response
	canReturn, err := decodeResponse(conn, resp)
	if canReturn {
		transport.addConn(conn)
	}

	return err
}

// getConn is used to get a connection from the pool.
func (transport *TCPTransport) getConn(target ServerAddress) (*netConn, error) {
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
func (transport *TCPTransport) getPooledConn(target ServerAddress) *netConn {
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
func (transport *TCPTransport) addConn(conn *netConn) {
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

// EncodePeer implements the Transport interface.
func (transport *TCPTransport) EncodePeer(id ServerID, address ServerAddress) []byte {
	return []byte(address)
}

// DecodePeer implements the Transport interface.
func (transport *TCPTransport) DecodePeer(buf []byte) ServerAddress {
	return ServerAddress(buf)
}

func (transport *TCPTransport) Close() error {
	transport.shutdownLock.Lock()
	defer transport.shutdownLock.Unlock()

	if !transport.shutdown {
		close(transport.shutdownCh)
		transport.shutdown = true
	}
	return nil
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
	// INFO: @see handleCommand()
	// Decode the error if any
	var rpcError string
	if err := conn.dec.Decode(&rpcError); err != nil {
		conn.Release()
		return false, err
	}

	// Decode the response
	if err := conn.dec.Decode(resp); err != nil {
		conn.Release()
		return false, err
	}

	// Format an error if any
	if rpcError != "" {
		return true, fmt.Errorf(rpcError)
	}
	return true, nil
}

// INFO: 判断 AppendEntriesRequest 是 heartbeat，而不是 log request
func isHeartbeatRequest(req *AppendEntriesRequest) bool {
	return req.Term != 0 && req.Leader != nil &&
		req.PrevLogIndex == 0 && req.PrevLogTerm == 0 &&
		len(req.Entries) == 0 && req.LeaderCommitIndex == 0
}
