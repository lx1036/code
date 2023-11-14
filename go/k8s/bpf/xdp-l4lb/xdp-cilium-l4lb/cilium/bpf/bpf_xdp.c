

#include <bpf/ctx/xdp.h>
#include <bpf/api.h>

#include <node_config.h>
#include <netdev_config.h>
#include <filter_config.h>


#include "lib/common.h"
#include "lib/maps.h"
#include "lib/eps.h"
#include "lib/events.h"
#include "lib/nodeport.h"

static __always_inline __maybe_unused int
bpf_xdp_exit(struct xdp_md* ctx, const int verdict)
{
	if (verdict == CTX_ACT_OK) {
		__u32 meta_xfer = ctx_load_meta(ctx, XFER_MARKER);

		/* We transfer data from XFER_MARKER. This specifically
		 * does not break packet trains in GRO.
		 */
		if (meta_xfer) {
			if (!ctx_adjust_meta(ctx, -(int)sizeof(meta_xfer))) {
				__u32 *data_meta = ctx_data_meta(ctx);
				__u32 *data = ctx_data(ctx);

				if (!ctx_no_room(data_meta + 1, data))
					data_meta[0] = meta_xfer;
			}
		}
	}

	return verdict;
}

#ifdef ENABLE_IPV4
// ENABLE_NODEPORT_ACCELERATION 默认开启
#ifdef ENABLE_NODEPORT_ACCELERATION
__section_tail(CILIUM_MAP_CALLS, CILIUM_CALL_IPV4_FROM_LXC)
int tail_lb_ipv4(struct __ctx_buff *ctx)
{
	int ret = CTX_ACT_OK;

	if (!bpf_skip_nodeport(ctx)) {
		ret = nodeport_lb4(ctx, 0);
		if (IS_ERR(ret))
			return send_drop_notify_error(ctx, 0, ret, CTX_ACT_DROP, METRIC_INGRESS);
	}

	return bpf_xdp_exit(ctx, ret);
}

static __always_inline int check_v4_lb(struct __ctx_buff *ctx)
{
	ep_tail_call(ctx, CILIUM_CALL_IPV4_FROM_LXC);
	return send_drop_notify_error(ctx, 0, DROP_MISSED_TAIL_CALL, CTX_ACT_DROP, METRIC_INGRESS);
}
#else
static __always_inline int check_v4_lb(struct __ctx_buff *ctx __maybe_unused)
{
	return CTX_ACT_OK;
}
#endif /* ENABLE_NODEPORT_ACCELERATION */

// ENABLE_PREFILTER 暂时默认不开启
#ifdef ENABLE_PREFILTER

static __always_inline int 
check_v4(struct __ctx_buff *ctx) {
// 	void *data_end = ctx_data_end(ctx);
// 	void *data = ctx_data(ctx);
// 	struct iphdr *ipv4_hdr = data + sizeof(struct ethhdr);
// 	struct lpm_v4_key pfx __maybe_unused;

// 	if (ctx_no_room(ipv4_hdr + 1, data_end))
// 		return CTX_ACT_DROP;

// #ifdef CIDR4_FILTER
// 	memcpy(pfx.lpm.data, &ipv4_hdr->saddr, sizeof(pfx.addr));
// 	pfx.lpm.prefixlen = 32;

// #ifdef CIDR4_LPM_PREFILTER
// 	if (map_lookup_elem(&CIDR4_LMAP_NAME, &pfx))
// 		return CTX_ACT_DROP;
// #endif /* CIDR4_LPM_PREFILTER */
// 	return map_lookup_elem(&CIDR4_HMAP_NAME, &pfx) ?
// 		CTX_ACT_DROP : check_v4_lb(ctx);
// #else
	return check_v4_lb(ctx);
// #endif /* CIDR4_FILTER */
}
#else
static __always_inline int 
check_v4(struct __ctx_buff *ctx) {
	return check_v4_lb(ctx);
}

#endif /* ENABLE_PREFILTER */
#endif /* ENABLE_IPV4 */



static __always_inline int check_filters(struct xdp_md* ctx) {
    int ret = CTX_ACT_OK;
	__u16 proto;

    if (!validate_ethertype(ctx, &proto))
		return CTX_ACT_OK;

	ctx_store_meta(ctx, XFER_MARKER, 0);
	bpf_skip_nodeport_clear(ctx);

    switch (proto) {
#ifdef ENABLE_IPV4
	case bpf_htons(ETH_P_IP): // ip 包, 0x0800->__u16
		ret = check_v4(ctx);
		break;
#endif /* ENABLE_IPV4 */
#ifdef ENABLE_IPV6
	// case bpf_htons(ETH_P_IPV6):
		// ret = check_v6(ctx);
		break;
#endif /* ENABLE_IPV6 */
	default:
		break;
	}

    return bpf_xdp_exit(ctx, ret);
}

SEC("from-netdev")
int bpf_xdp_entry(struct xdp_md* ctx) {
    return check_filters(ctx);
}

char _license[] SEC("license") = "GPL";
