//
// Created by 刘祥 on 7/6/22.
//

#ifndef BPF_POLICY_H
#define BPF_POLICY_H

#include "eps.h"

static __always_inline int
policy_sk_egress(__u32 identity, __u32 ip,  __u16 dport)
{
    void *map = lookup_ip4_endpoint_policy_map(ip);
    if (!map)
        return CTX_ACT_OK;
}




#endif //BPF_POLICY_H
