package cadvisor

import (
	"testing"

	"k8s.io/klog/v2"
)

// INFO: 必须 linux 环境貌似
func TestNew(test *testing.T) {
	client, err := New(nil, "/var/lib/kubelet", []string{"/kubepods"}, true)
	if err != nil {
		panic(err)
	}

	machineInfo, err := client.MachineInfo()
	if err != nil {
		panic(err)
	}

	klog.Infof("%d cores", machineInfo.NumCores)
}
