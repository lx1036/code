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

	"github.com/hashicorp/go-msgpack/codec"
	"k8s.io/klog/v2"
)

const (
	// connReceiveBufferSize is the size of the buffer we will use for reading RPC requests into
	// on followers
	connReceiveBufferSize = 256 * 1024 // 256KB

	// connSendBufferSize is the size of the buffer we will use for sending RPC request data from
	// the leader to followers.
	connSendBufferSize = 256 * 1024 // 256KB
)

var (
	// ErrTransportShutdown is returned when operations on a transport are
	// invoked after it's been terminated.
	ErrTransportShutdown = errors.New("transport shutdown")

	// ErrPipelineShutdown is returned when the pipeline is closed.
	ErrPipelineShutdown = errors.New("append pipeline closed")
)

type netConn struct {
	target string
	conn   net.Conn
	writer *bufio.Writer
	dec    *codec.Decoder
	enc    *codec.Encoder
}

func (n *netConn) Release() error {
	return n.conn.Close()
}

type TCPTransport struct {
	connPoolLock sync.Mutex
	connPool     map[string][]*netConn
	maxPool      int
	timeout      time.Duration
	listener     *net.TCPListener

	heartbeatFn     func(RPC)
	heartbeatFnLock sync.Mutex

	consumerCh chan RPC

	shutdownCh chan struct{}
}

func NewTCPTransport(addr string) (*TCPTransport, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	transport := &TCPTransport{
		listener: listener.(*net.TCPListener),

		consumerCh: make(chan RPC),
		timeout:    time.Second * 3,
		maxPool:    2,

		shutdownCh: make(chan struct{}),
	}

	go transport.listen()

	return transport, nil
}

func (transport *TCPTransport) SetHeartbeatHandler(cb func(rpc RPC)) {
	transport.heartbeatFnLock.Lock()
	defer transport.heartbeatFnLock.Unlock()
	transport.heartbeatFn = cb
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
		go transport.handleConn(context.TODO(), conn)
	}
}

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
	/*case rpcRequestVote:
	var req RequestVoteRequest
	if err := dec.Decode(&req); err != nil {
		return err
	}
	rpc.Command = &req*/

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

func (transport *TCPTransport) AppendEntries(id string, target string, request *AppendEntriesRequest, resp *AppendEntriesResponse) error {
	return transport.genericRPC(id, target, rpcAppendEntries, request, resp)
}

// genericRPC handles a simple request/response RPC.
func (transport *TCPTransport) genericRPC(id string, target string, rpcType uint8, request interface{}, resp interface{}) error {
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
func (transport *TCPTransport) getConn(target string) (*netConn, error) {
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
func (transport *TCPTransport) getPooledConn(target string) *netConn {
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

// INFO: 判断 AppendEntriesRequest 是 heartbeat，而不是 log request
func isHeartbeatRequest(req *AppendEntriesRequest) bool {
	return req.Term != 0 && req.Leader != nil &&
		req.PrevLogIndex == 0 && req.PrevLogTerm == 0 &&
		len(req.Entries) == 0 && req.LeaderCommitIndex == 0
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
