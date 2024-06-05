

#ifndef XDP_CILIUM_L4LB_HOST_FIREWALL_H
#define XDP_CILIUM_L4LB_HOST_FIREWALL_H



#include <linux/ip.h>

#include "common.h"
#include "conntrack.h"
#include "endpoints.h"
#include "policy.h"







static __always_inline int
ipv4_host_policy_ingress(struct __sk_buff *ctx, __u32 *src_id,
                         struct trace_ctx *trace)
{

    int ret = 0;

    struct ipv4_ct_tuple tuple = {};

    void *data_end = (void *)(long)(ctx->data_end);
    void *data = (void *)(long)(ctx->data);
    struct ethhdr *eth = data;
    struct iphdr *ip4;

    ip4 = data + sizeof(*eth);
    if (data + sizeof(*eth) + sizeof(*ip4) > data_end) {
        return TC_ACT_SHOT;
    }

    if (ip4->protocol != IPPROTO_TCP && ip4->protocol != IPPROTO_UDP) {
        return TC_ACT_OK;
    }

    /* Lookup connection in conntrack map. */
    tuple.nexthdr = ip4->protocol;
    tuple.daddr = ip4->daddr;
    tuple.saddr = ip4->saddr;
    ret = ct_lookup4(&cilium_ct_tcp4, &tuple, ctx, l4_off, CT_INGRESS, &ct_state, &trace->monitor);
    if (ret < 0)
        return ret;




    /* Perform policy lookup */





}



static __always_inline int
ipv4_host_policy_egress(struct __sk_buff *ctx, __u32 src_id,
                        __u32 ipcache_srcid __maybe_unused,
                        struct trace_ctx *trace)
{
    int ret, verdict, l4_off, l3_off = ETH_HLEN;
    struct ipv4_ct_tuple tuple = {};
    struct remote_endpoint_info *info;
    __u32 dst_id = 0;
    __u8 policy_match_type = POLICY_MATCH_NONE;
    __u8 audited = 0;
    struct ct_state ct_state_new = {}, ct_state = {};

    void *data_end = (void *)(long)(ctx->data_end);
    void *data = (void *)(long)(ctx->data);
    struct ethhdr *eth = data;
    struct iphdr *ip4;

    ip4 = data + sizeof(*eth);
    if (data + sizeof(*eth) + sizeof(*ip4) > data_end) {
        return TC_ACT_SHOT;
    }

    /* Lookup connection in conntrack map. */
    tuple.nexthdr = ip4->protocol;
    tuple.daddr = ip4->daddr;
    tuple.saddr = ip4->saddr;
    l4_off = l3_off + ipv4_hdrlen(ip4);
    ret = ct_lookup4(&cilium_ct_tcp4, &tuple, ctx, l4_off, CT_EGRESS, &ct_state, &trace->monitor);
    if (ret < 0)
        return ret;


    /* Retrieve destination identity. */
    info = ipcache_lookup4(ip4->daddr);
    if (info && info->sec_label) {
        dst_id = info->sec_label;
    }
    cilium_dbg(ctx, info ? DBG_IP_ID_MAP_SUCCEED4 : DBG_IP_ID_MAP_FAILED4, ip4->daddr, dst_id);
    /* Perform policy lookup. */
    verdict = policy_can_egress4(ctx, &tuple, src_id, dst_id, &policy_match_type, &audited);
    /* Reply traffic and related are allowed regardless of policy verdict. */
    if (ret != CT_REPLY && ret != CT_RELATED && verdict < 0) {
        return verdict;
    }

    // 创建一条 conntrack
    switch (ret) {
        case CT_NEW:
            ct_state_new.src_sec_id = HOST_ID;
            ret = ct_create4(&cilium_ct_tcp4, &cilium_ct_tcp4, &tuple, ctx, CT_EGRESS,
                             &ct_state_new, verdict > 0, false);
            if (IS_ERR(ret))
                return ret;
            break;

        case CT_REOPENED:

        case CT_ESTABLISHED:
        case CT_RELATED:
        case CT_REPLY:
            break;

        default:
            return DROP_UNKNOWN_CT;
    }

    return TC_ACT_OK;
}





#endif //XDP_CILIUM_L4LB_HOST_FIREWALL_H
