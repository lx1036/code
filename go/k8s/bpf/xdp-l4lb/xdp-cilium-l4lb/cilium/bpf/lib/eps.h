

#ifndef XDP_CILIUM_L4LB_EPS_H
#define XDP_CILIUM_L4LB_EPS_H


#include <linux/ip.h>
#include <linux/ipv6.h>

#include "maps.h"


static __always_inline __maybe_unused struct endpoint_info *
__lookup_ip4_endpoint(__u32 ip)
{
    struct endpoint_key key = {};

    key.ip4 = ip;
    key.family = ENDPOINT_KEY_IPV4;

    return map_lookup_elem(&ENDPOINTS_MAP, &key);
}

static __always_inline __maybe_unused struct endpoint_info *
lookup_ip4_endpoint(const struct iphdr *ip4)
{
    return __lookup_ip4_endpoint(ip4->daddr);
}









#endif //XDP_CILIUM_L4LB_EPS_H
