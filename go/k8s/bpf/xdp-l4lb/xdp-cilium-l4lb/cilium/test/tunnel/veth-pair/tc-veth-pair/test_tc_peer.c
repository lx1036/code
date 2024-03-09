
#include <stdint.h>
#include <stdbool.h>

#include <linux/bpf.h>
#include <linux/stddef.h>
#include <linux/pkt_cls.h>

// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>


enum {
    dev_src,
    dev_dst,
};

struct bpf_map_def SEC("maps") ifindex_map = {
    .type    = BPF_MAP_TYPE_ARRAY,
    .key_size  = sizeof(int),
    .value_size  = sizeof(int),
    .max_entries = 2,
};


static __always_inline int get_dev_ifindex(int which) {
    int *ifindex = bpf_map_lookup_elem(&ifindex_map, &which);
    return ifindex ? *ifindex : 0;
}

SEC("chk_egress")
int tc_chk(struct __sk_buff *skb) {
    return TC_ACT_SHOT;
}

/**
 * bpf_redirect_peer 意思是：redirect 到 ifindex 的 peer device，这个 peer device 和 ifindex 是一对 veth-pair，且在另一个 namespace 中。
 * 所以观察包流程发现有 *网络加速*：
 *  去包: ns-src(veth_src) -> ns-fwd, veth_src_fwd(ingress) -> bpf_redirect_peer(veth_dst_fwd) -> ns-dst(veth_dst)
 *  回包: ns-dst(veth_dst) -> ns-fwd, veth_dst_fwd(ingress) -> bpf_redirect_peer(veth_src_fwd) -> ns-src(veth_src)
 *  类似于: ns-src(veth_src) <-> ns-dst(veth_dst) 直接通信
 */

SEC("dst_ingress")
int tc_dst(struct __sk_buff *skb) {
    return bpf_redirect_peer(get_dev_ifindex(dev_src), 0);
}

SEC("src_ingress")
int tc_src(struct __sk_buff *skb) {
    return bpf_redirect_peer(get_dev_ifindex(dev_dst), 0);
}

char __license[] SEC("license") = "GPL";

