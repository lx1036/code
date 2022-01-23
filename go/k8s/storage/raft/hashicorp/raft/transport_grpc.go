package raft

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"net"
	"sync"

	pb "k8s-lx1036/k8s/storage/raft/hashicorp/raft/rpc"

	"k8s.io/klog/v2"
)

// INFO: @see https://github.com/Jille/raft-grpc-transport

type grpcNetConn struct {
	clientConn *grpc.ClientConn
	client     pb.TransportClient
}

type grpcServer struct {
	heartbeatFnLock sync.Mutex
	heartbeatFn     func(RPC)

	consumerCh chan RPC

	shutdownCh chan struct{}
}

func (server *grpcServer) handleCommand(command interface{}, data io.Reader) (interface{}, error) {
	respCh := make(chan RPCResponse, 1)
	rpc := RPC{
		Command:  command,
		RespChan: respCh,
		Reader:   data,
	}

	// INFO: Check for heartbeat fast-path, skip Dispatch the RPC to consumerCh
	if value, ok := command.(*AppendEntriesRequest); ok && isHeartbeatRequest(value) {
		server.heartbeatFnLock.Lock()
		fn := server.heartbeatFn
		server.heartbeatFnLock.Unlock()
		if fn != nil {
			fn(rpc)
			goto RESP
		}
	}

	// Dispatch the RPC to handle the request -> response
	server.consumerCh <- rpc

	// Wait for response
RESP:
	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Response, nil

	case <-server.shutdownCh:
		return nil, ErrTransportShutdown
	}
}

func (server *grpcServer) AppendEntriesPipeline(server2 pb.Transport_AppendEntriesPipelineServer) error {
	panic("implement me")
}

func (server *grpcServer) AppendEntries(ctx context.Context, request *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	server.handleCommand(decodeAppendEntriesRequest(request), nil)

}

func (server *grpcServer) RequestVote(ctx context.Context, request *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	panic("implement me")
}

type GrpcTransport struct {
	localAddr   string
	dialOptions []grpc.DialOption
	listener    *net.TCPListener
	grpcServer  *grpc.Server

	consumerCh chan RPC

	server *grpcServer

	shutdownCh chan struct{}
}

func NewGrpcTransport(localAddr string, dialOptions []grpc.DialOption) (*GrpcTransport, error) {
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, err
	}

	transport := &GrpcTransport{
		localAddr:   localAddr,
		dialOptions: dialOptions,
		listener:    listener.(*net.TCPListener),
		grpcServer:  grpc.NewServer(),

		consumerCh: make(chan RPC),

		shutdownCh: make(chan struct{}, 1),
	}

	transport.server = &grpcServer{
		consumerCh: transport.consumerCh,
	}

	pb.RegisterTransportServer(transport.grpcServer, transport.server)
	go transport.listen()
}

func (transport *GrpcTransport) listen() {
	defer transport.grpcServer.Stop()
	err := transport.grpcServer.Serve(transport.listener)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[Run]serve grpcServer err: %v", err))
		transport.shutdownCh <- struct{}{}
	}
}

func (transport *GrpcTransport) Consumer() <-chan RPC {

}

func (transport *GrpcTransport) AppendEntries(id ServerID, target ServerAddress, request *AppendEntriesRequest,
	resp *AppendEntriesResponse) error {

	conn, err := transport.getConn(target)
	if err != nil {
		return err
	}

	// Send the RPC
	response, err := conn.client.AppendEntries(context.TODO(), encodeAppendEntriesRequest(request))

	*resp = AppendEntriesResponse{
		Term:           response.Term,
		LastLog:        response.LastLog,
		Success:        response.Success,
		NoRetryBackoff: response.NoRetryBackoff,
	}

	return nil
}

func (transport *GrpcTransport) genericRPC(target ServerAddress) error {

}

func (transport *GrpcTransport) getConn(target ServerAddress) (*grpcNetConn, error) {
	// Check for a pooled conn
	if conn := transport.getPooledConn(target); conn != nil {
		return conn, nil
	}

	// Dial a new connection
	clientConn, err := grpc.Dial(string(target), transport.dialOptions...)
	if err != nil {

	}

	conn := &grpcNetConn{
		clientConn: clientConn,
		client:     pb.NewTransportClient(clientConn),
	}

	return conn, nil
}

func encodeAppendEntriesRequest(request *AppendEntriesRequest) *pb.AppendEntriesRequest {
	return &pb.AppendEntriesRequest{
		Term:              request.Term,
		Leader:            request.Leader,
		PrevLogIndex:      request.PrevLogIndex,
		PrevLogTerm:       request.PrevLogTerm,
		Entries:           encodeLogs(request.Entries),
		LeaderCommitIndex: request.LeaderCommitIndex,
	}
}

func encodeLogs(logs []*Log) []*pb.Log {

}

func decodeAppendEntriesRequest(request *pb.AppendEntriesRequest) *AppendEntriesRequest {
	return &AppendEntriesRequest{
		Term:              request.Term,
		Leader:            request.Leader,
		PrevLogIndex:      request.PrevLogIndex,
		PrevLogTerm:       request.PrevLogTerm,
		Entries:           decodeLogs(request.Entries),
		LeaderCommitIndex: request.LeaderCommitIndex,
	}
}

func decodeLogs(logs []*pb.Log) []*Log {

}
