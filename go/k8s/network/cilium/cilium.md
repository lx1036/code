


## Cilium
Cilium CNI 创建 pod network 具体流程原理：https://arthurchiao.art/blog/cilium-code-cni-create-network/


```

types.LoadNetConf(args.StdinData) -> connector.SetupVeth(ep.ContainerID, int(conf.DeviceMTU), ep)
-> netlink.LinkSetNsFd(*peer, int(netNs.Fd())) -> connector.SetupVethRemoteNs(netNs, tmpIfName, args.IfName)
-> c.IPAMAllocate("", podName, true) -> c.Ipam.PostIpam(params) 

-> ipam:*models.IPAMResponse

-> configureIface(ipam, args.IfName, &state) -> c.EndpointCreate(ep) -> c.Endpoint.PutEndpointID(params)

```





## Cilium 功能点
Cilium 通过 BGP 宣告 pod cidr：https://github.com/cilium/cilium/pull/16525
Cilium 多个 ippool 问题，目前还不支持，一个K8s里部署多个Cilium不同网段实例，貌似也不支持：https://github.com/cilium/cilium/issues/13227

本地安装 cilium CLI:
```shell
curl -L --remote-name-all https://github.com/cilium/cilium-cli/releases/latest/download/cilium-darwin-amd64.tar.gz{,.sha256sum}
shasum -a 256 -c cilium-darwin-amd64.tar.gz.sha256sum
sudo tar xzvfC cilium-darwin-amd64.tar.gz /usr/local/bin
rm cilium-darwin-amd64.tar.gz{,.sha256sum}
```


## Troubleshoot
(1)Cilium 支持多个网段问题，或者k8s 里部署多个不同网段的 Cilium实例？


# 笔记
(1) 根据 containerID 获取 PID
```shell
docker inspect ${containerID} | jq '.[] | .State | .Pid' # 27219
docker inspect ${containerID} | grep "Pid" # 27219

# nsenter -t <pid> -n <command>
nsenter -t 27219 -n
ip addr # 获得 Cilium 创建的 veth peer 在 container side 一侧
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
749: eth0@if750: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP
    link/ether 22:72:9c:50:45:e1 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.216.136.178/32 scope global eth0
       valid_lft forever preferred_lft forever
```


## 参考文献
**[Cilium Code Walk Through: CNI Create Network](https://arthurchiao.art/blog/cilium-code-cni-create-network/)**

**[Life of a Packet in Cilium：实地探索 Pod-to-Service 转发路径及 BPF 处理逻辑](https://arthurchiao.art/blog/cilium-life-of-a-packet-pod-to-service-zh/)**

**[Cilium Code Walk Through Series](http://arthurchiao.art/blog/cilium-code-series/)**

**[L4LB for Kubernetes: Theory and Practice with Cilium+BGP+ECMP](http://arthurchiao.art/blog/k8s-l4lb/)**
