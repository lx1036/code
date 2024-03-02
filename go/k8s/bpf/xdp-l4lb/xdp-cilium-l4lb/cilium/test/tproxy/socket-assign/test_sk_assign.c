

#include <stddef.h>
#include <stdbool.h>
#include <string.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <sys/socket.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>


/**
 * 注意：struct bpf_sk_lookup 的 bpf_sk_assign() 和 struct __sk_buff 的 bpf_sk_assign() 函数签名不一样，不是一个函数。
 */


#ifndef __section
#define __section(X) __attribute__((section(X), used))
#endif
#ifndef __section_maps
#define __section_maps __section("maps")
#endif
/* Pin map under /sys/fs/bpf/tc/globals/<map name> */
#define PIN_GLOBAL_NS 2

static const __u16 DST_PORT = 4321;

struct {
    __uint(type, BPF_MAP_TYPE_SOCKMAP);
    __uint(max_entries, 1);
    __type(key, int);
    __type(value, __u64);
//    __uint(pinning, LIBBPF_PIN_BY_NAME); // // 设置 LIBBPF_PIN_BY_NAME 导致 pin path 为 /sys/fs/bpf/tc/424e22c8e74276a6484f398886d426f441d9b849/
} server_map SEC(".maps");


static __always_inline int handle_udp(struct __sk_buff *skb) {
    int ret = 0;

    void *data_end = (void *) (long) skb->data_end;
    void *data = (void *) (long) skb->data;
    struct iphdr *iph = (struct iphdr *) (data + sizeof(struct ethhdr));
    struct bpf_sock_tuple *tuple;
    tuple->ipv4.saddr = iph->saddr;
    tuple->ipv4.daddr = iph->daddr;
    struct udphdr *udph;
    udph = (struct udphdr *) (iph + 1);
    if ((void *) (udph + 1) > data_end) {
        return TC_ACT_SHOT;
    }
    tuple->ipv4.sport = udph->source;
    tuple->ipv4.dport = udph->dest;
    if (udph->dest != bpf_htons(DST_PORT)) {
        // 2. 目标端口任意(不是 4321)，但是 socket 是 listen 状态，交给 netfilter 自己处理
        return TC_ACT_OK;
    }
    size_t tuple_len;
    tuple_len = sizeof(tuple->ipv4);
    if ((void *) (tuple + tuple_len) > data_end) {
        return TC_ACT_SHOT;
    }

    struct bpf_sock *sk;
    sk = bpf_sk_lookup_udp(skb, tuple, tuple_len, BPF_F_CURRENT_NETNS, 0);
    if (sk) {
        goto assign;
    }

    // 3. 目标端口是 4321，查找 server_map 得到 port 1234 的 server fd
    int zero = 0;
    sk = bpf_map_lookup_elem(&server_map, &zero);
    if (!sk) {
        return TC_ACT_SHOT;
    }

    assign:
    ret = bpf_sk_assign(skb, sk, 0);
    bpf_sk_release(sk);
    return ret;
}

static __always_inline int handle_tcp(struct __sk_buff *skb) {
    int ret = 0;

    void *data_end = (void *) (long) skb->data_end;
    void *data = (void *) (long) skb->data;
    struct iphdr *iph = (struct iphdr *) (data + sizeof(struct ethhdr));
    struct bpf_sock_tuple *tuple;
    tuple->ipv4.saddr = iph->saddr;
    tuple->ipv4.daddr = iph->daddr;
    struct tcphdr *tcph;
    tcph = (struct tcphdr *) (iph + 1);
    if ((void *) (tcph + 1) > data_end) {
        return TC_ACT_SHOT;
    }
    tuple->ipv4.sport = tcph->source;
    tuple->ipv4.dport = tcph->dest;
    if (tcph->dest != bpf_htons(4321)) {
        bpf_printk("not dst port %d", tcph->dest);
        // 2. 目标端口任意(不是 4321)，但是 socket 是 listen 状态，交给 netfilter 自己处理
        return TC_ACT_OK;
    }

    bpf_printk("handle_tcp");

    size_t tuple_len;
    tuple_len = sizeof(tuple->ipv4);
    if ((void *) (tuple + tuple_len) > data_end) {
        return TC_ACT_SHOT;
    }

    struct bpf_sock *sk;
    sk = bpf_skc_lookup_tcp(skb, tuple, tuple_len, BPF_F_CURRENT_NETNS, 0);
    if (sk) {
        // 1. 目标端口任意，获取到 socket 不是 listen 状态，说明已经是 establish 状态, bpf_sk_assign() 会报错 ESOCKTNOSUPPORT
        // 可以参见 xdp-cilium-l4lb/cilium/test/tproxy/socket-lookup/test_sk_lookup.c::sk_assign_estabsocknosupport()
        if (sk->state != BPF_TCP_LISTEN) {
            goto assign;
        }
        // 2. 目标端口任意，但是 socket 是 listen 状态
        bpf_sk_release(sk);
    }

    // 3. 目标端口是 4321，查找 server_map 得到 port 1234 的 server fd
    int zero = 0;
    sk = bpf_map_lookup_elem(&server_map, &zero);
    if (!sk) {
        return TC_ACT_SHOT;
    }

    // 4.检查 port 1234 的 server fd 得是 listen 状态
    if (sk->state != BPF_TCP_LISTEN) {
        bpf_sk_release(sk);
        return TC_ACT_SHOT;
    }

    assign:
    ret = bpf_sk_assign(skb, sk, 0);
    bpf_sk_release(sk);
    return ret;
}

SEC("classifier/sk_assign_test")
int bpf_sk_assign_test(struct __sk_buff *skb) {
    struct bpf_sock_tuple *tuple, ln = {0};
    bool tcp = false;
    int tuple_len;
    int ret = 0;

    void *data_end = (void *) (long) skb->data_end;
    void *data = (void *) (long) skb->data;
    struct ethhdr *eth;
    eth = (struct ethhdr *) (data);
    if ((void *) (eth + 1) > data_end) {
        return TC_ACT_SHOT;
    }
    if (eth->h_proto != bpf_htons(ETH_P_IP)) { // only ipv4
        return TC_ACT_OK;
    }
    struct iphdr *iph = (struct iphdr *) (data + sizeof(*eth));
    if ((void *) (iph + 1) > data_end) {
        return TC_ACT_SHOT;
    }
    if (iph->ihl != 5) {
        /* Options are not supported */
        return TC_ACT_SHOT;
    }

    /* Note that the verifier socket return type for bpf_skc_lookup_tcp()
     * differs from bpf_sk_lookup_udp(), so even though the C-level type is
     * the same here, if we try to share the implementations they will
     * fail to verify because we're crossing pointer types.
     */
    if (iph->protocol == IPPROTO_TCP) {
        return handle_tcp(skb);
    } else if (iph->protocol == IPPROTO_UDP) {
        return handle_udp(skb);
    } else {
        return TC_ACT_OK;
    }
}


int _version SEC("version") = 1;
char _license[] SEC("license") = "GPL";
