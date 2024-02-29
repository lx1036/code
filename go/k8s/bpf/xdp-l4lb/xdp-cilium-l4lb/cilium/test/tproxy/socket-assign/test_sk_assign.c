

#include <stddef.h>
#include <stdbool.h>
#include <string.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>
#include <linux/tcp.h>
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

static const int dst_port = 4321;

/* Must match struct bpf_elf_map layout from iproute2 */
struct bpf_elf_map {
    __u32 type;
    __u32 size_key;
    __u32 size_value;
    __u32 max_elem;
    __u32 flags;
    __u32 id;
    __u32 pinning;
    __u32 inner_id;
    __u32 inner_idx;
};
struct bpf_elf_map __section_maps server_map = {
        .type        = BPF_MAP_TYPE_SOCKMAP,
        .size_key    = sizeof(int),
        .size_value    = sizeof(__u64),
        .pinning    = PIN_GLOBAL_NS, // 设置 LIBBPF_PIN_BY_NAME 导致 pin path 为 /sys/fs/bpf/tc/424e22c8e74276a6484f398886d426f441d9b849/
        .max_elem    = 1,
};


static __always_inline struct bpf_sock_tuple *
get_tuple(struct __sk_buff *skb, bool *tcp) {
    void *data_end = (void *) (long) skb->data_end;
    void *data = (void *) (long) skb->data;
    struct bpf_sock_tuple *result;
    struct ethhdr *eth;
    __u64 tuple_len;
    __u8 proto = 0;
    __u64 ihl_len;

    eth = (struct ethhdr *) (data);
    if ((void *) (eth + 1) > data_end)
        return NULL;

    if (eth->h_proto == bpf_htons(ETH_P_IP)) {
        struct iphdr *iph = (struct iphdr *) (data + sizeof(*eth));
        if ((void *) (iph + 1) > data_end)
            return NULL;
        if (iph->ihl != 5)
            /* Options are not supported */
            return NULL;

        ihl_len = iph->ihl * 4; // iphdr 头字节长度 <<2
        proto = iph->protocol;
        result = (struct bpf_sock_tuple *) &iph->saddr; // 这个类型转换非常经典，bpf_sock_tuple 就是 saddr/daddr/sport/dport!!!
    } else {
        return (struct bpf_sock_tuple *) data;
    }

    if (proto != IPPROTO_TCP && proto != IPPROTO_UDP)
        return NULL;

    *tcp = (proto == IPPROTO_TCP);
    return result;
}

static __always_inline int handle_udp(struct __sk_buff *skb, struct bpf_sock_tuple *tuple) {
    struct bpf_sock_tuple ln = {0};
    struct bpf_sock *sk;
    const int zero = 0;
    size_t tuple_len;
    __u16 dport;
    int ret;

    tuple_len = sizeof(tuple->ipv4);
    if ((void *) (tuple + tuple_len) > (void *) (long) skb->data_end)
        return TC_ACT_SHOT;

    sk = bpf_sk_lookup_udp(skb, tuple, tuple_len, BPF_F_CURRENT_NETNS, 0);
    if (sk)
        goto assign;

    dport = tuple->ipv4.dport;
    if (dport != bpf_htons(dst_port))
        return TC_ACT_OK;

    sk = bpf_map_lookup_elem(&server_map, &zero);
    if (!sk)
        return TC_ACT_SHOT;

    assign:
    ret = bpf_sk_assign(skb, sk, 0);
    bpf_sk_release(sk);
    return ret;
}

static __always_inline int handle_tcp(struct __sk_buff *skb, struct bpf_sock_tuple *tuple) {
    struct bpf_sock_tuple ln = {0};
    struct bpf_sock *sk;
    const int zero = 0;
    size_t tuple_len;
    __u16 dport;
    int ret;

    tuple_len = sizeof(tuple->ipv4);
    if ((void *) (tuple + tuple_len) > (void *) (long) skb->data_end)
        return TC_ACT_SHOT;

    sk = bpf_skc_lookup_tcp(skb, tuple, tuple_len, BPF_F_CURRENT_NETNS, 0);
    if (sk) {
        // 1. 目标端口任意，获取到 socket 不是 listen 状态，说明已经是 establish 状态，直接 bpf_sk_assign()
        if (sk->state != BPF_TCP_LISTEN)
            goto assign;
        // 2. 目标端口任意，但是 socket 是 listen 状态
        bpf_sk_release(sk);
    }

    dport = tuple->ipv4.dport;
    if (dport != bpf_htons(dst_port)) // 2. 目标端口任意(不是 4321)，但是 socket 是 listen 状态，交给 netfilter 自己处理
        return TC_ACT_OK;

    // 3. 目标端口是 4321，查找 server_map 得到 port 1234 的 server fd
    sk = bpf_map_lookup_elem(&server_map, &zero);
    if (!sk)
        return TC_ACT_SHOT;

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

    tuple = get_tuple(skb, &tcp);
    if (!tuple)
        return TC_ACT_SHOT;

    /* Note that the verifier socket return type for bpf_skc_lookup_tcp()
     * differs from bpf_sk_lookup_udp(), so even though the C-level type is
     * the same here, if we try to share the implementations they will
     * fail to verify because we're crossing pointer types.
     */
    if (tcp)
        ret = handle_tcp(skb, tuple);
    else
        ret = handle_udp(skb, tuple);

    return ret == 0 ? TC_ACT_OK : TC_ACT_SHOT;
}


int _version SEC("version") = 1;
char _license[] SEC("license") = "GPL";
