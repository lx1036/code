


## Node Driver Registrar
使用 kubelet plugin 注册机制来注册 csi driver 信息。


代码和文档：
https://github.com/kubernetes-csi/node-driver-registrar
https://kubernetes-csi.github.io/docs/node-driver-registrar.html


node-driver-registrar主要解决一个问题：把你的 csi driver 以 plugin registration 方式注册到 kubelet 中。
