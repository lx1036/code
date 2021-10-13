package daemon

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"sync"

	"k8s-lx1036/k8s/network/cni/eni/rpc"
	"k8s-lx1036/k8s/network/cni/eni/types"
)

const (
	daemonModeVPC        = "VPC"
	daemonModeENIMultiIP = "ENIMultiIP"
	daemonModeENIOnly    = "ENIOnly"
)

type EniBackendServer struct {
	rpc.UnimplementedEniBackendServer

	daemonMode     string
	configFilePath string
	kubeConfig     string
	master         string

	ipFamily *types.IPFamily
}

func newEniBackendServer(daemonMode string) (rpc.EniBackendServer, error) {
	cniBinPath := os.Getenv("CNI_PATH")
	if cniBinPath == "" {
		cniBinPath = cniDefaultPath
	}
	server := &EniBackendServer{
		configFilePath: configFilePath,
		kubeConfig:     kubeconfig,
		master:         master,
		pendingPods:    sync.Map{},
		cniBinPath:     cniBinPath,
	}

	switch daemonMode {
	case daemonModeENIOnly, daemonModeENIMultiIP, daemonModeVPC:
		server.daemonMode = daemonMode
	default:
		return nil, fmt.Errorf("unsupport daemon mode %s", daemonMode)
	}

	return server, nil
}

func (server *EniBackendServer) AllocateIP(ctx context.Context, request *rpc.AllocateIPRequest) (*rpc.AllocateIPReply, error) {

	allocIPReply := &rpc.AllocateIPReply{IPv4: server.ipFamily.IPv4, IPv6: server.ipFamily.IPv6}

	// 4. grpc connection
	if ctx.Err() != nil {
		err = ctx.Err()
		return nil, errors.Wrapf(err, "error on grpc connection")
	}

	// 5. return allocate result
	return allocIPReply, err
}

func (server *EniBackendServer) ReleaseIP(ctx context.Context, request *rpc.ReleaseIPRequest) (*rpc.ReleaseIPReply, error) {
	panic("implement me")
}

func (server *EniBackendServer) GetIPInfo(ctx context.Context, request *rpc.GetInfoRequest) (*rpc.GetInfoReply, error) {
	panic("implement me")
}

func (server *EniBackendServer) RecordEvent(ctx context.Context, request *rpc.EventRequest) (*rpc.EventReply, error) {
	panic("implement me")
}
