


# TC(Traffic Control)
基本概念 Components of Linux Traffic Control 链接：
https://tldp.org/HOWTO/Traffic-Control-HOWTO/
https://tldp.org/HOWTO/Traffic-Control-HOWTO/components.html
https://tldp.org/HOWTO/Traffic-Control-HOWTO/software.html

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

```shell
# 测试流量出方向
tc qdisc del dev eth0 root # 清除 eth0 上所有队列

tc class show dev eth0 parent 1:
tc qdisc show

tc qdisc add dev eth0 root handle 1: htb default 1
tc class add dev eth0 parent 1: classid 1:1 htb rate 1Gbit
tc class add dev eth0 parent 1: classid 1:2 htb rate 1Gbit
tc class add dev eth0 parent 1:2 classid 1:3 htb rate 500Mbit ceil 1Gbit prio 3
tc class add dev eth0 parent 1:2 classid 1:5 htb rate 300Mbit ceil 1Gbit prio 5
tc class add dev eth0 parent 1:2 classid 1:7 htb rate 200Mbit ceil 1Gbit prio 7

# 将cgroup与物理网卡的qdisc绑定
# https://android.googlesource.com/kernel/common/+/bcmdhd-3.10/Documentation/cgroups/net_cls.txt
# The Traffic Controller (tc) can be used to assign different priorities to packets from different cgroups
tc filter add dev eth0 parent 1: protocol ip handle 1: cgroup

# 创建高优先级cgroup组high
mkdir /sys/fs/cgroup/net_cls/high
# 设定high组的classid 1:3
# classid的设定为16进制设定，前4位:后4位表示，1:3写为0x00010003，省略前置0后为0x10003
echo 0x10003 > /sys/fs/cgroup/net_cls/high/net_cls.classid
# 创建高优先级cgroup组low
mkdir /sys/fs/cgroup/net_cls/low
# 设定low组的classid 1:7
echo 0x10007 > /sys/fs/cgroup/net_cls/low/net_cls.classid

# 在 nodeIP 100.211.55.3 里启动网络服务端
iperf3 -s -p 5000 & iperf3 -s -p 5001 &
# 在客户端机器上开启两个terminal，然后压测网络
iperf3 -c 100.211.55.3 -p 5000 --bandwidth 10G -t 1000
iperf3 -c 100.211.55.3 -p 5001 --bandwidth 10G -t 1000

# 测试流量入方向
# https://serverfault.com/questions/350023/tc-ingress-policing-and-ifb-mirroring
modprobe ifb numifbs=1
ip link set dev ifb0 up # repeat for ifb1, ifb2, ...

# And redirect ingress traffic from the physical interfaces to corresponding ifb interface. For eth0 -> ifb0:
tc qdisc add dev eth0 handle ffff: ingress
tc filter add dev eth0 parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0

# Egress rules for eth0 go as usual in eth0. Let's limit bandwidth, for example:
tc qdisc add dev eth0 root handle 1: htb default 10
tc class add dev eth0 parent 1: classid 1:1 htb rate 1mbit
tc class add dev eth0 parent 1:1 classid 1:10 htb rate 1mbit

# Ingress rules for eth0, now go as egress rules on ifb0
tc qdisc add dev ifb0 root handle 1: htb default 10
tc class add dev ifb0 parent 1: classid 1:1 htb rate 1mbit
tc class add dev ifb0 parent 1:1 classid 1:10 htb rate 1mbit

```

```shell
# 创建 tc 重定向两个网卡流量
# create tc eth0<->tap0 redirect rules
tc qdisc add dev eth0 ingress
tc filter add dev eth0 parent ffff: protocol all u32 match u8 0 0 action mirred egress redirect dev tap1

tc qdisc add dev tap0 ingress
tc filter add dev tap0 parent ffff: protocol all u32 match u8 0 0 action mirred egress redirect dev eth1

# 使用 tc 加载 ebpf 程序到网卡上，tc 命令一般用来下发 tc bpf 程序
tc filter add dev veth09e1d2e egress bpf da obj tc-xdp-drop-tcp.o sec tc-test
# ip 命令一般用来下发 xdp bpf 程序
ip link set dev veth09e1d2e tc obj tc-xdp-drop-tcp.o sec tc-test


# 加载 ebpf 程序到 xdp 上
# 但是 Cilium 默认没有下发 xdp ebpf 程序
ip link set dev [network-device-name] xdp obj xdp_drop_all.o sec xdp
ip link set dev [network-device-name] xdp off # 卸载 xdp ebpf 程序
# 由于机器上没有最新的 iproute 和 glibc，只有 cilium pod namespace 里有，还得需要把 xdp.o 拷贝过去然后使用最新的 `ip` 命令
docker cp ./xdp.o 64d6796758b4:/mnt/xdp.o
nsenter -t 26416 -m -n ip -force link set dev lxc6e7eb5daff06 xdp obj /mnt/xdp.o sec xdp
nsenter -t 26416 -m -n ip -force link set dev lxc6e7eb5daff06 xdp off

# 抓包 tcp 且 80 端口
tcpdump -i lxc6e7eb5daff06 -nn tcp and port 80
```


(1) 如何给 network packet 打标签？如何设置优先级？
* 打标签：Network classifier cgroup, https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v1/net_cls.html
* 设置优先级：https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v1/net_prio.html



## 参考文献
linux使用TC并借助ifb实现入向限速: https://blog.csdn.net/bestjie01/article/details/107404231

HTB实现原理：http://luxik.cdi.cz/~devik/qos/htb/manual/theory.htm

tc man手册： https://man7.org/linux/man-pages/man8/tc.8.html

cgroup和tc结合设置文档参考：https://android.googlesource.com/kernel/common/+/bcmdhd-3.10/Documentation/cgroups/net_cls.txt

《Linux 高级路由与流量控制手册（2012）》第九章 用 tc qdisc 管理 Linux 网络带宽: http://arthurchiao.art/blog/lartc-qdisc-zh/
