

**[ipvs go client](https://github.com/moby/ipvs)**
**[ipvs k8s go client](https://github.com/kubernetes/kubernetes/blob/master/pkg/util/ipvs/ipvs_linux.go)**


# IPVS
IPVS 基于 netfilter 的散列表，比 iptables 性能更好。支持 TCP、UDP、SCTP、IPV4 和 IPV6 等协议，负载均衡策略支持:
IPVS支持十种负载均衡调度算法：
> 轮叫rr（Round Robin）：轮询模式，每一个rs按照均衡比例接收请求。以轮叫的方式依次将请求调度到不同的服务器，会略过权值是0的服务器。
以轮叫的方式依次将请求调度到不同的服务器，会略过权值是0的服务器。

> 加权轮叫wrr（Weighted Round Robin）：有权重的轮询模式，可以给rs设置权重。按权值的高低和轮叫方式分配请求到各服务器。服务器的缺省权值为1。假设服务器A的权值为1，B的权值为2，则表示服务器B的处理性能是A的两倍。
例如，有三个服务器A、B和C分别有权值4、3和2，则在一个调度周期内(mod sum(W(Si)))调度序列为AABABCABC。

> 最少链接lc（Least Connections）：把新的连接请求分配到当前连接数最小的服务器。

> 加权最少链接wlc（Weighted Least Connections）：调度新连接时尽可能使服务器的已建立连接数和其权值成比例，算法的实现是比较连接数与加权值的乘积，因为除法所需的CPU周期比乘法多，且在Linux内核中不允许浮点除法。

> 基于局部性的最少链接lblc（Locality-Based Least Connections）：主要用于Cache集群系统，将相同目标IP地址的请求调度到同一台服务器，来提高各台服务器的访问局部性和主存Cache命中率。
LBLC调度算法先根据请求的目标IP地址找出该目标IP地址最近使用的服务器，若该服务器是可用的且没有超载，将请求发送到该服务器；若服务器不存在，或者该服务器超载且有服务器处于其一半的工作负载，则用“最少链接”的原则选出一个可用的服务器，将请求发送到该服务器。

> 带复制的基于局部性最少链接（Locality-Based Least Connections with Replication）：主要用于Cache集群系统，它与LBLC算法的不同之处是它要维护从一个目标IP地址到一组服务器的映射。
LBLCR算法先根据请求的目标IP地址找出该目标IP地址对应的服务器组；按“最小连接”原则从该服务器组中选出一台服务器，若服务器没有超载，将请求发送到该服务器；若服务器超载；则按“最小连接”原则从整个集群中选出一台服务器，将该服务器加入到服务器组中，将请求发送到该服务器。
同时，当该服务器组有一段时间没有被修改，将最忙的服务器从服务器组中删除，以降低复制的程度。

> 目标地址散列dh（Destination Hashing）：通过一个散列（Hash）函数将一个目标IP地址映射到一台服务器，若该服务器是可用的且未超载，将请求发送到该服务器，否则返回空。使用素数乘法Hash函数：(dest_ip* 2654435761UL) & HASH_TAB_MASK。

> 源地址散列sh（Source Hashing）：根据请求的源IP地址，作为散列键（Hash Key）从静态分配的散列表找出对应的服务器，若该服务器是可用的且未超载，将请求发送到该服务器，否则返回空。

> 最短期望延迟sed（Shortest Expected Delay Scheduling）：将请求调度到有最短期望延迟的服务器。最短期望延迟的计算公式为(连接数 + 1) / 加权值。

> 最少队列调度（Never Queue Scheduling）：如果有服务器的连接数是0，直接调度到该服务器，否则使用上边的SEDS算法进行调度。

支持三种负载均衡模式：

kube-proxy 与 iptables
```shell script
kube-proxy --proxy-mode=ipvs --ipvs-scheduler=rr
```

查看linux主机上ipvs的内核模块：
```shell script
lsmod | grep ipvs
```
如果内核没有打开ipvs内核模块，

### ipvsadm
ipvsadm是IPVS的命令行管理工具，安装ipvs命令行客户端来操作内核中ipvs等几个模块：
```shell script
apt install -y ipvsadm
```
