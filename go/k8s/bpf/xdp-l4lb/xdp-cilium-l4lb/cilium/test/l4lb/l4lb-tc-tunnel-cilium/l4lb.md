


# 验证 l4lb

本 demo 来自于 cilium l4lb demo: https://github.com/cilium/cilium/blob/511463db8f42541cef3730138a58591dce2f3a44/test/l4lb


目标：不需要创建 ipip tunnel 网卡，通过 tc ingress bpf decap 程序来解包。

在 ubuntu 20.04 中完成验证。

软件版本:

```
k8s: v1.19.16
cilium: v1.10.20
```

验证：

```
curl 2.2.2.2:80
```
     

