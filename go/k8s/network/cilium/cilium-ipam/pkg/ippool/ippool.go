package ippool

import (
	"fmt"
	"os"

	apiv3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/projectcalico/calico/libcalico-go/lib/apiconfig"
	client "github.com/projectcalico/calico/libcalico-go/lib/clientv3"
	"github.com/projectcalico/calico/libcalico-go/lib/net"
	"github.com/projectcalico/calico/libcalico-go/lib/selector"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// CreateCalicoClient loads the client config from environments and creates the
// Calico client.
func CreateCalicoClient(filename string) (*apiconfig.CalicoAPIConfig, client.Interface) {
	// Load the client config from environment.
	cfg, err := apiconfig.LoadClientConfig(filename)
	if err != nil {
		fmt.Printf("ERROR: Error loading datastore config: %s\n", err)
		os.Exit(1)
	}
	c, err := client.New(*cfg)
	if err != nil {
		fmt.Printf("ERROR: Error accessing the Calico datastore: %s\n", err)
		os.Exit(1)
	}

	return cfg, c
}

// SelectsNode determines whether or not the IPPool's nodeSelector
// matches the labels on the given node.
func SelectsNode(pool apiv3.IPPool, node corev1.Node) (bool, error) {
	// No node selector means that the pool matches the node.
	if len(pool.Spec.NodeSelector) == 0 {
		return true, nil
	}
	// Check for valid selector syntax.
	sel, err := selector.Parse(pool.Spec.NodeSelector)
	if err != nil {
		return false, err
	}
	// Return whether or not the selector matches.
	return sel.Evaluate(node.Labels), nil
}

func DetermineEnabledIPPoolCIDRs(node corev1.Node, ipPoolList apiv3.IPPoolList) []*net.IPNet {
	var cidrs []*net.IPNet
	for _, ipPool := range ipPoolList.Items {
		_, poolCidr, err := net.ParseCIDR(ipPool.Spec.CIDR)
		if err != nil {
			klog.Errorf(fmt.Sprintf("Failed to parse CIDR %s for IPPool %s err:%v", ipPool.Spec.CIDR, ipPool.Name, err))
			continue
		}

		// Check if IP pool selects the node
		if selects, err := SelectsNode(ipPool, node); err != nil {
			klog.Errorf(fmt.Sprintf("select node %s for IPPool %s CIDR %s err: %v",
				node.Name, ipPool.Name, ipPool.Spec.CIDR, err))
			continue
		} else if selects {
			cidrs = append(cidrs, poolCidr)
		}
	}

	return cidrs
}
