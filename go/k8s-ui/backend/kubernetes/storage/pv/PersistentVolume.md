

1. **[详解 Kubernetes Volume 的实现原理](https://draveness.me/kubernetes-volume)**
2. **[CSI container storage interface标准文档](https://github.com/container-storage-interface/spec/blob/master/spec.md)**
3. **[Kubernetes Volume System Redesign Proposal](https://github.com/kubernetes/kubernetes/issues/18333)**
4. **[Detailed Design for Volume Attach/Detach Controller](https://github.com/kubernetes/kubernetes/issues/20262)**
5. **[Detailed Design for Volume Mount/Unmount Redesign](https://github.com/kubernetes/kubernetes/issues/21931)**


**[CSI design proposal](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface.md)**




1. Dynamic Provisioning of Volumes **[design docs](https://github.com/kubernetes/kubernetes/pull/17056)**
2. 


# Kubernetes Volumes 与 Docker Volumes 比较(**[Volume概念](https://kubernetes.io/docs/concepts/storage/volumes/)**)
Kubernetes Volume 有自己的生命周期lifecycle，可以持久化。且是mount/unmount到Pod的，而不是container，当Pod退出时，Volume也会退出。
Volume就是一个目录。


**[how to develop a CSI driver 官方文档](https://kubernetes-csi.github.io/docs)**

# Persistent Volumes


