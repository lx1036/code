package cilium

import (
	"k8s.io/klog/v2"
	"reflect"
	"testing"
)

func TestLabels(test *testing.T) {
	//labels1 := map[string]string{"a": "b"}
	//labels2 := map[string]string{}
	var labels3 map[string]string
	var labels4 map[string]string
	if reflect.DeepEqual(labels4, labels3) {
		klog.Infof("success")
	} else {
		klog.Infof("fail")
	}
}
