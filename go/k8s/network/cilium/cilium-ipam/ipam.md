
# Cilium IPAM
解决方案：https://cilium.slack.com/archives/C1MATJ5U5/p1644297760855179

IPAM: https://docs.cilium.io/en/stable/concepts/networking/ipam/

Cilium Operator 创建 CiliumNode，并通过 IPAM 来从 Cluster CIDR 中分配 Pod CIDR，给 Daemon Agent 使用。
Cilium IPAM Mode: cluster-pool(默认)、crd(自定义)、aws eni、azure、alibaba cloud
cluster-pool: https://docs.cilium.io/en/stable/concepts/networking/ipam/cluster-pool/
crd: https://docs.cilium.io/en/stable/gettingstarted/ipam-crd/

# Cilium Kubernetes IPAM
解决方案：https://cilium.slack.com/archives/C1MATJ5U5/p1644297760855179

cilium ipam:kubernetes 工作机制：**[kubernetes ipam](https://docs.cilium.io/en/stable/concepts/networking/ipam/kubernetes/)**
cilium agent 会等待并 retrieveNodeInformation(corev1.node)，读取 node.Spec.PodCIDRs，没有则读取 node.Spec.PodCIDR，没有则
读取 node.Annotations["io.cilium.network.ipv4-pod-cidr"]，可以见代码：
* https://github.com/cilium/cilium/blob/v1.12.0-rc0/pkg/k8s/init.go#L190-L260
* https://github.com/cilium/cilium/blob/v1.12.0-rc0/pkg/k8s/node.go#L52-L206

## K8s NodeIPAM Controller
**[KEP-2593: Enhanced NodeIPAM to support Discontiguous Cluster CIDR](https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/2593-multiple-cluster-cidrs)** : 该提议基本实现了我们的所有需求，
包括多个 ippool, 每个 node 可以多个 cidr，blockSize 可以根据 node 而变化，以及 ippool.Cidr 可以不连续的，等等。


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


(2) 一个节点可以多个 pod cidr
自定义 IPAM: @see https://mp.weixin.qq.com/s/l0kGo4Fb9NTfLgjQrt88pg


(3) 节点的 pod cidr 可以支持按需动态扩容和回收。监听节点 IP 使用率情况，当节点的 pod cidr 使用率过高时，支持动态扩容。


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
    - 30.216.255.0/24

```


# 参考文献
**[腾讯云自定义 Cilium IPAM](https://mp.weixin.qq.com/s/l0kGo4Fb9NTfLgjQrt88pg)**
