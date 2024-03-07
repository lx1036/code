# socket acceleration
ebpf for TCP/IP bypass.
ebpf 挂载的是 cgroup，所以服务 application 只能是 pod 内的，尤其对于 pod 内的一对 gRPC 服务，可以 bypass TCP/IP 来加速。


```md
https://github.com/ArthurChiao/socket-acceleration-with-ebpf
https://github.com/cyralinc/os-eBPF/blob/develop/sockredir/README.md
https://cyral.com/blog/how-to-ebpf-accelerating-cloud-native/
https://cyral.com/blog/lessons-using-ebpf-accelerating-cloud-native/

/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_sockmap_listen.c
/root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/sockmap_listen.c
/root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/sockmap_basic.c
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_sockmap_invalid_update.c
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_sockmap_update.c
```

bpf program type:
* BPF_PROG_TYPE_SOCK_OPS
* BPF_PROG_TYPE_SK_MSG
  * hook 点: sendmsg call on a socket
  * BPF_MAP_TYPE_SOCKMAP(These maps are key value stores where the value can only be a socket.)
  * BPF_MAP_TYPE_SOCKHASH


# 验证测试

```shell
# We can use a TCP listener spawned by SOCAT to mimic an echo server, and netcat to sent a TCP connection request.
sudo socat TCP4-LISTEN:1000,fork exec:cat
nc localhost 1000 # this should produce the trace in the kernel file trace_pipe
```

