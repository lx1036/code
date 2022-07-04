//
// Created by 刘祥 on 7/1/22.
//

#ifndef __LIB_COMMON_H_
#define __LIB_COMMON_H_

#include <bpf/ctx/ctx.h>
#include <bpf/ctx/common.h>
#include <bpf/api.h>
#include <linux/if_ether.h>


#include <endian.h>


#define DROP_UNSUPPORTED_L2 -166
#define CILIUM_CALL_IPV4_FROM_LXC		7


// 验证数据：
// (1)包的 header 长度是否正确
// (2)包的协议是否正确
static __always_inline bool validate_ethertype(struct __sk_buff *ctx, __u16 *proto)
{
    void *data = ctx_data(ctx); // 从网络包中
    void *data_end = ctx_data_end(ctx);
    struct ethhdr *eth = data;

    if (data + ETH_HLEN > data_end)
        return false;

    *proto = eth->h_proto;
    if (bpf_ntohs(*proto) < ETH_P_802_3_MIN)
        return false; /* non-Ethernet II unsupported */
    return true;
}


struct ipv4_ct_tuple {
    /* Address fields are reversed, i.e.,
     * these field names are correct for reply direction traffic.
     */
    __be32		daddr;
    __be32		saddr;
    /* The order of dport+sport must not be changed!
     * These field names are correct for original direction traffic.
     */
    __be16		dport;
    __be16		sport;
    __u8		nexthdr;
    __u8		flags;
} __packed;


static __always_inline __maybe_unused bool
__revalidate_data(struct __ctx_buff *ctx, void **data_, void **data_end_,
                  void **l3, const __u32 l3_len, const bool pull)
{
    const __u32 tot_len = ETH_HLEN + l3_len;
    void *data_end;
    void *data;

    /* Verifier workaround, do this unconditionally: invalid size of register spill. */
    if (pull)
        ctx_pull_data(ctx, tot_len);

    data_end = ctx_data_end(ctx);
    data = ctx_data(ctx);
    if (data + tot_len > data_end)
        return false;


    /* Verifier workaround: pointer arithmetic on pkt_end prohibited. */
    *data_ = data;
    *data_end_ = data_end;

    *l3 = data + ETH_HLEN;
    return true;
}


/* revalidate_data() initializes the provided pointers from the ctx.
 * Returns true if 'ctx' is long enough for an IP header of the provided type,
 * false otherwise.
 */
#define revalidate_data(ctx, data, data_end, ip)			\
	__revalidate_data(ctx, data, data_end, (void **)ip, sizeof(**ip), false)






#include "overloadable.h"

#endif //__LIB_COMMON_H_