

## Kubernetes学习笔记之CSI Plugin注册机制源码解析






























```text

第 1 步，在启动的时候有一个约定，比如说在 `/var/lib/kuberlet/plugins_registry` 这个目录每新加一个文件，就相当于每新加了一个 Plugin；
启动 Node-Driver-Registrar，它首先会向 CSI-Plugin 发起一个接口调用 GetPluginInfo，这个接口会返回 CSI 所监听的地址以及 CSI-Plugin 的一个 Driver name；

第 2 步，Node-Driver-Registrar 会监听 GetInfo 和 NotifyRegistrationStatus 两个接口；

第 3 步，会在 `/var/lib/kuberlet/plugins_registry` 这个目录下启动一个 Socket，生成一个 Socket 文件 ，例如 "diskplugin.csi.alibabacloud.com-reg.sock"，此时 Kubelet 通过 Watcher 发现这个 Socket 后，它会通过该 Socket 向 Node-Driver-Registrar 的 GetInfo 接口进行调用。GetInfo 会把刚才我们所获得的的 CSI-Plugin 的信息返回给 Kubelet，该信息包含了 CSI-Plugin 的监听地址以及它的 Driver name；

第 4 步，Kubelet 通过得到的监听地址对 CSI-Plugin 的 NodeGetInfo 接口进行调用；

第 5 步，调用成功之后，Kubelet 会去更新一些状态信息，比如节点的 Annotations、Labels、status.allocatable 等信息，同时会创建一个 CSINode 对象；

第 6 步，通过对 Node-Driver-Registrar 的 NotifyRegistrationStatus 接口的调用告诉它我们已经把 CSI-Plugin 注册成功了。

通过以上 6 步就实现了 CSI Plugin 注册机制。

```
