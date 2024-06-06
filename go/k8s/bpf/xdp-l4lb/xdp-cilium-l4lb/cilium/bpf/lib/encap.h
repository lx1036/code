


#ifndef __LIB_ENCAP_H_
#define __LIB_ENCAP_H_

#include "common.h"
#include "dbg.h"
#include "trace.h"
#include "l3.h"


//#ifdef ENCAP_IFINDEX
//#ifdef ENABLE_IPSEC

/* NOT_VTEP_DST is passed to an encapsulation function when the
 * destination of the tunnel is not a VTEP.
 */
#define NOT_VTEP_DST 0

static __always_inline int
__encap_with_nodeid(struct __sk_buff *ctx, __u32 tunnel_endpoint,
		    __u32 seclabel, __u32 monitor)
{
	struct bpf_tunnel_key key = {};
	__u32 node_id;
	int ret;

	/* When encapsulating, a packet originating from the local host is
	 * being considered as a packet from a remote node as it is being
	 * received.
	 */
	if (seclabel == HOST_ID)
		seclabel = LOCAL_NODE_ID;

	node_id = bpf_htonl(tunnel_endpoint);
	key.tunnel_id = seclabel;
	key.remote_ipv4 = node_id;
	key.tunnel_ttl = 64;

	cilium_dbg(ctx, DBG_ENCAP, node_id, seclabel);

	ret = (int)bpf_skb_set_tunnel_key(ctx, &key, sizeof(key), BPF_F_ZERO_CSUM_TX);
	if (unlikely(ret < 0)) {
        return DROP_WRITE_ERROR;
    }

//	send_trace_notify(ctx, TRACE_TO_OVERLAY, seclabel, 0, 0, ENCAP_IFINDEX, 0, monitor);
	return 0;
}

static __always_inline int
__encap_and_redirect_with_nodeid(struct __sk_buff *ctx, __u32 tunnel_endpoint,
                                 __u32 seclabel, __u32 vni,
                                 const struct trace_ctx *trace)
{
    int ret = __encap_with_nodeid(ctx, tunnel_endpoint, seclabel, vni, trace->reason, trace->monitor);
    if (ret != 0)
        return ret;

    return (int) bpf_redirect(ENCAP_IFINDEX, 0);
}

/* encap_and_redirect_with_nodeid returns IPSEC_ENDPOINT after ctx meta-data is
 * set when IPSec is enabled. Caller should pass the ctx to the stack at this
 * point. Otherwise returns CTX_ACT_TX on successful redirect to tunnel device.
 * On error returns CTX_ACT_DROP, DROP_NO_TUNNEL_ENDPOINT or DROP_WRITE_ERROR.
 */
static __always_inline int
encap_and_redirect_with_nodeid(struct __sk_buff *ctx, __u32 tunnel_endpoint,
                               __u8 key __maybe_unused, __u32 seclabel,
                               const struct trace_ctx *trace)
{
//#ifdef ENABLE_IPSEC
//    if (key)
//		return encap_and_redirect_nomark_ipsec(ctx, tunnel_endpoint, key, seclabel);
//#endif

    return __encap_and_redirect_with_nodeid(ctx, tunnel_endpoint, seclabel, NOT_VTEP_DST, trace);
}

static __always_inline int
encap_and_redirect_netdev(struct __sk_buff *ctx, struct endpoint_key *k,
                          __u32 seclabel, const struct trace_ctx *trace)
{
    struct endpoint_key *tunnel;
    tunnel = bpf_map_lookup_elem(&cilium_tunnel_map, k);
    if (!tunnel)
        return DROP_NO_TUNNEL_ENDPOINT;

    return __encap_and_redirect_with_nodeid(ctx, tunnel->ip4, seclabel, NOT_VTEP_DST, trace);
}


//#endif /* ENCAP_IFINDEX */
//#endif /* __LIB_ENCAP_H_ */
