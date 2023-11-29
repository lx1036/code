

#ifndef XDP_CILIUM_L4LB_HELPERS_SKB_H
#define XDP_CILIUM_L4LB_HELPERS_SKB_H


#include <linux/bpf.h>

#include "compiler.h"
#include "helpers.h"
#include "features_skb.h"






/* Packet redirection */
static int BPF_FUNC(redirect, int ifindex, __u32 flags);
static int BPF_FUNC(redirect_neigh, int ifindex, struct bpf_redir_neigh *params,
                    int plen, __u32 flags);
static int BPF_FUNC(redirect_peer, int ifindex, __u32 flags);

/* Packet manipulation */
static int BPF_FUNC(skb_load_bytes, struct __sk_buff *skb, __u32 off,
                    void *to, __u32 len);
static int BPF_FUNC(skb_store_bytes, struct __sk_buff *skb, __u32 off,
                    const void *from, __u32 len, __u32 flags);

static int BPF_FUNC(l3_csum_replace, struct __sk_buff *skb, __u32 off,
                    __u32 from, __u32 to, __u32 flags);
static int BPF_FUNC(l4_csum_replace, struct __sk_buff *skb, __u32 off,
                    __u32 from, __u32 to, __u32 flags);

static int BPF_FUNC(skb_adjust_room, struct __sk_buff *skb, __s32 len_diff,
                    __u32 mode, __u64 flags);




#endif //XDP_CILIUM_L4LB_HELPERS_SKB_H
