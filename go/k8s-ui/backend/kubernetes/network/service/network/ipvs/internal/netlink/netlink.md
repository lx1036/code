
**[[译] LINUX，NETLINK 和 GO – 第 1 部分：NETLINK](http://blog.studygolang.com/2017/07/linux-netlink-and-go-part-1-netlink/)**


# netlink 是什么？
**[Netlink](https://zh.wikipedia.org/wiki/Netlink)** 是一个 Linux 内核进程间通信机制，可实现用户空间进程与内核之间的通信，或多个用户空间进程通讯。 
Netlink 套接字是启用此通信的原语。也就是说，netlink 不仅仅可以实现用户空间两个进程通信，也可以实现用户空间和内核空间的两个进程通信。



## netlink 消息格式






# netlink go client
本客户端被用来 crud 网络设备interfaces, 设置IP地址和路由表set ip addresses and routes, and configure ipsec.

功能点：
* 添加一个新 bridge，并添加一个虚拟网卡，比如创建一个lx1036的bridge，然后把eth0网卡接入进去。



# [笔记]《k8s网络权威指南》1.3小节：Linux Bridge
```shell script
# 1. 创建一个 bridge 并启动
ip link add name br0 type bridge
ip link set br0 up
```

