
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

calicoctl:
```shell script
DATASTORE_TYPE=kubernetes KUBECONFIG=~/.kube/config calicoctl get nodes
```

## concepts
**[Calico on Kubernetes 从入门到精通](https://www.kubernetes.org.cn/4960.html)**


### calico in minikube
```shell script
minikube start --network-plugin=cni
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
# validate
curl $(minikube ip):$(kubectl get svc nginx-demo-1 -o=jsonpath='{.spec.ports[0].nodePort}') -v
```

### network policy
**[k8s network policy](https://docs.projectcalico.org/security/kubernetes-network-policy)**:

