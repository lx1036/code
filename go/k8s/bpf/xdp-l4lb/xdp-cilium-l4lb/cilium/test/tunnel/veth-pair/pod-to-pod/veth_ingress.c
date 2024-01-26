

// /root/linux-5.10.142/include/uapi/linux/bpf.h
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>

// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#ifndef __section
#define __section(X) __attribute__((section(X), used))
#endif
#ifndef __section_maps
#define __section_maps __section("maps")
#endif

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

#define PIN_GLOBAL_NS   2


struct endpointKey {
    __u32 ip;
};
struct endpointInfo {
    __u32 ifIndex;
    __u32 lxcIfIndex;
    __u8 mac[8];
    __u8 nodeMac[8];
};
//struct {
//    __uint(type, BPF_MAP_TYPE_HASH);
//    __uint(max_entries, 256);
//    __type(key, struct endpointKey);
//    __type(value, struct endpointInfo);
//    __uint(pinning, LIBBPF_PIN_BY_NAME);
//} ding_lxc SEC(".maps");

struct bpf_elf_map __section_maps ding_lxc = {
    .type		= BPF_MAP_TYPE_HASH,
    .size_key	= sizeof(struct endpointKey),
    .size_value	= sizeof(struct endpointInfo),
    .pinning	= PIN_GLOBAL_NS, // 设置 LIBBPF_PIN_BY_NAME 导致 pin path 为 /sys/fs/bpf/tc/424e22c8e74276a6484f398886d426f441d9b849/
    .max_elem	= 256,
};

SEC("tc_from_container")
int from_container(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;
    if (data + sizeof(struct ethhdr) + sizeof(struct iphdr) > data_end) {
        return TC_ACT_UNSPEC;
    }

    struct ethhdr  *eth  = data;
    struct iphdr   *ip   = (data + sizeof(struct ethhdr));
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        // 当eBPF程序返回 TC_ACT_UNSPEC 时，意味着它没有明确地指示内核如何处理这个数据包。
        // 在这种情况下，内核通常会按照默认的方式处理数据包，例如将其传递给下一个处理阶段或者直接丢弃

        /**
         * arp 协议由内核自己处理，bpf 程序没有处理
         */
        return TC_ACT_UNSPEC;
    }

    // 在 go 那头儿往 ebpf 的 map 里存的时候我这个 arm 是按照小端序存的
    // 这里给转成网络的大端序
    __u32 src_ip = bpf_htonl(ip->saddr);
    __u32 dst_ip = bpf_htonl(ip->daddr);
    bpf_printk("src ip 0x%x, remote ip 0x%x\n", src_ip, dst_ip);
    // 拿到 mac 地址
    __u8 src_mac[ETH_ALEN];
    __u8 dst_mac[ETH_ALEN];
    struct endpointKey epKey = {};
    epKey.ip = dst_ip;
    // 在 lxc 中查找
    struct endpointInfo *ep = bpf_map_lookup_elem(&ding_lxc, &epKey); // 去包: 100.0.1.1->100.0.1.2, 回包:
    if (ep) {
        // 如果能找到说明是要发往本机其他 pod 中的
        // 把 mac 地址改成目标 pod 的两对儿 veth 的 mac 地址
        __builtin_memcpy(src_mac, ep->nodeMac, ETH_ALEN);
        __builtin_memcpy(dst_mac, ep->mac, ETH_ALEN);
        bpf_skb_store_bytes(skb, offsetof(struct ethhdr, h_source), dst_mac, ETH_ALEN, 0);
        bpf_skb_store_bytes(skb, offsetof(struct ethhdr, h_dest), src_mac, ETH_ALEN, 0);
        bpf_printk("src mac 0x%x, dst mac 0x%x, lxcIfIndex 0x%x\n", ep->nodeMac, ep->mac, ep->lxcIfIndex);
        return bpf_redirect_peer(ep->lxcIfIndex, 0); // 这里 ep->lxcIfIndex 虽然是 lxc 侧的 veth，但是 bpf_redirect_peer() 会直接把包发到 ns 侧的 peer veth; 且 flags 必须是 0
    }

    return TC_ACT_UNSPEC;
}


char _license[] SEC("license") = "GPL";
int _version SEC("version") = 1;
