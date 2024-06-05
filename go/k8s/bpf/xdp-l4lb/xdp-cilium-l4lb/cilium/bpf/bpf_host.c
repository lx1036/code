
#include <stdbool.h>

#include <lib/maps.h>

#include <linux/if_ether.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_endian.h>

#include <lib/host_firewall.h>
#include <lib/endpoints.h>
#include <lib/l3.h>


#define CB_SRC_IDENTITY	0
#define CILIUM_NET_MAC  { .addr = { 0xce, 0x72, 0xa7, 0x03, 0x88, 0x57 } }

static __always_inline int rewrite_dmac_to_host(struct __sk_buff *ctx, __u32 src_identity) {
    /* When attached to cilium_host, we rewrite the DMAC to the mac of
     * cilium_host (peer) to ensure the packet is being considered to be
     * addressed to the host (PACKET_HOST).
     */
    union macaddr cilium_net_mac = CILIUM_NET_MAC;
    /* Rewrite to destination MAC of cilium_net (remote peer) */
    if (eth_store_daddr(ctx, (__u8 *) &cilium_net_mac.addr, 0) < 0) {
        return TC_ACT_SHOT;
//        return send_drop_notify_error(ctx, src_identity, DROP_WRITE_ERROR, TC_ACT_OK, METRIC_INGRESS);
    }

    return TC_ACT_OK;
}

static __always_inline int
handle_ipv4(struct __sk_buff *ctx, const bool from_host) {
    int ret;
    bool skip_redirect = false;
    struct endpoint_info *ep;

    void *data_end = (void *)(long)(ctx->data_end);
    void *data = (void *)(long)(ctx->data);
    struct ethhdr *eth = data;
    struct iphdr *ip4;

    ip4 = data + sizeof(*eth);
    if (data + sizeof(*eth) + sizeof(*ip4) > data_end) {
        return TC_ACT_SHOT;
    }


    __u32 proxy_identity = skb_load_meta(ctx, CB_SRC_IDENTITY);


    if (!from_host) {

    }

    /* Without bpf_redirect_neigh() helper, we cannot redirect a
	 * packet to a local endpoint in the direct routing mode, as
	 * the redirect bypasses nf_conntrack table. This makes a
	 * second reply from the endpoint to be MASQUERADEd or to be
	 * DROP-ed by k8s's "--ctstate INVALID -j DROP" depending via
	 * which interface it was inputed. With bpf_redirect_neigh()
	 * we bypass request and reply path in the host namespace and
	 * do not run into this issue.
	 */
    if (!from_host) {
        skip_redirect = true;
    }


    if (from_host) {
        // INFO: network policy
        /* We're on the egress path of cilium_host. */
//        ret = ipv4_host_policy_egress(ctx, proxy_identity, ipcache_srcid, &trace);
//        if (IS_ERR(ret))
//            return ret;
    }

    if (skip_redirect) {
        return TC_ACT_OK;
    }

    if (from_host) {
        /* If we are attached to cilium_host at egress, this will
         * rewrite the destination MAC address to the MAC of cilium_net.
         */
        ret = rewrite_dmac_to_host(ctx, proxy_identity);
        /* DIRECT PACKET READ INVALID */
        if (IS_ERR(ret))
            return ret;

        if (!revalidate_data(ctx, &data, &data_end, &ip4))
            return DROP_INVALID;
    }

    /* Lookup IPv4 address in list of local endpoints and host IPs */
    // INFO: 这里的 endpoint 是用户态写入的，代码逻辑为：
    ep = lookup_ip4_endpoint(ip4);
    if (ep) {
        /* Let through packets to the node-ip so they are processed by
         * the local ip stack.
         */
        if (ep->flags & ENDPOINT_F_HOST) {
            return TC_ACT_OK;
        }

        return ipv4_local_delivery(ctx, ETH_HLEN, proxy_identity, ip4, ep, METRIC_INGRESS, from_host);
    }

    /* Below remainder is only relevant when traffic is pushed via cilium_host.
     * For traffic coming from external, we're done here.
     */
    if (!from_host) {
        return TC_ACT_OK;
    }



    return TC_ACT_OK;
}

static __always_inline int
do_netdev(struct __sk_buff *ctx, __u16 proto, const bool from_host)
{
    int ret;
    if (from_host) {

    } else {

    }

    switch (proto) {
    case bpf_htons(ETH_P_ARP):
        ret = TC_ACT_OK;
        break;
    case bpf_htons(ETH_P_IP):
        if (from_host) {
            ret = handle_ipv4(ctx, from_host);
        } else {

        }
        break;
    default:
        ret = TC_ACT_OK;
        break;
    }

    return ret;
}

static __always_inline int
handle_netdev(struct __sk_buff *ctx, const bool from_host)
{
    __u16 proto = 0;
    void *data_end = (void *)(long)(ctx->data_end);
    void *data = (void *)(long)(ctx->data);
    struct ethhdr *eth = data;
    if (data + ETH_HLEN > data_end) {
        return TC_ACT_SHOT;
    }

    proto = eth->h_proto;
    return do_netdev(ctx, proto, from_host);
}

// 主要创建 conntrack 记录
SEC("to-host")
int to_host(struct __sk_buff *ctx) {
    __u16 proto = 0;
    int ret = TC_ACT_OK;

    void *data_end = (void *)(long)(ctx->data_end);
    void *data = (void *)(long)(ctx->data);
    struct ethhdr *eth = data;
    if (data + ETH_HLEN > data_end) {
        return TC_ACT_SHOT;
    }

    proto = eth->h_proto;
    switch (proto) {
    case bpf_htons(ETH_P_ARP):
        ret = TC_ACT_OK;
        break;

    case bpf_htons(ETH_P_IP):
        ret = ipv4_host_policy_ingress(ctx, &src_id, &trace);
        break;

    default:
        ret = TC_ACT_SHOT;
        break;
    }

    return ret;
}

SEC("from-host")
int from_host(struct __sk_buff *ctx) {
    return handle_netdev(ctx, true);
}





