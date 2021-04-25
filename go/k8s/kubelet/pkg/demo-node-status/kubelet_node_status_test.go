package demo_node_status

import (
	"fmt"
	goruntime "runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cadvisorapi "github.com/google/cadvisor/info/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

func notImplemented(action core.Action) (bool, runtime.Object, error) {
	return true, nil, fmt.Errorf("no reaction implemented for %s", action)
}

func addNotImplatedReaction(kubeClient *fake.Clientset) {
	if kubeClient == nil {
		return
	}

	kubeClient.AddReactor("*", "*", notImplemented)
}

// INFO: kubelet 启动时，会向 apiserver 中注册一个 node 新对象，或者更新已有的 node 对象
func TestRegisterWithApiServer(test *testing.T) {
	testKubelet := newTestKubelet(test, false /* controllerAttachDetachEnabled */)
	defer testKubelet.Cleanup()

	kubelet := testKubelet.kubelet
	kubeClient := testKubelet.fakeKubeClient
	kubeClient.AddReactor("create", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		// Return an error on create.
		return true, &v1.Node{}, &apierrors.StatusError{
			ErrStatus: metav1.Status{Reason: metav1.StatusReasonAlreadyExists},
		}
	})
	kubeClient.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		// Return an existing (matching) node on get.
		return true, &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: testKubeletHostname,
				Labels: map[string]string{
					v1.LabelHostname:      testKubeletHostname,
					v1.LabelOSStable:      goruntime.GOOS,
					v1.LabelArchStable:    goruntime.GOARCH,
					kubeletapis.LabelOS:   goruntime.GOOS,
					kubeletapis.LabelArch: goruntime.GOARCH,
				},
			},
		}, nil
	})
	kubeClient.AddReactor("patch", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "status" {
			return true, nil, nil
		}
		return notImplemented(action)
	})

	addNotImplatedReaction(kubeClient)

	machineInfo := &cadvisorapi.MachineInfo{
		MachineID:      "123",
		SystemUUID:     "abc",
		BootID:         "1b3",
		NumCores:       2,
		MemoryCapacity: 1024,
	}
	kubelet.setCachedMachineInfo(machineInfo)

	done := make(chan struct{})
	go func() {
		kubelet.registerWithAPIServer()
		done <- struct{}{}
	}()
	select {
	case <-time.After(wait.ForeverTestTimeout):
		assert.Fail(test, "timed out waiting for registration")
	case <-done:
		return
	}
}

func TestUpdateNewNodeStatus(test *testing.T) {
	fixtures := []struct {
		desc                string
		nodeStatusMaxImages int32
	}{
		{
			desc:                "5 image limit",
			nodeStatusMaxImages: 5,
		},
		{
			desc:                "no image limit",
			nodeStatusMaxImages: -1,
		},
	}

	for _, fixture := range fixtures {
		test.Run(fixture.desc, func(t *testing.T) {

		})
	}

}
