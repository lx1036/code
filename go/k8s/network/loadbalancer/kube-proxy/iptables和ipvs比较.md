

# iptables/ipvs
https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/ipvs/README.md

做四层负载均衡，ipvs 相比于 iptables 优点：
(1)ipvs 相比于 iptables 性能更好，尤其 service 数量越来越多时，iptables rule 用的链表存储，而 ipvs 用的哈希表map来存储。
(2)ipvs 支持的负载均衡算法很多，比如 weight rr轮询，hash，least connection/load，kube-proxy 使用 ipvs 负载均衡算法是可以配置的，默认是 rr 轮询；
而 kube-proxy 使用 iptables statistic module 实现的算法只有轮询(随机)。
(3)ipvs 支持 rs health check，和 connection retry。这个还挺重要的。

