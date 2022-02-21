package ippool

import (
	"context"
	"fmt"
	"testing"

	"github.com/projectcalico/calico/libcalico-go/lib/options"

	"k8s.io/klog/v2"
)

func TestIPPool(test *testing.T) {
	calicoConfig, calicoClient := CreateCalicoClient("")

	klog.Infof(fmt.Sprintf("%+v", calicoConfig.Spec))

	ippoolList, err := calicoClient.IPPools().List(context.TODO(), options.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}

	for _, ippool := range ippoolList.Items {
		klog.Infof(fmt.Sprintf("%+v", ippool.Spec))
	}
}
