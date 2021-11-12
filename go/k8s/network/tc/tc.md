


# TC(Traffic Control)
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


## 参考文献
linux使用TC并借助ifb实现入向限速: https://blog.csdn.net/bestjie01/article/details/107404231
