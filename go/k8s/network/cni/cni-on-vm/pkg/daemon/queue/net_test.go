package queue

import (
	"fmt"
	k8snet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/klog/v2"
	"testing"
)

func TestNet(test *testing.T) {
	v4, err := k8snet.ResolveBindAddress(nil) // 获取主机地址
	if err != nil {
		klog.Fatal(err)
	}

	klog.Infof(fmt.Sprintf("%s", v4.String()))

	netConfs := []string{"net1", "net2"}
	for _, conf := range netConfs {
		klog.Info(conf)

		dataPath := "ipvlan"
		avaliable := true
		switch dataPath {
		case "ipvlan":
			if avaliable {
				klog.Info("ipvlan")
				continue
			}
			fallthrough
		case "policyRoute":
			klog.Info("policyRoute")
		default:
			klog.Info("default")
		}
	}
}