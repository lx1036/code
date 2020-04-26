
## **[Docker 网络基础](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720c8561c92c7091bd4ddd)**
Docker 网络三种类型：
(1) None: Null 网络驱动，没有网络。
```shell script
docker run --rm --network none -it alpine sh
ifconfig
hostname -i
```

(2) Host:
```shell script
docker run --rm --network host -it alpine sh
ifconfig
ls --time-style="+" -l  /proc/`docker inspect $(docker ps -ql) --format "{{ .State.Pid }}"`/ns | grep net
```

(3) Bridge:
```shell script
docker run --rm --network bridge -it alpine sh
ifconfig
ip r
ifconfig bridge0 # 宿主机上该网桥
```

(4) container 网络：就是一个 container 网络加入另一个 container 的网络，共享一个 network。K8S 的 Pod 就是这么做的。
```shell script
docker network create container-network
docker run --rm -d --network container-network redis
docker run --rm -it --network container-network alpine sh
ps -ef
netstat  -ntlp
docker info --format "{{ .Plugins.Network }}" # Docker 支持的网络驱动列表
# [bridge host ipvlan macvlan null overlay]
```

docker --link 并不是公用网络堆栈，只是在 `/etc/hosts` 里增加了一条记录:
```shell script
docker run --rm -it --name redis-network -d  redis
docker run --rm -it --name alpine-network --link redis-network alpine sh
```

**[定制 bridge](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720adc61c92c7091bd4dcc)**
**[容器网络的灵活使用](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720b0f61c92c7091bd4dcd)**

```shell script
docker network create --help
# CIDR(Classless Inter-Domain Routing): CIDR主要是一个按位的、基于前缀的，用于解释IP地址的标准。
# --subnet: Subnet in CIDR format that represents a network segment
docker network create --gateway 192.168.31.1 --subnet 192.168.31.1/24 custom-network
docker run --rm -d --network custom-network --name redis-custom-network redis
hostname -i # 192.168.31.2

# 把运行中的容器加入已有的网络
# (1)
docker run -d --name redis-default-network redis
# (2)
docker run --rm -it --network custom-network --name connect-alpine-custom-network alpine sh
ping -c 1 redis-default-network
# (3)
docker network connect/disconnect custom-network redis-default-network
ping -c 1 redis-default-network
```

## **[Docker 与 iptables 之间的联系](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720b7a61c92c7091bd4dd4)**


## **[手动进行容器网络的管理](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720bd161c92c7091bd4dd6)**


## **[docker-proxy 的原理](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720be061c92c7091bd4dd7)**


## **[Docker 内部 DNS 原理](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720c6061c92c7091bd4ddb)**


## **[Docker 网络原理](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720c8561c92c7091bd4ddd)**


## **[Docker 与 Kubernetes](https://gitbook.cn/gitchat/column/5d70cfdc4dc213091bfca46f/topic/5d720ca461c92c7091bd4ddf)**




# iptables

![iptables 工作流程图](./img/tables_traverse.jpg)


# Docker 内置的 DNS 服务

