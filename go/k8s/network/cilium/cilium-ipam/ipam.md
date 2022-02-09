
# 需求
(1) 像 calico 一样配置多个 multi pool, 每一个 multi pool 可以根据 nodeSelector 配置
@see https://projectcalico.docs.tigera.io/getting-started/kubernetes/hardway/configure-ip-pools
@see https://github.com/cilium/cilium/issues/13227#issuecomment-698150732
@see https://docs.cilium.io/en/stable/gettingstarted/ipam-crd/

```yaml
apiVersion: projectcalico.org/v3
kind: IPPool
metadata:
  name: 100.162.224.0
spec:
  blockSize: 27
  cidr: 100.162.224.0/19
  nodeSelector: topology.kubernetes.io/zone == "beijing"
```


(2) 一个节点可以多个 pod cidr, 节点的 pod cidr 可以支持按需动态扩容和回收
自定义 IPAM: @see https://mp.weixin.qq.com/s/l0kGo4Fb9NTfLgjQrt88pg

# 设计
(1) choose specified IPPool ippool1 based on nodeSelector
(2) get node cidr from ippool1 for specified node
(3) create CiliumNode(include pool and podCIDRs)
```yaml
apiVersion: cilium.io/v2
kind: CiliumNode
metadata:
  name: node1
spec:
  ipam:
    pool:
      20.216.255.1: { }
      20.216.255.2: { }
      20.216.255.3: { }
      20.216.255.4: { }
      20.216.255.5: { }
      20.216.255.6: { }
      20.216.255.7: { }
      20.216.255.8: { }
      20.216.255.9: { }
      20.216.255.10: { }
      20.216.255.11: { }
      20.216.255.12: { }
      20.216.255.13: { }
      20.216.255.14: { }
      20.216.255.15: { }
      20.216.255.16: { }
      20.216.255.17: { }
      20.216.255.18: { }
      20.216.255.19: { }
      20.216.255.20: { }
    podCIDRs:
    - 20.216.255.0/24

```
