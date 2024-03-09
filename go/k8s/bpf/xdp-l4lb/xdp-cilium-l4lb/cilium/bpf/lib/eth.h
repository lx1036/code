

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

static __always_inline int eth_store_saddr_aligned(struct __ctx_buff *ctx, const __u8 *mac, int off) {
    return ctx_store_bytes(ctx, off + ETH_ALEN, mac, ETH_ALEN, 0);
}

static __always_inline int eth_store_daddr_aligned(struct __ctx_buff *ctx, const __u8 *mac, int off) {
    return ctx_store_bytes(ctx, off, mac, ETH_ALEN, 0);
}

// smac -> *ctx
static __always_inline int eth_store_saddr(struct __ctx_buff *ctx, const __u8 *mac, int off) {
#if !CTX_DIRECT_WRITE_OK
    return eth_store_saddr_aligned(ctx, mac, off);
#else
    void *data_end = ctx_data_end(ctx);
    void *data = ctx_data(ctx);

    if (ctx_no_room(data + off + ETH_ALEN * 2, data_end))
        return -EFAULT;
    /* Need to use builtin here since mac came potentially from
     * struct bpf_fib_lookup where it's not aligned on stack. :(
     */
    __bpf_memcpy_builtin(data + off + ETH_ALEN, mac, ETH_ALEN);
    return 0;
#endif
}

// dmac -> *ctx
static __always_inline int eth_store_daddr(struct __ctx_buff *ctx, const __u8 *mac, int off) {
#if !CTX_DIRECT_WRITE_OK
    return eth_store_daddr_aligned(ctx, mac, off);
#else
    void *data_end = ctx_data_end(ctx);
    void *data = ctx_data(ctx);

    if (ctx_no_room(data + off + ETH_ALEN, data_end))
        return -EFAULT;
    /* Need to use builtin here since mac came potentially from
     * struct bpf_fib_lookup where it's not aligned on stack. :(
     */
    __bpf_memcpy_builtin(data + off, mac, ETH_ALEN);
    return 0;
#endif
}



#endif /* __LIB_ETH__ */
