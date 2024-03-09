
# 目的
验证 xdp ebpf 加速了通节点 pod 之间的网络包的转发过程。

# xdp veth

```md
/root/linux-5.10.142/tools/testing/selftests/bpf/test_xdp_veth.sh
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/xdp_redirect_map.c
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/xdp_tx.c
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/xdp_dummy.c
```

