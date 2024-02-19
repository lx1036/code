# socket reuseport

> 注意：和 sk_lookup 里 sk_reuseport 中，构造 client/server 的对比，这里用的是创建一组 sockets，
> 然后使用 epoll 函数监听该所有 sockets 的网络事件

代码在:

```md
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_select_reuseport_kern.c
/root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/select_reuseport.c

```





