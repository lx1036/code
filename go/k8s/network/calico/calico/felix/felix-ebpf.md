

# BPF
calico felix 组件负责下发 eBPF 规则，并实现 service 四层负载均衡。
开启 eBPF https://projectcalico.docs.tigera.io/maintenance/ebpf/enabling-ebpf :

```yaml
apiVersion: crd.projectcalico.org/v1
kind: FelixConfiguration
metadata:
  name: default
spec:
  bpfEnabled: true # 设置为 true 则开启
  bpfExternalServiceMode: DSR # 负载均衡 DSR mode
  bpfLogLevel: ''
  logSeverityScreen: Info
  reportingInterval: 0s
```

开启后的 ebpf maps 对象为：

```shell
ls /sys/fs/bpf/tc/globals
# cali_v4_arp2  cali_v4_ct2  cali_v4_ct_nats  cali_v4_fsafes2  
# cali_v4_ip_sets  cali_v4_nat_aff  cali_v4_nat_be  
# cali_v4_nat_fe3  cali_v4_routes  cali_v4_srmsg  cali_v4_state3
```

