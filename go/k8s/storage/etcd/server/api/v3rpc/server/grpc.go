package server

import (
	"context"
	"math"

	"k8s-lx1036/k8s/storage/etcd/storage/mvcc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"k8s.io/klog/v2"
)

const (
	grpcOverheadBytes = 512 * 1024
	maxStreams        = math.MaxUint32
	maxSendBytes      = math.MaxInt32
)

// Server INFO: @see https://github.com/etcd-io/etcd/blob/main/server/etcdserver/api/v3rpc/grpc.go#L39-L93
func Server(watchableStore mvcc.WatchableKV) *grpc.Server {
	var opts []grpc.ServerOption
	chainUnaryInterceptors := []grpc.UnaryServerInterceptor{
		logGRPC,
		grpc_prometheus.UnaryServerInterceptor,
		otelgrpc.UnaryServerInterceptor(),
	}

	chainStreamInterceptors := []grpc.StreamServerInterceptor{
		grpc_prometheus.StreamServerInterceptor,
		otelgrpc.StreamServerInterceptor(),
	}

	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(chainUnaryInterceptors...)))
	opts = append(opts, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(chainStreamInterceptors...)))

	opts = append(opts, grpc.MaxRecvMsgSize(grpcOverheadBytes))
	opts = append(opts, grpc.MaxSendMsgSize(maxSendBytes))
	opts = append(opts, grpc.MaxConcurrentStreams(maxStreams))

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterWatchServer(grpcServer, NewWatchServer(watchableStore))

	hsrv := health.NewServer()
	hsrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hsrv)
	// set zero values for metrics registered for this grpc server
	grpc_prometheus.Register(grpcServer)

	return grpcServer
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	klog.Infof("GRPC call: %s, parameters: %+v", info.FullMethod, protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("GRPC call: %s, error: %v", info.FullMethod, err)
	} else {
		klog.Infof("GRPC call %s, response: %+v", info.FullMethod, protosanitizer.StripSecrets(resp))
	}
	return resp, err
}
