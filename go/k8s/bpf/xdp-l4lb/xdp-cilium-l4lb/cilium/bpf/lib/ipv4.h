
#ifndef XDP_CILIUM_L4LB_IPV4_H
#define XDP_CILIUM_L4LB_IPV4_H

#include <linux/ip.h>

#include "dbg.h"
#include "metrics.h"


struct ipv4_frag_l4ports {
    __be16	sport;
    __be16	dport;
} __packed;

struct ipv4_frag_id {
    __be32	daddr;
    __be32	saddr;
    __be16	id;		/* L4 datagram identifier */
    __u8	proto;
    __u8	pad;
} __packed;

//#ifdef ENABLE_IPV4_FRAGMENTS
struct bpf_elf_map __section_maps IPV4_FRAG_DATAGRAMS_MAP = {
        .type           = BPF_MAP_TYPE_LRU_HASH,
        .size_key	= sizeof(struct ipv4_frag_id),
        .size_value	= sizeof(struct ipv4_frag_l4ports),
        .pinning	= PIN_GLOBAL_NS,
        .max_elem	= CILIUM_IPV4_FRAG_MAP_MAX_ENTRIES,
};
//#endif

// 这个函数意思是计算 ipv4 header 字节大小
// https://en.wikipedia.org/wiki/Internet_Protocol_version_4#IHL
static __always_inline int ipv4_hdrlen(const struct iphdr *ip4) {
    return ip4->ihl * 4; // 4 表示 32bits，4字节. 最小是 5*4=20字节，最大 15*4=60字节, 5<=ip4->ihl<=15
}

static __always_inline bool ipv4_is_not_first_fragment(const struct iphdr *ip4)
{
    /* Ignore "More fragments" bit to catch all fragments but the first */
    return ip4->frag_off & bpf_htons(0x1FFF);
}

/* Simply a reverse of ipv4_is_not_first_fragment to avoid double negative. */
static __always_inline bool ipv4_has_l4_header(const struct iphdr *ip4)
{
    return !ipv4_is_not_first_fragment(ip4);
}

// 如果是 fragment 包，则 ip4->frag_off = 0x3FFF
static __always_inline bool ipv4_is_fragment(const struct iphdr *ip4)
{
    /* The frag_off portion of the header consists of:
     *
     * +----+----+----+----------------------------------+
     * | RS | DF | MF | ...13 bits of fragment offset... |
     * +----+----+----+----------------------------------+
     *
     * If "More fragments" or the offset is nonzero, then this is an IP
     * fragment (RFC791).
     */
    return ip4->frag_off & bpf_htons(0x3FFF); // 0011, https://datatracker.ietf.org/doc/html/rfc791#section-3.1
}

/*
 * IP包的分片（fragmentation）是在IP层进行的，它主要是为了解决在不同网络之间传输数据包时可能遇到的最大传输单元（MTU，Maximum Transmission Unit）不匹配问题。
 * 当一个IP包的大小超过了路径中某个链路的MTU时，就需要对该IP包进行分片。

 * 分片可以在以下情况产生：
当要发送的IP包大于链路层的最大帧大小时，IP包将在发送端被分片。
当IP包在网络中传输时，如果遇到的链路MTU小于IP包的大小，那么路由器将对IP包进行分片。

例如，如果一个IP包的大小是2000字节，而下一跳的MTU只有1500字节，路由器就需要分片。它会把原来的IP数据包分成两个IP数据包，一个1500字节，
 一个500字节（实际上可能更小，因为需要包含IP分片所需的额外头部信息）。这两个新的IP包都包含原始IP头的复制品（除了一些字段如片偏移、更多片等），
 并且它们都有相同的标识字段。这样，当这些片到达目的地时，接收端能够根据片偏移和标识字段，将这些片重新组装成原始的IP包。

需要注意的是，IP分片可能会对性能产生影响，并可能引发一些安全问题。因此，通常会尽量避免IP分片。在实际环境中，通常会使用路径MTU发现（Path MTU Discovery）技术，
 找出数据包传输路径中的最小MTU，然后调整数据包的大小，使其不超过这个MTU，从而避免IP分片。
 */
static __always_inline int
ipv4_handle_fragmentation(struct __ctx_buff *ctx, const struct iphdr *ip4, int l4_off, int ct_dir,
                          struct ipv4_frag_l4ports *ports, bool *has_l4_header) {
    int ret, dir;
    bool is_fragment, not_first_fragment;

    struct ipv4_frag_id frag_id = {
            .daddr = ip4->daddr,
            .saddr = ip4->saddr,
            .id = ip4->id,
            .proto = ip4->protocol,
            .pad = 0,
    };

    is_fragment = ipv4_is_fragment(ip4);
    dir = ct_to_metrics_dir(ct_dir);

    if (unlikely(is_fragment)) { // 如果 is_fragment 为 false，即该包不是 ip 分片的包
        update_metrics(ctx_full_len(ctx), dir, REASON_FRAG_PACKET);

        not_first_fragment = ipv4_is_not_first_fragment(ip4);
        if (has_l4_header)
            *has_l4_header = !not_first_fragment;

        if (likely(not_first_fragment))
            return ipv4_frag_get_l4ports(&frag_id, ports);
    }

    /* load sport + dport into tuple */
    ret = ctx_load_bytes(ctx, l4_off, ports, 4);
    if (ret < 0)
        return ret;

    if (unlikely(is_fragment)) {
        /* First logical fragment for this datagram (not necessarily the first
         * we receive). Fragment has L4 header, create an entry in datagrams map.
         */
        if (map_update_elem(&IPV4_FRAG_DATAGRAMS_MAP, &frag_id, ports, BPF_ANY))
            update_metrics(ctx_full_len(ctx), dir, REASON_FRAG_PACKET_UPDATE);

        /* Do not return an error if map update failed, as nothing prevents us
         * to process the current packet normally.
         */
    }

    return 0;
}





#endif //XDP_CILIUM_L4LB_IPV4_H
