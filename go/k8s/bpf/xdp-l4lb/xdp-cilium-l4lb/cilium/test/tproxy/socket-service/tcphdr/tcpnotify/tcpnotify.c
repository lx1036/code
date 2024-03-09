
//go:build ignore


// /root/linux-5.10.142/include/uapi/linux/bpf.h
#include <linux/bpf.h>
// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>


#define TESTPORT 12877

struct tcp_notifier {
    __u8    type;
    __u8    subtype;
    __u8    source;
    __u8    hash;
};
// 需要加这一行否则报错，见命令 `//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" -type tcp_notifier bpf tcpnotify.c -- -I.`
struct tcp_notifier *unused_event __attribute__((unused));

struct tcpnotify_globals {
    __u32 total_retrans;
    __u32 ncalls;
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    // noPrealloc flag may be incompatible with map type Array
    // __uint(map_flags, BPF_F_NO_PREALLOC);
    __uint(max_entries, 2);
    __type(key, __u32);
    __type(value, struct tcpnotify_globals);
} global_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(max_entries, 2);
    __uint(key_size, sizeof(int));
    __uint(value_size, sizeof(__u32));
} perf_event_map SEC(".maps");

// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tcpnotify_kern.c
// perf event retrans
SEC("sockops")
int bpf_sockops_cb(struct bpf_sock_ops *skops) {
    int rv = -1;

    if (bpf_ntohl(skops->remote_port) != TESTPORT) {
        skops->reply = -1;
        return 0;
    }

    int op = (int) skops->op;
    switch (op) {
    case BPF_SOCK_OPS_TIMEOUT_INIT:
    case BPF_SOCK_OPS_RWND_INIT:
    case BPF_SOCK_OPS_NEEDS_ECN:
    case BPF_SOCK_OPS_BASE_RTT:
    case BPF_SOCK_OPS_RTO_CB:
        rv = 1;
        break;

    case BPF_SOCK_OPS_TCP_CONNECT_CB:
    case BPF_SOCK_OPS_TCP_LISTEN_CB:
    case BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB:
    case BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB:
        bpf_sock_ops_cb_flags_set(skops, BPF_SOCK_OPS_RETRANS_CB_FLAG|BPF_SOCK_OPS_RTO_CB_FLAG);
        rv = 1;
        break;

    // 重传时的回调函数
    case BPF_SOCK_OPS_RETRANS_CB: {
        struct tcp_notifier msg = {
            .type = 0xde,
            .subtype = 0xad,
            .source = 0xbe,
            .hash = 0xef,
        };
        rv = 1;
        struct tcpnotify_globals *global, g;
        __u32 key = 0;
        global = bpf_map_lookup_elem(&global_map, &key);
        if (!global) {
            break;
        }
        g = *global;
        g.total_retrans = skops->total_retrans; // 重传总次数
        g.ncalls++;
        bpf_map_update_elem(&global_map, &key, &g, BPF_ANY);
        bpf_perf_event_output(skops, &perf_event_map, BPF_F_CURRENT_CPU, &msg, sizeof(msg));
    }
        break;

    default:
        rv = -1;
        break;
    }

    skops->reply = rv;
    return 1;
}


char _license[] SEC("license") = "GPL";

