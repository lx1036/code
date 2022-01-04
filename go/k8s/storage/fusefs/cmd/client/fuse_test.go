package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"k8s.io/klog/v2"
)

func TestFuseFS(test *testing.T) {
	configFile := "../fuse_360.json"
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("read file err %v", err))
	}
	klog.Info(string(content))
	var mountOption MountOption
	err = json.Unmarshal(content, &mountOption)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("json unmarshal config file err %v", err))
	}
	klog.Infof(mountOption.MountPoint)

	super, err := NewSuper(&mountOption)
	if err != nil {
		klog.Fatal(err)
	}

	total, used := fs.metaClient.Statfs()
	klog.Infof(fmt.Sprintf("volume %s stat: total %dMB, used %dB", mountOption.Volname, total>>20, used))

}
