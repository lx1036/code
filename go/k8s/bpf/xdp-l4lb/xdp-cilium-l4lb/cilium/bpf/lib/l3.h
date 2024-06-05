
#ifndef XDP_CILIUM_L4LB_L3_H
#define XDP_CILIUM_L4LB_L3_H

#include <stdbool.h>

#include "eth.h"
#include "maps.h"


/* Performs IPv4 L2/L3 handling and delivers the packet to the destination pod
 * on the same node, either via the stack or via a redirect call.
 * Depending on the configuration, it may also enforce ingress policies for the
 * destination pod via a tail call.
 */
static __always_inline int ipv4_local_delivery(struct __sk_buff *ctx, int l3_off,
                                               __u32 seclabel, struct iphdr *ip4,
                                               const struct endpoint_info *ep,
                                               __u8 direction __maybe_unused,
                                               bool from_host __maybe_unused)
{
    mac_t router_mac = ep->node_mac;
    mac_t lxc_mac = ep->mac;
    int ret;
    if (ipv4_dec_ttl(ctx, l3_off, ip4)) {
        /* FIXME: Send ICMP TTL */
        return DROP_INVALID;
    }

    if (router_mac && eth_store_saddr(ctx, (__u8 *) &router_mac, 0) < 0)
        return DROP_WRITE_ERROR;
    if (lxc_mac && eth_store_daddr(ctx, (__u8 *) &lxc_mac, 0) < 0)
        return DROP_WRITE_ERROR;


    ctx->mark |= MARK_MAGIC_IDENTITY;
    set_identity_mark(ctx, seclabel);

    return redirect_ep(ctx, (int)ep->ifindex, from_host);
}





#endif //XDP_CILIUM_L4LB_L3_H
