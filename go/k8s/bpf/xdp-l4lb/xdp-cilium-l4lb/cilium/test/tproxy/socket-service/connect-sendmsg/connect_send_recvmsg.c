

/**
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/connect4_prog.c
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/sendmsg4_prog.c
 * /root/linux-5.10.142/tools/testing/selftests/bpf/test_sock_addr.c
 */

#include <sys/socket.h>

#include <linux/bpf.h>
#include <linux/in.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

SEC("cgroup/connect4")
int connect_v4_prog(struct bpf_sock_addr *ctx) {

}


SEC("cgroup/sendmsg4")
int sendmsg_v4_prog(struct bpf_sock_addr *ctx) {

}


char _license[] SEC("license") = "GPL";
