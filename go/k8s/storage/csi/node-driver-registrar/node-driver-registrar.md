


## Node Driver Registrar
使用 kubelet plugin 注册机制来注册 csi driver 信息。


代码和文档：
https://github.com/kubernetes-csi/node-driver-registrar
https://kubernetes-csi.github.io/docs/node-driver-registrar.html

作为sidecar container部署，供kubelet调用时用来注册CSI driver，与kubelet交互方式可以参考： **[Device plugin registration](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/#device-plugin-registration)**




## CSI Plugin 注册机制以及带来的结果？

