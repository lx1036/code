
# CRI
CRI 标准文档：https://github.com/kubernetes/cri-api






# CRI Privileged
https://kubernetes.io/zh/docs/concepts/policy/pod-security-policy/#privileged
Privileged - 决定是否 Pod 中的某容器可以启用特权模式。默认情况下，容器是不可以访问宿主上的任何设备的，不过一个“privileged（特权的）” 容器则被授权访问宿主上所有设备。 
这种容器几乎享有宿主上运行的进程的所有访问权限。 对于需要使用 Linux 权能字（如操控网络堆栈和访问设备）的容器而言是有用的。
> 所以运行 docker run fuse-client 容器时需要加上 --privileged，csi-pod 也是需要加上 securityContext: {privileged: true} 配置



## 参考文献
拦截kubelet发给docker/containerd请求，修改cgroup parent path: 
https://github.com/Tencent/caelus/blob/master/contrib/lighthouse-plugin/README.md


