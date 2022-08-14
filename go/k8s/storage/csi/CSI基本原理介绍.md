


# CSI

## CSI 相关组件介绍
* PVController in kube-controller-manager：借助 external-provisioner sidecar 容器，主要负责 PVC/PV 的 bind，以及 PV 的动态创建和删除 Provision/Delete 操作。
因为 PVController 无法直接操作我们部署的 CSI Deployment，需要借助该 sidecar 容器 grpc 调用 CSI ControllerServer service 来创建删除卷。

* ADController in kube-controller-manager：借助 external-attacher sidecar 容器，负责处理 volume 的 attach/detach 操作，将设备挂载/卸载到目标节点，为块存储设计的。
ADController 观察到使用 CSI PV 的 pod 被调度到节点后，调用 in-tree csi plugin 的 Attach() 函数去创建对应的 VolumeAttachment 对象，该对象被 external-attacher sidecar 容器使用来处理 attach/detach 操作。
早期这个操作默认在 kubelet VolumeManager 里实现，后期移动到 kube-controller-manager 中心化处理。
因为 ADController 无法直接操作我们部署的 CSI Deamonset，需要借助该 sidecar 容器 grpc 调用 CSI NodeServer service 的 NodeStageVolume/NodeUnstageVolume 来挂载卸载卷。

* PluginManager in kubelet(除了处理 CSI，还有 DevicePlugin): 负责注册 CSI plugin，kubelet watch 指定目录，CSI 在该目录下创建对用的 socket 实现 CSI IdentityServer service，从而被 kubelet 识别注册。 

* VolumeManager in kubelet: 负责处理 volume 的 Mount/Umount 操作，以及volume 的格式化等操作。Mount/Umount 操作是 kubelet 直接调用 CSI NodeServer service 
的 NodePublishVolume/NodeUnpublishVolume 接口实现。Attach/Detach 操作默认关闭，并逐渐废弃。

* in-tree plugins in kubelet: kubelet 组件内置了很多 volume，比如 csi 框架、ceph rbd、hostPath 等。

> Kubernetes Volumes 与 Docker Volumes 比较(**[Volume概念](https://kubernetes.io/docs/concepts/storage/volumes/)**)
Kubernetes Volume 有自己的生命周期lifecycle，可以持久化。且是mount/unmount到Pod的，而不是container，当Pod退出时，Volume也会退出。Volume就是一个目录。

## CSI 基本内容
如何部署 CSI: 主要就是部署一个 daemonset 和一个 deployment。daemonset 主要用来注册 csi plugin。
deployment controller: csi driver(CSI Controller Service) + sidecar container(external-provisioner, external-attacher, external-snapshotter, and external-resizer)
官方建议部署文档：
https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface.md#recommended-mechanism-for-deploying-csi-drivers-on-kubernetes
https://kubernetes-csi.github.io/docs/deploying.html

### Identity Service RPC
Identity Service RPC: 主要用来 csi 的 info, capability 等信息，纯属基本信息这种。
* info: csi name 和 csi version
* capability: csi.PluginCapability_Service_CONTROLLER_SERVICE, csi.PluginCapability_VolumeExpansion_ONLINE(支持volume扩容)

```
GetPluginInfo(): node-driver-registrar 会rpc调用 Identity Service 的 GetPluginInfo() 获取 csi 的 info name 信息
GetPluginCapabilities(): 

```

### Node Service RPC
https://github.com/container-storage-interface/spec/blob/master/spec.md#node-service-rpc
```
NodeStageVolume:

```


### PV/PVC protection
https://kubernetes.io/zh/docs/concepts/storage/persistent-volumes/#storage-object-in-use-protection


### CSI 原理
plugin_manager pkg 主要去监听 /var/lib/kubelet/plugins socket的注册和注销，代码在 pkg/kubelet/pluginmanager/plugin_manager.go
csi_plugin 主要实现csi定义的方法，如 NodeGetInfo/NodeStageVolume/NodePublishVolume 等方法，而这些方法通过rpc调用 node-driver-registrar
方法，来注册自己写的csi plugin。代码在 pkg/volume/csi/csi_plugin.go。
其中，csi-external-provisioner和csi-external-attacher controller会watch pvc/pv/storageclass 再去调用自己写的csi plugin方法实现create/delete volume，
和attach/detach volume。

### Troubleshoot
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

(4)mount propagation
设计文档： https://github.com/kubernetes/community/blob/master/contributors/design-proposals/node/propagation.md


## 面试题
(1)重点调查下 VolumeAttachment 资源对象是怎么被创建的完整过程？为何 CSIDriver spec.attachRequired=false 就可以控制 k8s 跳过 attach/detach 操作步骤？
ADController in kube-controller-manager 创建的，可以参考上文的该组件介绍。CSIDriver spec.attachRequired=false 会在 ADController 去判断是否 csiPlugin.CanAttach()。


## 参考文献
**[一文读懂 K8s 持久化存储流程](https://mp.weixin.qq.com/s/jpopq16BOA_vrnLmejwEdQ)**

**[一文读懂容器存储接口 CSI](https://mp.weixin.qq.com/s/A9xWKMmrxPyOEiCs_sicYQ)**

**[CSI container storage interface标准文档](https://github.com/container-storage-interface/spec/blob/master/spec.md)**

**[Kubernetes Volume System Redesign Proposal](https://github.com/kubernetes/kubernetes/issues/18333)**

**[Detailed Design for Volume Attach/Detach Controller](https://github.com/kubernetes/kubernetes/issues/20262)**

**[Detailed Design for Volume Mount/Unmount Redesign](https://github.com/kubernetes/kubernetes/issues/21931)**

**[CSI design proposal](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface.md)**

**[Dynamic Provisioning of Volumes design docs](https://github.com/kubernetes/kubernetes/pull/17056)**

**[how to develop a CSI driver 官方文档](https://kubernetes-csi.github.io/docs)**

**[详解 Kubernetes Volume 的实现原理](https://draveness.me/kubernetes-volume)**
