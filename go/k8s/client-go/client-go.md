
# Code
**[client-go](https://github.com/kubernetes/client-go)**



**[install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)**
**[install minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)**
**[minikube docs](https://minikube.sigs.k8s.io/)**

```shell script
minikube start
minikube stop
```

**[client-go demos](https://github.com/kubernetes/client-go/blob/master/examples/README.md)**


**[Kubernetes Client-Go Informer 实现源码剖析](https://xigang.github.io/2019/09/21/client-go/)**
**[深入浅出kubernetes之client-go的Indexer](https://blog.csdn.net/weixin_42663840/article/details/81530606)**


![client-go-architecture](./imgs/client-go-architecture.jpg)


# List and Watch
**[理解 K8S 的设计精髓之 List-Watch机制和Informer模块](https://zhuanlan.zhihu.com/p/59660536)**:
使用HTTP长连接来实现异步通信(为啥不用websocket)：api-server会把crud的资源打包为事件(数据持久化到etcd里)，发给各个客户端。各个客户端使用ListAndWatch来订阅这些事件，而不是轮询。
订阅事件就是通过HTTP长连接来实现异步通信的，下面是demo演示：
```shell script
kubectl proxy # 开启个不需要验证的api-server proxy
curl -i localhost:8001/api/v1/watch/namespaces/default/pods # watch pods资源，作为客户端Curl
# 创建pod资源，会restful api调用api-server，创建pod对象数据存入etcd，同时api-server会把这些crud事件发给客户端Curl、
# controller-manager(pod controller)、kube-scheduler(去调度这个对象给node上的kubelet)、kubelet来创建这个pod。
kubectl apply -f go/k8s/network/nginx/minikube-nginx.yml

HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 28 Jun 2020 14:03:58 GMT
Transfer-Encoding: chunked

{"type":"ADDED","object":{"kind":"Pod","apiVersion":"v1","metadata":{"name":"minikube-production-68df7c4898-pxhn9","generateName":"minikube-production-68df7c4898-","namespace":"default","selfLink":"/api/v1/namespaces/default/pods/minikube-production-68df7c4898-pxhn9","uid":"7ca67c2b-066d-4413-85e7-e9c72ea60445","resourceVersion":"237193"...}
{"type":"MODIFIED","object":{"kind":"Pod","apiVersion":"v1","metadata":{"name":"nginx-demo-1-7f67f8bdd8-jc2zr","generateName":"nginx-demo-1-7f67f8bdd8-","namespace":"default","selfLink":"/api/v1/namespaces/default/pods/nginx-demo-1-7f67f8bdd8-jc2zr","uid":"32fd5db3-2f26-47de-87a3-f77f2bfba085","resourceVersion":"239251"...}
```
