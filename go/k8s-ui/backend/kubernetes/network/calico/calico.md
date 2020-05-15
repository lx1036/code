
# Calico
**[Calico docs](https://docs.projectcalico.org/introduction/)**

## calico architecture
**[calico architecture and network foundamentals](https://www.tigera.io/video/tigera-calico-fundamentals)**

![how-it-works](./imgs/how-it-works.png)

**[calico architecture](https://docs.projectcalico.org/reference/architecture/overview)** 主要有以下几部分组成:
* Felix: calico agent，运行在每一个 Node 节点上
* Orchestrator plugin: 编排器插件，比如k8s 插件，这样可以使k8s调用这个插件来调用calico。
* etcd: data store
* BIRD: BGP(border gateway protocol) client，分发路由信息。
BIRD是布拉格查理大学数学与物理学院的一个学校项目，项目名是BIRD Internet Routing Daemon的缩写。目前，它由CZ.NIC实验室开发和支持。
* BGP route reflector: 大型网络仅仅使用 BGP client 形成 mesh 全网互联的方案就会导致规模限制，
所有节点需要 N^2 个连接，为了解决这个规模问题，可以采用 BGP 的 Router Reflector 的方法，使所有 BGP Client 仅与特定 RR 节点互联并做路由同步，
从而大大减少连接数。
route reflector 路由反射器：提供了在大型IBGP实现中IBGP全网状连接问题的一个简单解决方案。
* calicoctl: calico 命令行管理工具

## calicoctl:
```shell script
brew install calicoctl
DATASTORE_TYPE=kubernetes KUBECONFIG=~/.kube/config calicoctl get nodes
```

## concepts
**[Calico on Kubernetes 从入门到精通](https://www.kubernetes.org.cn/4960.html)**


### calico in minikube
```shell script
minikube start --network-plugin=cni
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
# or
cd install && make install

# validate
curl $(minikube ip):$(kubectl get svc nginx-demo-1 -o=jsonpath='{.spec.ports[0].nodePort}') -v
```

### network policy
**[k8s network policy](https://docs.projectcalico.org/security/kubernetes-network-policy)**:
k8s 内所有 pod 默认都是互联互通的，但是这不符合生产实践，所以需要 network policy 来 namespace-scoped 级别去网络隔离，比如不同 namespace 下的 pod 不可以网络互通。
k8s 只是定义了 network policy api，具体实现是由 cni network plugin 实现的，比如 calico 实现了 network policy 功能，控制 pod 的网络流量的流入 ingress 和流出 egress。
network policy api 定义内容：
* policy 是 namespace scoped
* policy 作用于可以使用 label selector 来过滤出来的 pod
* poliy rule 支持协议(TCP/UDP/SCTP)和端口指定
* policy rule 可以使用 namespace/cidr(ip 段)/pod 来定义流量的流入和流出 

**[calico network policy demo](https://docs.projectcalico.org/security/tutorials/kubernetes-policy-demo/kubernetes-demo)** 


### calico/kube-controllers image
包含几个controllers:
(1) policy controller: watches network policies and programs Calico policies.
watch api-server，读取 `NetworkPolicy events`，然后去 sync k8s network policy，写入到 datastore 里。
只有 datastore 是 etcd 时才有效。

(2) namespace controller: watches namespaces and programs Calico profiles.


(3)serviceaccount controller: watches service accounts and programs Calico profiles.


(4)workloadendpoint controller: watches for changes to pod labels and updates Calico workload endpoints.

只有 datastore 是 etcd 时才有效。

(5)node controller 
node controller 会 watch api-server，读取 `Node events`，用来更新有关Node的配置(比如: crud node)。必须通过 `ENABLED_CONTROLLERS` 环境变量显示开启。

