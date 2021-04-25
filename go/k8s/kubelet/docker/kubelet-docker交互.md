



```go

dockerEndpoint = "unix:///var/run/docker.sock"
remoteRuntimeEndpoint = "unix:///var/run/dockershim.sock"

```



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


## CRI
接口在 k8s 源码 staging/src/k8s.io/cri-api/pkg/apis/runtime/v1alpha2/api.proto 中定义。

```shell
# 下载 crictl CRI 容器工具，类似于 docker cli
wget https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.21.0/crictl-v1.21.0-darwin-amd64.tar.gz
tar zxf crictl-v1.21.0-darwin-amd64.tar.gz
mv crictl /usr/local/bin/
```
