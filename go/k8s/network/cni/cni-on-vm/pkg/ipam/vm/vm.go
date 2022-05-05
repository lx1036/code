package vm

import (
	"context"
	"net"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
)

type Manager struct {
	ecs *EcsClient
}

func NewVMManager() ipam.API {
	return &Manager{}
}

// AllocateENI INFO: 调用 VM API 创建弹性网卡
func (manager *Manager) AllocateENI(ctx context.Context, vSwitch, securityGroup, instanceID string, trunk bool, ipCount int, eniTags map[string]string) (*types.ENI, error) {
	var eni *types.ENI

	resp, err := manager.ecs.CreateNetworkInterface(ctx, instanceType, vSwitch, []string{securityGroup}, ipv4Count, ipv6Count, eniTags)
	if err != nil {
		return nil, err
	}

	return eni, nil
}

func (manager *Manager) GetAttachedENIs(ctx context.Context, containsMainENI bool) ([]*types.ENI, error) {
	panic("implement me")
}

func (manager *Manager) GetSecondaryENIMACs(ctx context.Context) ([]string, error) {
	panic("implement me")
}

func (manager *Manager) GetENIByMac(ctx context.Context, mac string) (*types.ENI, error) {
	panic("implement me")
}

func (manager *Manager) FreeENI(ctx context.Context, eniID string, instanceID string) error {
	panic("implement me")
}

func (manager *Manager) GetENIIPs(ctx context.Context, mac string) ([]net.IP, []net.IP, error) {
	panic("implement me")
}

func (manager *Manager) AssignNIPsForENI(ctx context.Context, eniID, mac string, count int) ([]types.IPSet, error) {
	panic("implement me")
}

func (manager *Manager) UnAssignIPsForENI(ctx context.Context, eniID, mac string, ipv4s []net.IP, ipv6s []net.IP) error {
	panic("implement me")
}

func (manager *Manager) GetAttachedSecurityGroups(ctx context.Context, instanceID string) ([]string, error) {
	panic("implement me")
}

func (manager *Manager) CheckEniSecurityGroup(ctx context.Context, sgIDs []string) error {
	panic("implement me")
}

func (manager *Manager) DescribeVSwitchByID(ctx context.Context, vSwitch string) (*vpc.VSwitch, error) {
	panic("implement me")
}

func (manager *Manager) AllocateEipAddress(ctx context.Context, bandwidth int, chargeType interface{}, eipID, eniID string, eniIP net.IP, allowRob bool) (*interface{}, error) {
	panic("implement me")
}

func (manager *Manager) UnassociateEipAddress(ctx context.Context, eipID, eniID, eniIP string) error {
	panic("implement me")
}

func (manager *Manager) ReleaseEipAddress(ctx context.Context, eipID, eniID string, eniIP net.IP) error {
	panic("implement me")
}

func (manager *Manager) QueryEniIDByIP(ctx context.Context, vpcID string, address net.IP) (string, error) {
	panic("implement me")
}
