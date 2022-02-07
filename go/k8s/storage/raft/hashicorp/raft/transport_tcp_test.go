package raft

import (
	"bufio"
	"fmt"
	"io"
	pb "k8s-lx1036/k8s/storage/raft/hashicorp/raft/rpc"
	"net"
	"testing"
	"time"

	"github.com/hashicorp/go-msgpack/codec"
	"k8s.io/klog/v2"
)

func TestRpcCodec(test *testing.T) {
	stopCh := make(chan struct{})
	bindAddr := "127.0.0.1:7000"
	go func() {
		listener, err := net.Listen("tcp", bindAddr)
		if err != nil {
			klog.Error(err)
			return
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				klog.Error(err)
				return
			}

			klog.Info("listener.Accept()")

			go func() {
				defer conn.Close()
				reader := bufio.NewReaderSize(conn, connReceiveBufferSize)
				writer := bufio.NewWriter(conn)
				dec := codec.NewDecoder(reader, &codec.MsgpackHandle{})
				enc := codec.NewEncoder(writer, &codec.MsgpackHandle{})
				for {
					rpcType, err := reader.ReadByte()
					if err != nil {
						klog.Error(err)
						return
					}
					var req pb.RequestVoteRequest
					switch rpcType {
					case rpcRequestVote:
						if err = dec.Decode(&req); err != nil {
							klog.Error(err)
							return
						}
					default:
						klog.Errorf(fmt.Sprintf("rpc type is unknown"))
						return
					}

					klog.Infof(fmt.Sprintf("accept request from rpc client, request:%+v", req))

					resp := &pb.RequestVoteResponse{
						Term:    req.Term,
						Granted: false,
					}
					if req.LastLogIndex > 10 && req.LastLogTerm > 0 {
						resp.Granted = true
					}

					// Send the response
					if err := enc.Encode(resp); err != nil {
						if err != io.EOF {
							klog.Error(err)
						}
						klog.Infof(fmt.Sprintf("send the response"))
						return
					}
					if err = writer.Flush(); err != nil {
						klog.Error(err)
						return
					}

					klog.Infof(fmt.Sprintf("send response to rpc client, response:%+v", *resp))
				}
			}()
		}
	}()

	go func() {
		time.Sleep(time.Second * 3)

		var err error
		conn, err := net.Dial("tcp", bindAddr)
		if err != nil {
			klog.Error(err)
			return
		}
		defer func() {
			if err != nil {
				klog.Error(err)
				conn.Close()
			}
		}()

		reader := bufio.NewReaderSize(conn, connReceiveBufferSize)
		writer := bufio.NewWriterSize(conn, connSendBufferSize)
		enc := codec.NewEncoder(writer, &codec.MsgpackHandle{})
		dec := codec.NewDecoder(reader, &codec.MsgpackHandle{})

		// Send the request
		if err = writer.WriteByte(rpcRequestVote); err != nil {
			return
		}
		req := &pb.RequestVoteRequest{
			Term:         1,
			LastLogIndex: 11,
			LastLogTerm:  1,
		}
		if err = enc.Encode(req); err != nil {
			return
		}
		if err = writer.Flush(); err != nil {
			return
		}

		klog.Infof(fmt.Sprintf("send the request to rpc server, request:%+v", *req))

		// Decode the response
		var resp pb.RequestVoteResponse
		if err = dec.Decode(&resp); err != nil {
			return
		}

		klog.Infof(fmt.Sprintf("accept the response from rpc server, response:%+v", resp))

		if resp.Granted {
			klog.Infof(fmt.Sprintf("RequestVote is granted at term:%d", resp.Term))
		}
	}()

	<-stopCh
}
