package demo_node_status

import (
	"path/filepath"

	cadvisorapiv1 "github.com/google/cadvisor/info/v1"
	"k8s.io/kubernetes/pkg/kubelet/config"
)

func (kl *Kubelet) setCachedMachineInfo(info *cadvisorapiv1.MachineInfo) {
	kl.machineInfoLock.Lock()
	defer kl.machineInfoLock.Unlock()
	kl.machineInfo = info
}

// getPluginsRegistrationDir returns the full path to the directory under which
// plugins socket should be placed to be registered.
// More information is available about plugin registration in the pluginwatcher
// module
func (kl *Kubelet) getPluginsRegistrationDir() string {
	return filepath.Join(kl.getRootDir(), config.DefaultKubeletPluginsRegistrationDirName)
}

// getRootDir returns the full path to the directory under which kubelet can
// store data.  These functions are useful to pass interfaces to other modules
// that may need to know where to write data without getting a whole kubelet
// instance.
func (kl *Kubelet) getRootDir() string {
	return kl.rootDirectory
}

// getPodsDir returns the full path to the directory under which pod
// directories are created.
func (kl *Kubelet) getPodsDir() string {
	return filepath.Join(kl.getRootDir(), config.DefaultKubeletPodsDirName)
}

// getPluginsDir returns the full path to the directory under which plugin
// directories are created.  Plugins can use these directories for data that
// they need to persist.  Plugins should create subdirectories under this named
// after their own names.
func (kl *Kubelet) getPluginsDir() string {
	return filepath.Join(kl.getRootDir(), config.DefaultKubeletPluginsDirName)
}

// getPodResourcesSocket returns the full path to the directory containing the pod resources socket
func (kl *Kubelet) getPodResourcesDir() string {
	return filepath.Join(kl.getRootDir(), config.DefaultKubeletPodResourcesDirName)
}
