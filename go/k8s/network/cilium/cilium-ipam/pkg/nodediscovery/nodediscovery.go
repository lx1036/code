package nodediscovery

import (
	"context"
	"fmt"

	ipamTypes "github.com/cilium/cilium/pkg/ipam/types"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"

	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// @see https://github.com/cilium/cilium/blob/v1.11.1/pkg/nodediscovery/nodediscovery.go

// 初始化创建 CiliumNode 对象

const (
	maxRetryCount = 10
)

// NodeDiscovery represents a node discovery action
type NodeDiscovery struct {
}

// UpdateCiliumNodeResource updates the CiliumNode resource representing the
// local node
func (n *NodeDiscovery) UpdateCiliumNodeResource() {
	klog.Infof(fmt.Sprintf("Creating or updating CiliumNode resource for node:%s", node))

	performGet := true
	for retryCount := 0; retryCount < maxRetryCount; retryCount++ {
		var nodeResource *ciliumv2.CiliumNode
		performUpdate := true
		if performGet {
			var err error
			nodeResource, err = ciliumClient.CiliumV2().CiliumNodes().Get(context.TODO(), nodeTypes.GetName(), metav1.GetOptions{})
			if err != nil {
				performUpdate = false
				nodeResource = &ciliumv2.CiliumNode{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
					Spec: ciliumv2.NodeSpec{
						IPAM: ipamTypes.IPAMSpec{
							PodCIDRs: []string{},
							Pool:     map[string]ipamTypes.AllocationIP{},
						},
					},
				}
			} else {
				performGet = false
			}
		}

		if err := n.mutateNodeResource(nodeResource); err != nil {
			log.WithError(err).WithField("retryCount", retryCount).Warning("Unable to mutate nodeResource")
			continue
		}

		// if we retry after this point, is due to a conflict. We will do
		// a new GET  to ensure we have the latest information before
		// updating.
		performGet = true
		if performUpdate {
			if _, err := ciliumClient.CiliumV2().CiliumNodes().Update(context.TODO(), nodeResource, metav1.UpdateOptions{}); err != nil {
				if k8serrors.IsConflict(err) {
					log.WithError(err).Warn("Unable to update CiliumNode resource, will retry")
					continue
				}
				log.WithError(err).Fatal("Unable to update CiliumNode resource")
			} else {
				return
			}
		} else {
			if _, err := ciliumClient.CiliumV2().CiliumNodes().Create(context.TODO(), nodeResource, metav1.CreateOptions{}); err != nil {
				if k8serrors.IsConflict(err) {
					log.WithError(err).Warn("Unable to create CiliumNode resource, will retry")
					continue
				}
				log.WithError(err).Fatal("Unable to create CiliumNode resource")
			} else {
				log.Info("Successfully created CiliumNode resource")
				return
			}
		}
	}

	klog.Fatal("Could not create or update CiliumNode resource, despite retries")
}

func (n *NodeDiscovery) mutateNodeResource(nodeResource *ciliumv2.CiliumNode) error {

}
