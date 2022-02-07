package reference

import (
	"fmt"
	"k8s.io/klog/v2"
	"testing"
)

func TestParse(test *testing.T) {
	ref := "docker.io/library/nginx:1.17.8"
	spec, err := Parse(ref)
	if err != nil {
		klog.Fatal(err)
	}

	klog.Infof(fmt.Sprintf("%+v", spec)) // {Locator:docker.io/library/redis Object:1.17.8}

	digest := spec.Digest()
	klog.Infof(fmt.Sprintf("%s", digest)) // ""
}
