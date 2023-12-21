
/**
 * struct sock 定义见 /root/linux-5.10.142/include/net/sock.h
 *
 */



#include <vmlinux.h>

#include <linux/bpf.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_endian.h>

//#include "tcp_connect.h"

#define TASK_COMM_LEN 16;
#define AF_INET 2
#define NULL    ((void *)0)

// 必须加 volatile，表示这个 const 可能在运行中会变的
const volatile __u32 targ_tgid = 0;
const volatile __u64 targ_min_us = 0;


struct event {
    union {
        __u32 saddr_v4;
        __u8 saddr_v6[16];
    };
    union {
        __u32 daddr_v4;
        __u8 daddr_v6[16];
    };
    char comm[TASK_COMM_LEN];
    __u64 delta_us;
    __u64 ts_us;
    __u32 tgid;
    int af;
    __u16 lport;
    __u16 dport;
};

// 需要加这一行否则报错，见命令 `//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" -type tcp_notifier bpf tcpnotify.c -- -I.`
struct event *unused_event __attribute__((unused));

struct piddata {
    char comm[TASK_COMM_LEN];
    __u64 ts;
    __u32 tgid;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 4096);
    __type(key, struct sock *);
    __type(value, struct piddata);
} start SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(__u32));
} events SEC(".maps");


// 获取当前进程的 name/tgid/ts
static int enter_tcp_connect(struct sock *sk) {
    // pid=bpf_get_current_pid_tgid(), tgid=bpf_get_current_pid_tgid() >> 32
    __u32 tgid = bpf_get_current_pid_tgid() >> 32;
    struct piddata piddata = {};

    if (targ_tgid && targ_tgid != tgid)
        return 0;

    // bpf_get_current_comm() 获取当前进程的 name，并存放在 piddata.comm 中
    bpf_get_current_comm(&piddata.comm, sizeof(piddata.comm));
    piddata.ts = bpf_ktime_get_ns();
    piddata.tgid = tgid;
    bpf_map_update_elem(&start, &sk, &piddata, 0);
    return 0;
}

static int handle_tcp_rcv_state_process(void *ctx, struct sock *sk) {
    struct piddata *piddatap;
    __u64 ts;
    s64 delta;
    struct event event = {};

    if (sk->__sk_common.skc_state != TCP_SYN_SENT) {
        return 0;
    }
    if (sk->__sk_common.skc_family != AF_INET) { // only ipv4
        return 0;
    }

    piddatap = bpf_map_lookup_elem(&start, &sk);
    if (!piddatap)
        return 0;

    ts = bpf_ktime_get_ns();
    delta = (s64)(ts - piddatap->ts);
    if (delta < 0)
        goto cleanup;

    event.delta_us = delta / 1000U;
    if (targ_min_us && event.delta_us < targ_min_us)
        goto cleanup;

    __builtin_memcpy(&event.comm, piddatap->comm, sizeof(event.comm));
    event.ts_us = ts / 1000;
    event.tgid = piddatap->tgid;
    event.af = sk->__sk_common.skc_family;
    event.lport = sk->__sk_common.skc_num;
//    event.lport = BPF_CORE_READ(sk, __sk_common.skc_num);
    event.dport = sk->__sk_common.skc_dport;
    event.saddr_v4 = sk->__sk_common.skc_rcv_saddr;
    event.daddr_v4 = sk->__sk_common.skc_daddr;
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &event, sizeof(event));

cleanup:
    bpf_map_delete_elem(&start, &sk);
    return 0;
}

// 用户态调用 connect() 发送 syn 包时，内核里会调用 tcp_v4_connect() 函数，这里进行 bpf hook
// /root/linux-5.10.142/net/ipv4/tcp_ipv4.c::tcp_v4_connect(struct sock *sk, struct sockaddr *uaddr, int addr_len)
SEC("kprobe/tcp_v4_connect")
int tcp_v4_connect(struct sock *sk) {
    return enter_tcp_connect(sk);
}

//SEC("kretprobe/tcp_v4_connect")
//int tcp_v4_connect(struct sock *sk) {
//    return exit_tcp_connect(ctx, ret, 4);
//}

// 在 socket 状态变化时，内核调用 tcp_rcv_state_process() 函数
// /root/linux-5.10.142/net/ipv4/tcp_input.c::tcp_rcv_state_process(struct sock *sk, struct sk_buff *skb)
SEC("kprobe/tcp_rcv_state_process")
int tcp_rcv_state_process(struct sock *sk) {
    return handle_tcp_rcv_state_process(ctx, sk);
}

// 计算 rtt
// /root/linux-5.10.142/net/ipv4/tcp_input.c::tcp_rcv_established(struct sock *sk, struct sk_buff *skb)
// https://github.com/eunomia-bpf/bpf-developer-tutorial/blob/main/src/14-tcpstates/tcprtt.bpf.c
SEC("kprobe/tcp_rcv_established")
int tcp_rcv_established(struct sock *sk) {

}


char _license[] SEC("license") = "GPL";

