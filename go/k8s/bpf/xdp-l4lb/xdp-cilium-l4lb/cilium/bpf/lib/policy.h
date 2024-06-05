

#ifndef XDP_CILIUM_L4LB_POLICY_H
#define XDP_CILIUM_L4LB_POLICY_H




#include "maps.h"


static __always_inline int
__policy_can_access(const void *map, struct __sk_buff *ctx, __u32 local_id,
                    __u32 remote_id, __u16 dport, __u8 proto, int dir,
                    bool is_untracked_fragment, __u8 *match_type)
{
    struct policy_entry *policy;
    struct policy_key key = {
            .sec_label = remote_id,
            .dport = dport,
            .protocol = proto,
            .egress = !dir,
            .pad = 0,
    };

    /* L4 lookup can't be done on untracked fragments. */
    if (!is_untracked_fragment) {
        /* Start with L3/L4 lookup. */
        policy = bpf_map_lookup_elem((void *) map, &key);
        if (likely(policy)) {
            *match_type = POLICY_MATCH_L3_L4;
            if (unlikely(policy->deny))
                return DROP_POLICY_DENY;
            return policy->proxy_port;
        }

        /* L4-only lookup. */
        key.sec_label = 0;
        policy = bpf_map_lookup_elem((void *) map, &key);
        if (likely(policy)) {
//            account(ctx, policy);
            *match_type = POLICY_MATCH_L4_ONLY;
            if (unlikely(policy->deny))
                return DROP_POLICY_DENY;
            return policy->proxy_port;
        }
        key.sec_label = remote_id;
    }


    /* If L4 policy check misses, fall back to L3. */
    key.dport = 0;
    key.protocol = 0;
    policy = bpf_map_lookup_elem((void *) map, &key);
    if (likely(policy)) {
//        account(ctx, policy);
        *match_type = POLICY_MATCH_L3_ONLY;
        if (unlikely(policy->deny)) // INFO: 这里观察是否 allow/deny
            return DROP_POLICY_DENY;
        return TC_ACT_OK;
    }

    /* Final fallback if allow-all policy is in place. */
    key.sec_label = 0;
    policy = bpf_map_lookup_elem((void *) map, &key);
    if (policy) {
//        account(ctx, policy);
        *match_type = POLICY_MATCH_ALL;
        if (unlikely(policy->deny))
            return DROP_POLICY_DENY;
        return TC_ACT_OK;
    }

    if (skb_load_meta(ctx, CB_POLICY))
        return TC_ACT_OK;

    if (is_untracked_fragment)
        return DROP_FRAG_NOSUPPORT;

    return DROP_POLICY;
}

static __always_inline int
policy_can_egress(struct __sk_buff *ctx, __u32 src_id, __u32 dst_id,
                  __u16 dport, __u8 proto, __u8 *match_type, __u8 *audited)
{
    int ret;
    ret = __policy_can_access(&cilium_policy, ctx, src_id, dst_id, dport, proto, CT_EGRESS, false, match_type);
    return ret;
}

static __always_inline int policy_can_egress4(struct __sk_buff *ctx,
                                              const struct ipv4_ct_tuple *tuple,
                                              __u32 src_id, __u32 dst_id,
                                              __u8 *match_type, __u8 *audited)
{
    return policy_can_egress(ctx, src_id, dst_id, tuple->dport, tuple->nexthdr, match_type, audited);
}



#endif //XDP_CILIUM_L4LB_POLICY_H
