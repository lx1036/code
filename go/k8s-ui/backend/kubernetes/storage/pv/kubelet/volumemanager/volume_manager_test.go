package volumemanager

import (
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	utiltesting "k8s.io/client-go/util/testing"
	"k8s.io/kubernetes/pkg/kubelet/configmap"
	kubepod "k8s.io/kubernetes/pkg/kubelet/pod"
	podtest "k8s.io/kubernetes/pkg/kubelet/pod/testing"
	"k8s.io/kubernetes/pkg/kubelet/secret"
	"k8s.io/kubernetes/pkg/kubelet/volumemanager"
	"os"
	"testing"
)

func TestGetMountedVolumesForPodAndGetVolumesInUse(test *testing.T) {
	fixtures := []struct {
		name                string
		pvMode, podMode     v1.PersistentVolumeMode
		disableBlockFeature bool
		expectMount         bool
		expectError         bool
	}{
		{
			name:        "filesystem volume",
			pvMode:      v1.PersistentVolumeFilesystem,
			podMode:     v1.PersistentVolumeFilesystem,
			expectMount: true,
			expectError: false,
		},
		{
			name:        "block volume",
			pvMode:      v1.PersistentVolumeBlock,
			podMode:     v1.PersistentVolumeBlock,
			expectMount: true,
			expectError: false,
		},
		{
			name:                "block volume with block feature off",
			pvMode:              v1.PersistentVolumeBlock,
			podMode:             v1.PersistentVolumeBlock,
			disableBlockFeature: true,
			expectMount:         false,
			expectError:         false,
		},
		{
			name:        "mismatched volume",
			pvMode:      v1.PersistentVolumeBlock,
			podMode:     v1.PersistentVolumeFilesystem,
			expectMount: false,
			expectError: true,
		},
	}

	for _, fixture := range fixtures {
		test.Run(fixture.name, func(test *testing.T) {
			if fixture.disableBlockFeature {

			}
			tmpDir, err := utiltesting.MkTmpdir("volumeManagerTest")
			if err != nil {
				test.Fatalf("can't make a temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			cpm := podtest.NewMockCheckpointManager()
			podManager := kubepod.NewBasicPodManager(podtest.NewFakeMirrorClient(), secret.NewFakeManager(), configmap.NewFakeManager(), cpm)
			node, pod, pv, claim := createObjects(fixture.pvMode, fixture.podMode)
			kubeClient := fake.NewSimpleClientset(node, pod, pv, claim)
			volumeManager := newTestVolumeManager(test, tmpDir, podManager, kubeClient)
		})
	}
}

// createObjects returns objects for making a fake clientset. The pv is
// already attached to the node and bound to the claim used by the pod.
func createObjects(pvMode, podMode v1.PersistentVolumeMode) (*v1.Node, *v1.Pod, *v1.PersistentVolume, *v1.PersistentVolumeClaim) {

}

func newTestVolumeManager(t *testing.T, tmpDir string, podManager kubepod.Manager, kubeClient clientset.Interface) volumemanager.VolumeManager {

}
