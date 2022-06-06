

# host local cluster wide ipam
解决的问题：host-local ipam 只会根据一个 node 一个 pod cidr，然后从当前的 pod cidr 分配一个 ip。


**[whereabouts](https://github.com/k8snetworkplumbingwg/whereabouts)** : 
主要通过 k8s leader election 获取中心式/分布式锁来进行 ip 分配，可以创建多个 ippool，然后 pod 选择对应的 ippool，然后分配出对应的 ip
给 pod，并去更新其 ippool 已使用的 ip 地址列表。

```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: sriov-net2
  annotations:
    k8s.v1.cni.cncf.io/resourceName: intel.com/intel_sriov_netdevice
spec:
  config: '{
    "cniVersion":"0.3.1",
    "name":"nsnetwork-sample",
    "plugins":[
        {
            "ipam":{
                "gateway":"192.168.6.254",
                "range":"192.168.6.0/24",
                "type":"whereabouts" # 使用whereabouts
            },
            "type":"sriov"
        },
        {
            "type":"sbr"
        }
    ]
}
```

