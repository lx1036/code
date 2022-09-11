
# K8S Configmap 和 Secret 作为 Volume 的热更新原理
configmap/secret 作为 volume 挂载在容器内，如果 configmap 值发生变化，最大等待时间在 kubelet resyncInterval(60s) 内
该 mount 的 key 就会变成最新值。比如 cilium pod 挂载 cilium-config configmap，如果修改该 configmap 的 debug:false 为 true，
最多等待 60s，容器内该 debug 文件值就是 true。

但是作为环境变量 env 和 volume subpath 不支持热更新，环境变量在初始化过程就固定了。

# 热更新原理
(1) kubelet 会在每 60s 内去 syncPod()，检查 pod 的 volume kubelet.volumeManager.WaitForAttachAndMount(pod)，
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/kubelet.go#L1592-L1600
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/volumemanager/volume_manager.go#L375-L378

这里重点是 ReprocessPod()，会把这个 pod 又标记为未处理，等待 desiredStateOfWorldPopulator 下一次循环去 MarkRemountRequired()

(2) desiredStateOfWorldPopulator 下一次循环，会走 findAndAddNewPods() -> processPodVolumes()
这里重点是 dswp.actualStateOfWorld.MarkRemountRequired(uniquePodName)，在 actual 里 MarkRemountRequired
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/volumemanager/populator/desired_state_of_world_populator.go#L358-L364
这里会判断每一个 volumePlugin.RequiresRemount()，而对于 configmap/secret volume 是 true，对于 csi 是 false
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/volumemanager/cache/actual_state_of_world.go#L541-L566
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/volume/configmap/configmap.go#L81-L83
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/volume/csi/csi_plugin.go#L337-L339

(3) 然后再下一次循环里去 mountAttachVolumes() 
PodExistsInVolume() 会走 podObj.remountRequired，因为 MarkRemountRequired() 已经设置了需要 remount，然后 mountAttachVolumes() 里走
MountVolume() 逻辑：https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/volumemanager/reconciler/reconciler.go#L247-L273
这样就走 configmap/secret mount 逻辑。

(4) configmap/secret mount 会使用 emptyDir plugin 来创建落盘目录
configmap 用的 v1.StorageMediumDefault https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/volume/configmap/configmap.go#L166-L174

secret 用的 v1.StorageMediumMemory https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/volume/secret/secret.go#L51-L55 ，
对于 secret 首次 mount 会使用命令 `mount -t tmpfs xxx`:
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/volume/emptydir/empty_dir.go#L232-L233
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/volume/emptydir/empty_dir.go#L265-L286

对于 configmap 这里的 wrapped 是 emptyDir，主要用来创建文件和权限
```go
wrapped, err := b.plugin.host.NewWrapperMounter(b.volName, wrappedVolumeSpec(), &b.pod, *b.opts)
wrapped.SetUpAt(dir, mounterArgs)

// 这里的 getConfigMap 是 configmapManager 的 configMapManager.GetConfigMap()
// https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/configmap/configmap_manager.go#L82-L91
// 注意，kubelet 默认使用 kubeletconfiginternal.WatchChangeDetectionStrategy 的 configmapManager，所以 configmap
// 发生变化，configmapManager 立刻拿到最新的值：https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/kubelet.go#L538-L540
// 只是需要等待 kubelet 每次的 resyncInterval 60s 去 syncPod，所以每次修改 configmap 最大等待时间是 60s。
configMap, err := b.getConfigMap(b.pod.Namespace, b.source.Name)

// 然后把最新的 configmap 对象数据写到每一个文件里
payload, err := MakePayload(b.source.Items, configMap, b.source.DefaultMode, optional)
err = writer.Write(payload)

```


# 参考文献

**[mounted-configmaps-are-updated-automatically](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#mounted-configmaps-are-updated-automatically)**

**[mounted-configmaps-are-updated-automatically](https://kubernetes.io/docs/concepts/configuration/configmap/#mounted-configmaps-are-updated-automatically)**

**[Kubernetes Pod 中的 ConfigMap 配置更新](https://dockone.io/article/8632)**

**[分别测试使用 ConfigMap 挂载 Env 和 Volume 的情况](https://codeantenna.com/a/pf1zJAzHF6)**

开始只有 NewCachingConfigMapManager()，除了 kubelet resyncInterval 时间还有个 ttl 时间，经过讨论后期加了 NewWatchingConfigMapManager,
直接 watch 立刻拿到最新的 configmap，只需要等待最大 kubelet resyncInterval 时间。下面链接是 issue 和 pr：

**[Kubelet watches necessary secrets/configmaps instead of periodic polling](https://github.com/kubernetes/kubernetes/pull/64752)**

**[Migrate kubelet to ConfigMapManager interface and use TTL-based caching manager](https://github.com/kubernetes/kubernetes/pull/46470)**

**[kubelet refresh times for configmaps is long and random](https://github.com/kubernetes/kubernetes/issues/30189)**