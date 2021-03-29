

## CSI Provisioner

本仓库代码合并以下两个仓库：
https://github.com/kubernetes-csi/external-provisioner
https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner



The external-provisioner is a sidecar container that dynamically provisions volumes by calling ControllerCreateVolume and ControllerDeleteVolume functions of CSI drivers.
The external-provisioner is an external controller that monitors PersistentVolumeClaim objects created by user and creates/deletes volumes for them.


> 该csi provisioner可以参考cephfs provisioner实现： https://github.com/kubernetes-retired/external-storage/blob/master/ceph/cephfs/cephfs-provisioner.go



# Kubernetes学习笔记之CSI External Provisioner源码解析

## Overview
最近在部署K8s持久化存储插件时，需要按照CSI官网说明部署一个Deployment pod，由于我们的自研存储类型是文件存储不是块存储，所以部署pod不需要包含容器 **[external-attacher](https://kubernetes-csi.github.io/docs/external-attacher.html)** ，
只需要包含 **[external-provisioner](https://kubernetes-csi.github.io/docs/external-provisioner.html)** sidecar container和我们自研的csi-plugin容器就行，部署yaml类似如下：

```yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "2"
  name: sunnyfs-csi-controller-share
  namespace: sunnyfs
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: sunnyfs-csi-controller-share
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: sunnyfs-csi-controller-share
    spec:
      containers:
        - args:
            - --csi-address=/csi/sunnyfs-provisioner-share.sock
            - --timeout=150s
          image: quay.io/k8scsi/csi-provisioner:v2.0.2
          imagePullPolicy: IfNotPresent
          name: csi-provisioner
          resources:
            limits:
              cpu: "4"
              memory: 8000Mi
            requests:
              cpu: "2"
              memory: 8000Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
        - args:
            - --v=5
            - --endpoint=unix:///csi/sunnyfs-provisioner-share.sock
            - --nodeid=$(NODE_ID)
            - --drivername=csi.sunnyfs.share.com
            - --version=v1.0.0
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
          image: sunnyfs-csi-driver:v1.0.3
          imagePullPolicy: IfNotPresent
          lifecycle:
            preStop:
              exec:
                command:
                  - /bin/sh
                  - -c
                  - rm -rf /csi/sunnyfs-provisioner-share.sock
          name: sunnyfs-csi-plugin
          resources:
            limits:
              cpu: "2"
              memory: 4000Mi
            requests:
              cpu: "1"
              memory: 4000Mi
          securityContext:
            capabilities:
              add:
                - SYS_ADMIN
            privileged: true
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: sunnyfs-csi-controller-account
      serviceAccountName: sunnyfs-csi-controller-account
      terminationGracePeriodSeconds: 30
      volumes:
        - hostPath:
            path: /var/lib/kubelet/plugins/csi.sunnyfs.share.com
            type: DirectoryOrCreate
          name: socket-dir

```


当我们新建一个带有storage class的pvc时，会动态创建pv对象，并在我们自研的存储引擎服务创建对应的volume。这也是利用了 **[storage class](https://kubernetes.io/docs/concepts/storage/storage-classes/)** 来动态创建pv和存储服务对应的volume。

重要问题是，这是如何做到的呢？

答案很简单：external-provisioner sidecar container是一个controller去watch pvc/pv对象，当新建一个由storageclass创建pv的pvc(或删除pv对象)，该sidecar container会grpc调用
我们自研的csi-plugin CreateVolume(DeleteVolume)方法来实际创建一个外部存储volume，并新建一个pv对象写入k8s api server。

## external-provisioner源码解析

external-provisioner sidecar container主要逻辑很简单：
先实例化 **[csiProvisioner对象](https://github.com/kubernetes-csi/external-provisioner/blob/release-2.0/cmd/csi-provisioner/csi-provisioner.go#L247-L270)** ，然后使用
csiProvisioner实例化 **[provisionController](https://github.com/kubernetes-csi/external-provisioner/blob/release-2.0/cmd/csi-provisioner/csi-provisioner.go#L272-L278)** 对象，最后启动
**[provisionController.Run](https://github.com/kubernetes-csi/external-provisioner/blob/release-2.0/cmd/csi-provisioner/csi-provisioner.go#L337-L358)** 去watch pvc/pv对象实现主要业务逻辑，
即根据新建的pvc去调用csi-plugin CreateVolume创建volume，和新建一个pv对象写入k8s api server。

provisionController在实例化时，会watch pvc/pv对象，代码在 **[L695-L739](https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/blob/master/controller/controller.go#L695-L739)** ：

```go

// 实例化provisionController
func NewProvisionController(
	client kubernetes.Interface,
	provisionerName string,
	provisioner Provisioner,
	kubeVersion string,
	options ...func(*ProvisionController) error,
) *ProvisionController {
	// ...
	controller := &ProvisionController{
	client:                    client,
	provisionerName:           provisionerName,
	provisioner:               provisioner, // 在sync pvc时会调用provisioner来创建volume
	// ...
	}
	
	controller.claimQueue = workqueue.NewNamedRateLimitingQueue(rateLimiter, "claims")
	controller.volumeQueue = workqueue.NewNamedRateLimitingQueue(rateLimiter, "volumes")
	informer := informers.NewSharedInformerFactory(client, controller.resyncPeriod)
    // ----------------------
    // PersistentVolumeClaims
	claimHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { controller.enqueueClaim(obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { controller.enqueueClaim(newObj) },
		DeleteFunc: func(obj interface{}) {
			// NOOP. The claim is either in claimsInProgress and in the queue, so it will be processed as usual
			// or it's not in claimsInProgress and then we don't care
		},
	}
    // ...
	// -----------------
	// PersistentVolumes
	volumeHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { controller.enqueueVolume(obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { controller.enqueueVolume(newObj) },
		DeleteFunc: func(obj interface{}) { controller.forgetVolume(obj) },
	}

	// --------------
	// StorageClasses
	// no resource event handler needed for StorageClasses
	if controller.classInformer == nil {
		if controller.kubeVersion.AtLeast(utilversion.MustParseSemantic("v1.6.0")) {
			controller.classInformer = informer.Storage().V1().StorageClasses().Informer()
		} else {
			controller.classInformer = informer.Storage().V1beta1().StorageClasses().Informer()
		}
	}
	controller.classes = controller.classInformer.GetStore()
	
	if controller.createProvisionerPVLimiter != nil {
		// 会调用volumeStore来新建pv对象写入api server中
		controller.volumeStore = NewVolumeStoreQueue(client, controller.createProvisionerPVLimiter, controller.claimsIndexer, controller.eventRecorder)
	} else {
		// ...
	}

	return controller
}

```

这里主要看下新建一个pvc时，是如何调谐的，看代码 **[L933-L986](https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/blob/master/controller/controller.go#L933-L986)** ：

```go

func (ctrl *ProvisionController) processNextVolumeWorkItem(ctx context.Context) bool {
    // ...
	err := func() error {
		// ...
		if err := ctrl.syncVolumeHandler(ctx, key); err != nil {
			// ...
		}
		ctrl.volumeQueue.Forget(obj)
		return nil
	}()
	// ...
	return true
}
func (ctrl *ProvisionController) syncClaimHandler(ctx context.Context, key string) error {
    // ...
    return ctrl.syncClaim(ctx, claimObj)
}
func (ctrl *ProvisionController) syncClaim(ctx context.Context, obj interface{}) error {
    // ...
	// 起始时，在pv controller调谐pvc去更新pvc annotation后，该shouldProvision才会返回true
    should, err := ctrl.shouldProvision(ctx, claim)
    if err != nil {
    	// ...
        return err
    } else if should {
    	// 调用provisioner来创建后端存储服务的volume，调用volumeStore对象创建pv对象并写入k8s api server
        status, err := ctrl.provisionClaimOperation(ctx, claim)
        // ...
        return err
    }
    return nil
}

const (
    annStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
)
func (ctrl *ProvisionController) shouldProvision(ctx context.Context, claim *v1.PersistentVolumeClaim) (bool, error) {
    // ...
	// 这里主要查看pvc是否存在"volume.beta.kubernetes.io/storage-provisioner" annotation，起初创建pvc时是没有该annotation的
	// 该annotation会由kube-controller-manager组件中pv controller去添加，该pv controller也会去watch pvc对象，当发现该pvc定义的storage class
	// 的provisioner定义的plugin不是k8s in-tree plugin，会给该pvc打上"volume.beta.kubernetes.io/storage-provisioner" annotation
	// 可以参考方法 https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/controller/volume/persistentvolume/pv_controller_base.go#L544-L566
	// 所以起始时，在pv controller调谐pvc去更新pvc annotation后，该shouldProvision才会返回true
    if provisioner, found := claim.Annotations[annStorageProvisioner]; found {
        if ctrl.knownProvisioner(provisioner) {
            claimClass := GetPersistentVolumeClaimClass(claim)
            class, err := ctrl.getStorageClass(claimClass)
            // ...
            if class.VolumeBindingMode != nil && *class.VolumeBindingMode == storage.VolumeBindingWaitForFirstConsumer {
                if selectedNode, ok := claim.Annotations[annSelectedNode]; ok && selectedNode != "" {
                    return true, nil
                }
                return false, nil
            }
            return true, nil
        }
    }
    
    return false, nil
}

```

所以，以上代码关键逻辑是provisionClaimOperation函数，该函数主要实现两个业务逻辑：调用provisioner来创建后端存储服务的volume；调用volumeStore对象创建pv对象并写入k8s api server。
查看下 **[provisionClaimOperation代码](https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/blob/master/controller/controller.go#L1324-L1465)** ：

```go

func (ctrl *ProvisionController) provisionClaimOperation(ctx context.Context, claim *v1.PersistentVolumeClaim) (ProvisioningState, error) {
	// ...
	// 准备相关参数
	claimClass := util.GetPersistentVolumeClaimClass(claim)
	pvName := ctrl.getProvisionedVolumeNameForClaim(claim)
	claimRef, err := ref.GetReference(scheme.Scheme, claim)
	class, err := ctrl.getStorageClass(claimClass)
	options := ProvisionOptions{
		StorageClass: class,
		PVName:       pvName,
		PVC:          claim,
		SelectedNode: selectedNode,
	}

    // (1) 调用provisioner来创建后端存储服务的volume
	volume, result, err := ctrl.provisioner.Provision(ctx, options)

	volume.Spec.ClaimRef = claimRef
    // 添加"pv.kubernetes.io/provisioned-by" annotation
	metav1.SetMetaDataAnnotation(&volume.ObjectMeta, annDynamicallyProvisioned, ctrl.provisionerName)
    // (2) 调用volumeStore对象创建pv对象并写入k8s api server
	if err := ctrl.volumeStore.StoreVolume(claim, volume); err != nil {
		return ProvisioningFinished, err
	}
	// 更新本地缓存
	if err = ctrl.volumes.Add(volume); err != nil {
		utilruntime.HandleError(err)
	}
	return ProvisioningFinished, nil
}

```

以上代码主要逻辑比较简单，关键逻辑是调用了 `provisioner.Provision()` 方法创建后端存储服务的volume，看下关键逻辑代码 **[Provision()](https://github.com/kubernetes-csi/external-provisioner/blob/release-2.0/pkg/controller/controller.go#L432-L755)** ：

```go

func (p *csiProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	pvName, err := makeVolumeName(p.volumeNamePrefix, fmt.Sprintf("%s", options.PVC.ObjectMeta.UID), p.volumeNameUUIDLength)
	req := csi.CreateVolumeRequest{
		Name:               pvName,
		Parameters:         options.StorageClass.Parameters,
		VolumeCapabilities: volumeCaps,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: int64(volSizeBytes),
		},
	}
	// 获取 provision secret credentials
	provisionerSecretRef, err := getSecretReference(provisionerSecretParams, options.StorageClass.Parameters, pvName, &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.PVC.Name,
			Namespace: options.PVC.Namespace,
		},
	})
	provisionerCredentials, err := getCredentials(ctx, p.client, provisionerSecretRef)
	req.Secrets = provisionerCredentials
	// ...

	// 关键逻辑：通过grpc调用我们自研csi-plugin中的controller-service CreateVolume方法，在后端存储服务中创建一个真实的volume
	// 该csiClient为controller-service client，controller-service rpc标准可以参考官方文档 https://github.com/container-storage-interface/spec/blob/master/spec.md#controller-service-rpc
	rep, err = p.csiClient.CreateVolume(createCtx, &req)
    // ...
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes:  options.PVC.Spec.AccessModes,
			MountOptions: options.StorageClass.MountOptions,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): bytesToGiQuantity(respCap),
			},
			// TODO wait for CSI VolumeSource API
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:                     p.driverName,
					VolumeHandle:               p.volumeIdToHandle(rep.Volume.VolumeId),
					VolumeAttributes:           volumeAttributes,
					ControllerPublishSecretRef: controllerPublishSecretRef,
					NodeStageSecretRef:         nodeStageSecretRef,
					NodePublishSecretRef:       nodePublishSecretRef,
					ControllerExpandSecretRef:  controllerExpandSecretRef,
				},
			},
		},
	}

	return pv, controller.ProvisioningFinished, nil
}

```

以上代码也比较清晰简单，关键逻辑是通过grpc调用我们自研csi-plugin的controller-service CreateVolume方法来创建外部存储服务中的一个真实volume。

同理，external-provisioner sidecar container也会去watch pv，如果删除pv时，会首先判断是否同时需要删除后端存储服务的真实volume，如果需要
删除则调用provisioner.Delete()，即自研csi-plugin的controller-service DeleteVolume方法去删除volume。删除volume可以参考代码 **[deleteVolumeOperation](https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/blob/master/controller/controller.go#L1467-L1544)** 。

至此，就可以解释当我们创建一个带有storage class的pvc时，external-provisioner sidecar container会watch pvc，并调用provisioner.Provision去
创建volume，而provisioner.CreateVolume又会去调用自研csi-plugin controller-service的CreateVolume()去真实创建一个volume，最后再根据该volume
获取相关pv对象参数，并新建一个pv对象写入k8s api server中。以上过程都是动态创建，自动化的，无需人工操作，这也是storage class的功能。


## 总结
本文主要学习了external-provisioner sidecar container相关原理逻辑，解释了创建一个带有storage class的pvc时，如何新建一个k8s pv对象，以及
如何创建一个后端存储服务的真实volume。

至此，已经有了一个pvc对象，且该pvc对象已经bound了一个带有后端存储服务真实volume的pv，现在就可以在pod内使用这个pvc了，pod containers内的mount path可以像使用本地
目录一样使用这个volume path。但是，该volume path是如何被mount到pod containers中的呢？后续有空再更新。


## 参考文献
**[一文读懂 K8s 持久化存储流程](https://mp.weixin.qq.com/s/jpopq16BOA_vrnLmejwEdQ)**

**[从零开始入门 K8s | Kubernetes 存储架构及插件使用](https://mp.weixin.qq.com/s/QWLGkpqpMdsY1w6npZj-yQ)**

**[Kubernetes Container Storage Interface (CSI) Documentation](https://kubernetes-csi.github.io/docs/introduction.html)**

**[node-driver-registrar](https://github.com/kubernetes-csi/node-driver-registrar)**

**[external-provisioner设计文档](https://github.com/kubernetes-csi/external-provisioner/blob/master/doc/design.md)**
