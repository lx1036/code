package volumebinding

// INFO: PV 延迟绑定
//  (1)VolumeBinding plugin 会在 PreBind 阶段给 pvc 加一个 annotation volume.kubernetes.io/selected-node: node1，
//  (2)然后 PVController 根据这个 annotation 然后再给 pvc 加上 annotation volume.beta.kubernetes.io/storage-provisioner: csi.fusefs.com，
//  (3)然后 external-provisioner sidecar 容器根据这个 annotation 再去创建对应的 pv，然后 PVController 周期性去 sync unbound pvc 里去做 bind pvc 操作。
//  而 VolumeBinding plugin 会每秒周期检查等待 bind pvc，发现 bind pvc 结束后则结束 PreBind 阶段。
