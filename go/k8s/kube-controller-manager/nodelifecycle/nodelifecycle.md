

# NodeLifecycle
https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/controller/nodelifecycle/node_lifecycle_controller.go
作用：给坏 node 打taint，并且过了 podEvictionTimeout 时间后，驱逐这个坏 node 上的 pod。这样可以提高稳定性。 

## 原理解析
TODO: 



## TroubleShooting

