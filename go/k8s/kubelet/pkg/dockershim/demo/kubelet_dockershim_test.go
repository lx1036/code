package demo

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	kubeletconfig "k8s-lx1036/k8s/kubelet/pkg/apis/config"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim"
	dockerremote "k8s-lx1036/k8s/kubelet/pkg/dockershim/remote"

	"k8s.io/klog/v2"
)

func TestDockershim(test *testing.T) {
	remoteRuntimeEndpoint := "unix:///var/run/dockershim.sock"
	err := runDockershim(remoteRuntimeEndpoint, remoteRuntimeEndpoint)
	if err != nil {
		panic(err)
	}
}

// remoteRuntimeEndpoint=remoteImageEndpoint="unix:///var/run/dockershim.sock"
func runDockershim(remoteRuntimeEndpoint string, remoteImageEndpoint string) error {
	abs, err := filepath.Abs("../fixtures")
	if err != nil {
		panic(err)
	}
	pluginConfDir := fmt.Sprintf("%s%s", abs, "/etc/cni/net.d")
	klog.Info(fmt.Sprintf("cni conf dir: %s", pluginConfDir))
	nonMasqueradeCIDR := "10.0.0.0/8"
	pluginSettings := dockershim.NetworkPluginSettings{
		HairpinMode:        kubeletconfig.HairpinMode(kubeletconfig.PromiscuousBridge),
		NonMasqueradeCIDR:  nonMasqueradeCIDR,
		PluginName:         "cni",
		PluginConfDir:      pluginConfDir,
		PluginBinDirString: "/opt/cni/bin",
		PluginCacheDir:     "/var/lib/cni/cache",
		MTU:                0,
	}
	dockerClientConfig := &dockershim.ClientConfig{
		DockerEndpoint:            "unix:///var/run/docker.sock",
		RuntimeRequestTimeout:     2 * time.Minute,
		ImagePullProgressDeadline: 1 * time.Minute,
	}
	cgroupDriver := "cgroupfs"
	dockershimRootDirectory := "/tmp/dockershim" // 默认是 /var/lib/dockershim, mac 上写 /tmp/dockershim
	podSandboxImage := "k8s.gcr.io/pause:3.2"
	runtimeCgroups := ""
	dockerService, err := dockershim.NewDockerService(dockerClientConfig, podSandboxImage,
		&pluginSettings, runtimeCgroups, cgroupDriver, dockershimRootDirectory)
	if err != nil {
		return err
	}

	// The unix socket for kubelet <-> dockershim communication, dockershim start before runtime service init.
	klog.V(5).InfoS("Using remote runtime endpoint and image endpoint", "runtimeEndpoint", remoteRuntimeEndpoint, "imageEndpoint", remoteImageEndpoint)
	klog.V(2).InfoS("Starting the GRPC server for the docker CRI shim.")

	// setup grpc server
	dockerServer := dockerremote.NewDockerServer(remoteRuntimeEndpoint, dockerService)
	if err := dockerServer.Start(); err != nil {
		return err
	}

	return nil
}
