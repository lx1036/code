


# TC(Traffic Controller)
TC工作原理：使用classful的qdisc(排队规则queueing discipline)，通过tc对流量进行控制，使用HTB算法实现带宽优先级和抢占控制。
使用tc中的classful队列规定（qdisc）进行流量限制，涉及tc的几个基本概念：
* qdisc：队列，流量根据Filter的计算后会放入队列中，然后根据队列的算法和规则进行发送数据
* class：类，用来对流量进行处理，可以进行限速和优先级设置，每个类中包含了一个隐含的子qdisc，默认的是pfifo队列
* filter：过滤器，用于对流量进行分类，放到不同的qdisc或class中
* 队列算法HTB：实现了一个丰富的连接共享类别体系。使用HTB可以很容易地保证每个类别的带宽，虽然它也允许特定的类可以突破带宽上限，占用别的类的带宽。

出流量限制: 通过cgroup对不通的pod设定不同的classid，进入不同的队列，实现优先级划分和网络流量限制
入流量限制: 通过增加ifb设备，将物理网卡流量转发到ifb设备，在ifb设备的入方向使用tc进行限制，限制使用filter对destip进行分类，不同的ip对应的pod的优先级决定入何种优先级的队列


## TC CNI Plugin
bandwidth: https://www.cni.dev/plugins/current/meta/bandwidth/
IFB(Intermediate Functional Block): 和tun一样，ifb也是一个虚拟网卡


```shell
# 加载 ifb 驱动并创建一个 ifb 虚拟网卡，然后 up 网卡
modprobe ifb numifbs=1
ip link set dev ifb0 up

# 清除原有的根队列(根据实际情况操作,非必要)
tc qdisc del dev eth0 root 2>/dev/null
tc qdisc del dev eth0 ingress 2>/dev/null
tc qdisc del dev ifb0 root 2>/dev/null

# 将eth0的ingress流量全部重定向到 ifb0 处理
tc qdisc add dev eth0 handle ffff: ingress
tc filter add dev eth0 parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0

# eth0的出向限速:eth0添加根队列,使用htb,添加1:1类,使用htb 
tc qdisc add dev eth0 root handle 1: htb r2q 625 default 65
tc class add dev eth0 parent 1: classid 1:1 htb rate 1000Mbit

# eth0的入向限速:ifb0添加根队列,使用htb,添加1:1类,使用htb 
tc qdisc add dev ifb0 root handle 1: htb r2q 625 default 65
tc class add dev ifb0 parent 1: classid 1:1 htb rate 1000Mbit

# eth0的出向限速:eth0设备添加子类\对应的filter配置规则和子类的队列
tc class add dev eth0 parent 1:1 classid 1:10 htb rate 10Mbit
tc filter add dev eth0 parent 1: protocol all prio 1 u32 match ip dst 192.168.0.2 classid 1:10
tc qdisc add dev eth0 parent 1:10 handle 10: sfq

# eth0的出向限速:eth0设备添加子类\对应的filter配置规则和子类的队列
tc class add dev eth0 parent 1:1 classid 1:11 htb rate 20Mbit
tc filter add dev eth0 parent 1: protocol all prio 1 u32 match ip dst 192.168.0.3 classid 1:11
tc qdisc add dev eth0 parent 1:11 handle 11: sfq

# eth0的入向限速:ifb0设备添加子类\对应的filter配置规则和子类的队列
tc class add dev ifb0 parent 1:1 classid 1:10 htb rate 10Mbit
tc filter add dev ifb0 parent 1: protocol all prio 1 u32 match ip src 192.168.0.2 classid 1:10
tc qdisc add dev ifb0 parent 1:10 handle 10: sfq

# eth0的入向限速:ifb0设备添加子类\对应的filter配置规则和子类的队列
tc class add dev ifb0 parent 1:1 classid 1:11 htb rate 20Mbit
tc filter add dev ifb0 parent 1: protocol all prio 1 u32 match ip src 192.168.0.3 classid 1:11
tc qdisc add dev ifb0 parent 1:11 handle 11: sfq

```



## 参考文献
linux使用TC并借助ifb实现入向限速: https://blog.csdn.net/bestjie01/article/details/107404231
HTB实现原理：http://luxik.cdi.cz/~devik/qos/htb/manual/theory.htm
tc man手册： https://man7.org/linux/man-pages/man8/tc.8.html
cgroup和tc结合设置文档参考：https://android.googlesource.com/kernel/common/+/bcmdhd-3.10/Documentation/cgroups/net_cls.txt