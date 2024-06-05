

#ifndef XDP_CILIUM_L4LB_HOST_FIREWALL_H
#define XDP_CILIUM_L4LB_HOST_FIREWALL_H



#include <linux/ip.h>

#include "common.h"
#include "conntrack.h"







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
    ret = ct_lookup4(&cilium_ct_tcp4, &tuple, ctx, l4_off, CT_INGRESS,
                     &ct_state, &trace->monitor);
    if (ret < 0)
        return ret;




    /* Perform policy lookup */





}





#endif //XDP_CILIUM_L4LB_HOST_FIREWALL_H
