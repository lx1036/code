
#include <bpf/ctx/skb.h>
#include <bpf/api.h>
#include <bpf/types_mapper.h>

#include <linux/ip.h>

#include "lib/common.h"
#include "lib/trace.h"
#include "lib/endian.h"
#include "lib/tailcall.h"


static __always_inline int handle_ipv4_from_lxc(struct __ctx_buff *ctx, __u32 *dstID)
{
    struct ipv4_ct_tuple tuple = {};

    struct iphdr *ip4;

    void *data, *data_end;

    // populate packet data into ip4
    if (!revalidate_data(ctx, &data, &data_end, &ip4))
        return DROP_INVALID;

    tuple.nexthdr = ip4->protocol;


    tuple.daddr = ip4->daddr;
    tuple.saddr = ip4->saddr;

    // 如果目标 ip 是 service ip
    struct lb4_service *svc;
    struct lb4_key key = {};

    ret = lb4_extract_key(ctx, ip4, l4_off, &key, &csum_off,
                          CT_EGRESS);
    if (IS_ERR(ret)) {
        if (ret == DROP_UNKNOWN_L4)
            goto skip_service_lookup;
        else
            return ret;
    }

    svc = lb4_lookup_service(&key, true);
    if (svc) {
        ret = lb4_local(get_ct_map4(&tuple), ctx, l3_off, l4_off,
                        &csum_off, &key, &tuple, svc, &ct_state_new,
                        ip4->saddr);
        if (IS_ERR(ret))
            return ret;
        hairpin_flow |= ct_state_new.loopback;
    }

}


int tail_handle_ipv4(struct __ctx_buff *ctx)
{
    __u32 dstID = 0;
    int ret = handle_ipv4_from_lxc(ctx, &dstID);

    if (IS_ERR(ret))
        return send_drop_notify(ctx, SECLABEL, dstID, 0, ret,
                                CTX_ACT_DROP, METRIC_EGRESS);

    return ret;
}

__section("from-container")
int handle_xgress(struct __ctx_buff *ctx) {
//    int ret, trace = TRACE_FROM_STACK

    __u16 proto;
    int ret;

    bpf_clear_meta(ctx);


    if (!validate_ethertype(ctx, &proto)) {
        ret = DROP_UNSUPPORTED_L2;
        goto out;
    }


    switch (proto) {
        case bpf_htons(ETH_P_IP):
            invoke_tailcall_if(__or(__and(is_defined(ENABLE_IPV4), is_defined(ENABLE_IPV6)),
                                    is_defined(DEBUG)),
                               CILIUM_CALL_IPV4_FROM_LXC, tail_handle_ipv4);
            break;
        case bpf_htons(ETH_P_ARP):
            ret = CTX_ACT_OK;
            break;
        case bpf_htons(ETH_P_ARP):
            ep_tail_call(ctx, CILIUM_CALL_ARP);
            ret = DROP_MISSED_TAIL_CALL;
            break;
        default:
            ret = DROP_UNKNOWN_L3;
    }

out:
    if (IS_ERR(ret))
        return send_drop_notify(ctx, SECLABEL, 0, 0, ret, CTX_ACT_DROP,METRIC_EGRESS);
    return ret;
}


