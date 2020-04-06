
# kube-proxy
概念：每台机器上都运行一个 kube-proxy 服务，它监听 API server 中 service 和 endpoint 的变化情况，并通过 iptables 等来为服务配置负载均衡（仅支持 TCP 和 UDP）。



# K8S 中的负载均衡
由kube-proxy实现的Service是一个四层的负载均衡，Ingress是一个七层的负载均衡。一个Service有对应的ClusterIP，即VIP，但在集群里不存在，没法ping通。


