
# 网络安全 Network Policy
@see https://github.com/cloudnativelabs/kube-router/blob/master/docs/how-it-works.md#pod-ingress-firewall
使用 iptables, ipset and conntrack 技术实现 Network Policy。并挂载在 filter table。

```shell
yum install -y conntrack
```


(1)Terway + Cilium
https://github.com/AliyunContainerService/terway/blob/main/docs/terway-with-cilium.md

(2)kube-router
https://github.com/cloudnativelabs/kube-router/blob/master/README.md#network-policy-controller----run-firewall



## 测试案例
(1) 不同 namespace 下的 pod 互相隔离，相同 namespace 下的 pod 可以访问
```yaml
# foo namespace 下的 pod 可以相互访问，其他 namespace 的 pod 不能访问 foo namespace 下的 pod 
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: foo
spec:
  ingress:
  - from:
    - podSelector: {}
  podSelector: {}
  policyTypes:
    - Ingress
```

(2) 不同 namespace 下的 pod 互相隔离，相同 namespace 下的 pod 也不可以访问
```yaml
# foo namespace 下的 pod 不能被任何 pod 访问，foo namespace 下的也不行
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: foo
spec:
  podSelector: {}
  policyTypes:
    - Ingress
```

(3) 不同 namespace 下的 pod 互相隔离，但是白名单里的 B 可以访问当前 namespace 下的 A
```yaml
# 所有 label app=bar 的 pod 可以访问 foo namespace 下的 tcp:6379 端口
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: foo
spec:
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          app: bar
    ports:
      - protocol: TCP
        port: 6379
  podSelector: {}
  policyTypes:
    - Ingress
```

(4) 允许指定 namespace 下的 pod 可以访问 K8s 集群外的指定 CIDR，其他外部 IP 全部隔离
```yaml
# 只有 foo namespace 下的 pod 可以访问网段 14.215.0.0/16 的 5978 端口，可以应用在为集群内特定服务开启访问外部服务的白名单
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: foo
spec:
  egress:
  - to:
    - ipBlock:
        cidr: 14.215.0.0/16
    ports:
    - protocol: TCP
      port: 5978
  podSelector: {}
  policyTypes:
    - Egress
```

(5) 不同 namespace 下的 pod 互相隔离，但是白名单里的 B 可以访问当前 namespace 下的 A 中对应的 Pod 以及端口
```yaml
# foo namespace 下的 pod，只有外部 IP 在网段 14.215.0.0/16 才可以访问 5978 端口
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: foo
spec:
  ingress:
  - from:
    - ipBlock:
        cidr: 14.215.0.0/16
    ports:
    - protocol: TCP
      port: 5978
  podSelector: {}
  policyTypes:
    - Ingress
```

(6) 以上用例，当 source pod 和 destination pod 在同一个 node 时，隔离是否生效


## 参考文献
**[K8s Network Policy Controller之Kube-router功能介绍](https://tencentcloudcontainerteam.github.io/2018/10/30/k8s-npc-kr-function/)**
**[K8s Network Policy](https://kubernetes.io/zh/docs/concepts/services-networking/network-policies/)**

