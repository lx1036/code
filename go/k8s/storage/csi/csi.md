


# CSI







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


**[详解 Kubernetes Volume 的实现原理](https://draveness.me/kubernetes-volume)**





# CSI Service RPC 
如何部署 CSI: 主要就是部署一个 daemonset 和一个 deployment。daemonset 主要用来注册 csi plugin。
deployment controller: csi driver(CSI Controller Service) + sidecar container(external-provisioner, external-attacher, external-snapshotter, and external-resizer)



官方建议部署文档：
https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface.md#recommended-mechanism-for-deploying-csi-drivers-on-kubernetes
https://kubernetes-csi.github.io/docs/deploying.html


## Identity Service RPC
Identity Service RPC: 主要用来 csi 的 info, capability 等信息，纯属基本信息这种。
* info: csi name 和 csi version
* capability: csi.PluginCapability_Service_CONTROLLER_SERVICE, csi.PluginCapability_VolumeExpansion_ONLINE(支持volume扩容)




```
GetPluginInfo(): node-driver-registrar 会rpc调用 Identity Service 的 GetPluginInfo() 获取 csi 的 info name 信息
GetPluginCapabilities(): 

```


## Node Service RPC
https://github.com/container-storage-interface/spec/blob/master/spec.md#node-service-rpc
```
NodeStageVolume:

```


## PV/PVC protection

https://kubernetes.io/zh/docs/concepts/storage/persistent-volumes/#storage-object-in-use-protection


## CSI 原理
plugin_manager pkg 主要去监听 /var/lib/kubelet/plugins socket的注册和注销，代码在 pkg/kubelet/pluginmanager/plugin_manager.go
csi_plugin 主要实现csi定义的方法，如 NodeGetInfo/NodeStageVolume/NodePublishVolume 等方法，而这些方法通过rpc调用 node-driver-registrar
方法，来注册自己写的csi plugin。代码在 pkg/volume/csi/csi_plugin.go。
其中，csi-external-provisioner和csi-external-attacher controller会watch pvc/pv/storageclass 再去调用自己写的csi plugin方法实现create/delete volume，
和attach/detach volume。

### mount propagation
设计文档： https://github.com/kubernetes/community/blob/master/contributors/design-proposals/node/propagation.md


## Troubleshoot
(1)csi 没法平滑升级
juicefs csi 的解决方案：https://mp.weixin.qq.com/s/hPupPQmCPKZpGIA4SCPBzQ
让 csi-driver pod 去处理 corrupted mount point PR，早期 kubelet 这里逻辑是如果是 corrupted，则直接返回，没有给机会处理: 
https://github.com/kubernetes/kubernetes/pull/88569
chubaofs-csi 使用 VolumeAttachment 解决：https://github.com/chubaofs/chubaofs-csi/pull/54


(2)为何一些volume drivers，如NFS，或者一些FS，不需要 attach operation，CSIDriver里 `attachRequired: false`？
```yaml
# csidriver 需要部署时创建，重点是 podInfoOnMount 参数，见：
# CSIDriver: https://kubernetes-csi.github.io/docs/csi-driver-object.html
# Skip Attach: https://kubernetes-csi.github.io/docs/skip-attach.html
# Pod Info on Mount: https://kubernetes-csi.github.io/docs/pod-info.html
apiVersion: storage.k8s.io/v1
kind: CSIDriver
metadata:
  name: csi.lxfs.com
spec:
  podInfoOnMount: true
  attachRequired: false # controller-server没有实现ControllerPublishVolume()，不需要volume attach operation
  volumeLifecycleModes:
    - Persistent
```


(3)对于 fusefs csi 创建的 pvc/pv，如果 pvc 被删除了，但是 pv 还在，怎么恢复？
1. 先删除 pv claimRef
2. 再新建 pvc.yaml，必须指定 volumeName
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: fusefs-pvc-test
  namespace: default
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 5Gi
  storageClassName: fusefs-storageclass
  volumeMode: Filesystem
  volumeName: pvc-44141bdf-a3bb-4019-bd16-e0b2770a1570 # 必须指定 pv name!!!
```

原理：https://github.com/kubernetes/kubernetes/blob/v1.23.1/pkg/controller/volume/persistentvolume/pv_controller.go#L412-L432



## 面试题
重点调查下 VolumeAttachment 资源对象是怎么被创建的完整过程？？
为何 CSIDriver spec.attachRequired=false 就可以控制 k8s 跳过 attach/detach 操作步骤？


## 参考文献
**[一文读懂 K8s 持久化存储流程](https://mp.weixin.qq.com/s/jpopq16BOA_vrnLmejwEdQ)**

**[一文读懂容器存储接口 CSI](https://mp.weixin.qq.com/s/A9xWKMmrxPyOEiCs_sicYQ)**

