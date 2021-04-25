package main

import (
	"testing"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
)

/*

## kubelet, cri, docker 交互
* (1) docker daemon 会 serve 在 unix:///var/run/docker.sock

* (2) 服务端：dockershim 会 serve 在 unix:///var/run/dockershim.sock，dockershim 的 DockerService 接口包含了docker官方的
dockerClient对象。而官方 dockerClient 对象会通过 unix:///var/run/docker.sock 和 docker daemon 交互，比如 CreateContainer()/
RemoveContainer()/ListContainers()等操作，server 这块代码见：pkg/kubelet/kubelet_dockershim.go::runDockershim()。初始化dockershim
代码见：https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/kubelet.go#L299-L310
DockerService 接口作为grpc server端，监听在 unix:///var/run/dockershim.sock 。
DockerService 接口就是容器运行时接口CRI，这些接口函数在 staging/src/k8s.io/cri-api/pkg/apis/runtime/v1alpha2/api.proto 定义。

* (3) 客户端：RemoteRuntimeService 作为客户端对象，通过 grpc dial unix:///var/run/dockershim.sock，来调用以上容器操作函数，
代码见：pkg/kubelet/cri/remote/remote_runtime.go::RemoteRuntimeService 对象，初始化代码见：
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/kubelet.go#L319-L321

总之，客户端对象 RemoteRuntimeService grpc dial unix:///var/run/dockershim.sock，通过 CRI 定义的函数去调用监听在这个socket的
server端 DockerService 接口，而 DockerService 包含了 dockerClient 对象，该 dockerClient 对象又会 grpc unix:///var/run/docker.sock
把函数传给 docker daemon。

*/

// INFO: 根据以上信息，这里是没法这样调用的。
func TestRemoteRuntimeService(test *testing.T) {
	endpoint := "unix:///var/run/docker.sock" // unix:///var/run/dockershim.sock
	connectionTimeout := time.Second * 30
	runtimeService, err := remote.NewRemoteRuntimeService(endpoint, connectionTimeout)
	if err != nil {
		panic(err)
	}

	runtimeStatus, err := runtimeService.Status()
	if err != nil {
		panic(err)
	}

	klog.Infof("runtime status: %s", runtimeStatus.String())
}
