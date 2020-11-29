
# Installation





# Docker Container 是什么？
Docker Container 是一组进程以及它们所能访问资源的集合，使用 Namespace/CGroups 技术实现资源隔离和资源访问控制。
如果不用Docker，Linux 默认启动的进程会在默认名为 'root' 的 Namespace/CGroups，而 Docker 启动的 abc container 会在
它自己的 Namespace/CGroups。最后，在宿主机上会形成 'root', 'abc', '123' 这里一个个资源独立的容器。
一篇不错的文章：**[容器概述](https://segmentfault.com/a/1190000006908063)**

# 交换机 Switch
一个有着多个端口的网络设备，交换机里多一张路由表(mac 地址和交换机端口的映射表)，如：

| mac地址 | 端口 |
| :--- | :---: |
| 02:42:83:06:75:13 | 2 |
| 08:00:27:03:d0:e7 | 2 |
| ee:35:41:bb:a4:60 | 3 |
| 02:42:34:8F:0E:FE | 4 |

> 这里02:42:83:06:75:13和08:00:27:03:d0:e7都与端口2相连，表示与端口2连接的是一个交换机或者有多个虚拟网卡的主机。

交换机在刚启动时，这张表是空的，当收到第一个数据包的时候，它也不知道要从哪个端口转发出去，于是它采用和集线器一样的方式广播出去。
当交换机每次从一个端口收到数据包时，都会提取数据包里面的源mac地址，然后将这个mac地址和端口的对应关系添加到（或者更新）转发表，
这样很快就会将转发表构造起来，就算有网线换了端口，也会及时的更新转发表。


**[网络为什么要分层](https://segmentfault.com/a/1190000008741770)**


**[Docker 镜像构建原理及源码分析](https://gitbook.cn/books/5d0b4be966a9e7233095d290/index.html)**
**[containerd(docker精简版)](https://containerd.io/)**


# TODO
**[Docker 核心知识必知必会(深入底层解读 Docker 核心技术)](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f)**
**[自己动手写Docker](http://www.duokan.com/reader/www/app.html?id=af432a1b21c645b09fcae2581d340c76)**
**[xianlubird/mydocker](https://github.com/xianlubird/mydocker)**
**[Kubernetes 从上手到实践](https://juejin.im/book/5b9b2dc86fb9a05d0f16c8ac)**



# 知识树



## Docker Plugin
(1) golang写一个plugin，然后做成docker image，再通过docker plugin install去拉取镜像并enable the plugin。

