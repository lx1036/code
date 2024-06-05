

#ifndef __LIB_ETH__
#define __LIB_ETH__

#include <linux/if_ether.h>

#ifndef ETH_HLEN
#define ETH_HLEN __ETH_HLEN
#endif


union macaddr {
    struct {
        __u32 p1;
        __u16 p2;
    };
    __u8 addr[6];
};

static __always_inline int eth_store_saddr_aligned(struct __sk_buff *ctx, const __u8 *mac, int off) {
    return (int)bpf_skb_store_bytes(ctx, off+ETH_ALEN, mac, ETH_ALEN, 0);
}

static __always_inline int eth_store_daddr_aligned(struct __sk_buff *ctx, const __u8 *mac, int off) {
    return (int)bpf_skb_store_bytes(ctx, off, mac, ETH_ALEN, 0);
}

// smac -> *ctx
static __always_inline int eth_store_saddr(struct __sk_buff *ctx, const __u8 *mac, int off) {
    return eth_store_saddr_aligned(ctx, mac, off);
}

// dmac -> *ctx
static __always_inline int eth_store_daddr(struct __sk_buff *ctx, const __u8 *mac, int off) {
    return eth_store_daddr_aligned(ctx, mac, off);
}



#endif /* __LIB_ETH__ */
