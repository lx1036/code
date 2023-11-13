



#ifndef __NODEPORT_H_
#define __NODEPORT_H_

#include <bpf/ctx/ctx.h>
#include <bpf/api.h>

#include "tailcall.h"
#include "nat.h"
// #include "edt.h"
#include "lb.h"
#include "common.h"
#include "overloadable.h"
#include "egress_policies.h"
#include "eps.h"
#include "conntrack.h"
#include "csum.h"
#include "encap.h"
#include "trace.h"
#include "ghash.h"
#include "pcap.h"
// #include "host_firewall.h"



#ifndef DSR_ENCAP_MODE
#define DSR_ENCAP_MODE 0
#define DSR_ENCAP_IPIP 2
#endif


static __always_inline 
bool nodeport_uses_dsr(__u8 nexthdr __maybe_unused)
{
# if defined(ENABLE_DSR) && !defined(ENABLE_DSR_HYBRID)
	return true;
# elif defined(ENABLE_DSR) && defined(ENABLE_DSR_HYBRID)
	if (nexthdr == IPPROTO_TCP)
		return true;
	return false;
# else
	return false;
# endif
}



#ifdef ENABLE_IPV4

static __always_inline bool nodeport_uses_dsr4(const struct ipv4_ct_tuple *tuple)
{
	return nodeport_uses_dsr(tuple->nexthdr);
}

// https://cilium.io/blog/2020/06/22/cilium-18/
__section_tail(CILIUM_MAP_CALLS, CILIUM_CALL_IPV4_NODEPORT_NAT)
int tail_nodeport_nat_ipv4(struct __ctx_buff *ctx)
{
	int ret, dir = ctx_load_meta(ctx, CB_NAT);
	struct bpf_fib_lookup_padded fib_params = {
		.l = {
			.family		= AF_INET,
			.ifindex	= DIRECT_ROUTING_DEV_IFINDEX,
		},
	};
	struct ipv4_nat_target target = {
		.min_port = NODEPORT_PORT_MIN_NAT,
		.max_port = NODEPORT_PORT_MAX_NAT,
		.src_from_world = true,
	};
	union macaddr *dmac = NULL;
	void *data, *data_end;
	struct iphdr *ip4;
	bool l2_hdr_required = true;

	target.addr = IPV4_DIRECT_ROUTING;
#ifdef TUNNEL_MODE
	if (dir == NAT_DIR_EGRESS) {
		struct remote_endpoint_info *info;

		if (!revalidate_data(ctx, &data, &data_end, &ip4)) {
			ret = DROP_INVALID;
			goto drop_err;
		}

		info = ipcache_lookup4(&IPCACHE_MAP, ip4->daddr, V4_CACHE_KEY_LEN);
		if (info != NULL && info->tunnel_endpoint != 0) {
			ret = __encap_with_nodeid(ctx, info->tunnel_endpoint,
			/* The dir == NAT_DIR_EGRESS branch is executed for
			 * N/S LB requests which needs to be fwd-ed to a remote
			 * node. As the request came from outside, we need to
			 * set the security id in the tunnel header to WORLD_ID.
			 * Otherwise, the remote node will assume, that the
			 * request originated from a cluster node which will
			 * bypass any netpol which disallows LB requests from
			 * outside.
			 */
#ifdef PRESERVE_WORLD_ID
						  WORLD_ID,
#else
						  SECLABEL,
#endif /* PRESERVE_WORLD_ID */
						  TRACE_PAYLOAD_LEN);
			if (ret)
				goto drop_err;

			target.addr = IPV4_GATEWAY;
			fib_params.l.ifindex = ENCAP_IFINDEX;

			/* fib lookup not necessary when going over tunnel. */
			if (eth_store_daddr(ctx, fib_params.l.dmac, 0) < 0) {
				ret = DROP_WRITE_ERROR;
				goto drop_err;
			}
			if (eth_store_saddr(ctx, fib_params.l.smac, 0) < 0) {
				ret = DROP_WRITE_ERROR;
				goto drop_err;
			}
		}
	}
#endif
	/* Handles SNAT on NAT_DIR_EGRESS and reverse SNAT for reply packets
	 * from remote backends on NAT_DIR_INGRESS.
	 */
	ret = snat_v4_process(ctx, dir, &target, false);
	if (IS_ERR(ret)) {
		/* In case of no mapping, recircle back to main path. SNAT is very
		 * expensive in terms of instructions (since we don't have BPF to
		 * BPF calls as we use tail calls) and complexity, hence this is
		 * done inside a tail call here.
		 */
		if (dir == NAT_DIR_INGRESS) {
			bpf_skip_nodeport_set(ctx);
			ep_tail_call(ctx, CILIUM_CALL_IPV4_FROM_LXC);
			ret = DROP_MISSED_TAIL_CALL;
			goto drop_err;
		}
		if (ret != NAT_PUNT_TO_STACK)
			goto drop_err;
	}

	bpf_mark_snat_done(ctx);

	if (dir == NAT_DIR_INGRESS) {
		/* Handle reverse DNAT for reply packets from remote backends. */
		ep_tail_call(ctx, CILIUM_CALL_IPV4_NODEPORT_REVNAT);
		ret = DROP_MISSED_TAIL_CALL;
		goto drop_err;
	}
#ifdef TUNNEL_MODE
	if (fib_params.l.ifindex == ENCAP_IFINDEX)
		goto out_send;
#endif
	if (!revalidate_data(ctx, &data, &data_end, &ip4)) {
		ret = DROP_INVALID;
		goto drop_err;
	}

	ret = maybe_add_l2_hdr(ctx, DIRECT_ROUTING_DEV_IFINDEX,
			       &l2_hdr_required);
	if (ret != 0)
		goto drop_err;
	if (!l2_hdr_required)
		goto out_send;
	else if (!revalidate_data_with_eth_hlen(ctx, &data, &data_end, &ip4,
						__ETH_HLEN))
		return DROP_INVALID;

	if (nodeport_lb_hairpin())
		dmac = map_lookup_elem(&NODEPORT_NEIGH4, &ip4->daddr);
	if (dmac) {
		union macaddr mac = NATIVE_DEV_MAC_BY_IFINDEX(fib_params.l.ifindex);

		if (eth_store_daddr_aligned(ctx, dmac->addr, 0) < 0) {
			ret = DROP_WRITE_ERROR;
			goto drop_err;
		}
		if (eth_store_saddr_aligned(ctx, mac.addr, 0) < 0) {
			ret = DROP_WRITE_ERROR;
			goto drop_err;
		}
	} else {
		fib_params.l.ipv4_src = ip4->saddr;
		fib_params.l.ipv4_dst = ip4->daddr;

		ret = fib_lookup(ctx, &fib_params.l, sizeof(fib_params),
				 BPF_FIB_LOOKUP_DIRECT | BPF_FIB_LOOKUP_OUTPUT);
		if (ret != 0) {
			ret = DROP_NO_FIB;
			goto drop_err;
		}
		if (nodeport_lb_hairpin())
			map_update_elem(&NODEPORT_NEIGH4, &ip4->daddr,
					fib_params.l.dmac, 0);

		if (eth_store_daddr(ctx, fib_params.l.dmac, 0) < 0) {
			ret = DROP_WRITE_ERROR;
			goto drop_err;
		}
		if (eth_store_saddr(ctx, fib_params.l.smac, 0) < 0) {
			ret = DROP_WRITE_ERROR;
			goto drop_err;
		}
	}
out_send:
	cilium_capture_out(ctx);
	return ctx_redirect(ctx, fib_params.l.ifindex, 0);
drop_err:
	return send_drop_notify_error(ctx, 0, ret, CTX_ACT_DROP,
				      dir == NAT_DIR_INGRESS ?
				      METRIC_INGRESS : METRIC_EGRESS);
}



#ifdef ENABLE_DSR

__section_tail(CILIUM_MAP_CALLS, CILIUM_CALL_IPV4_NODEPORT_DSR)
int tail_nodeport_ipv4_dsr(struct __ctx_buff *ctx) {
    struct bpf_fib_lookup_padded fib_params = {
		.l = {
			.family		= AF_INET,
			.ifindex	= DIRECT_ROUTING_DEV_IFINDEX,
		},
	};
	union macaddr *dmac = NULL;
	void *data, *data_end;
	int ret, ohead = 0;
	struct iphdr *ip4;
	bool l2_hdr_required = true;

	if (!revalidate_data(ctx, &data, &data_end, &ip4)) {
		ret = DROP_INVALID;
		goto drop_err;
	}

#if DSR_ENCAP_MODE == DSR_ENCAP_IPIP
	ret = dsr_set_ipip4(ctx, ip4,
			    ctx_load_meta(ctx, CB_ADDR_V4),
			    ctx_load_meta(ctx, CB_HINT), &ohead);
#elif DSR_ENCAP_MODE == DSR_ENCAP_NONE
	ret = dsr_set_opt4(ctx, ip4,
			   ctx_load_meta(ctx, CB_ADDR_V4),
			   ctx_load_meta(ctx, CB_PORT), &ohead);
#else
# error "Invalid load balancer DSR encapsulation mode!"
#endif
	if (unlikely(ret)) {
		if (dsr_fail_needs_reply(ret))
			return dsr_reply_icmp4(ctx, ip4, ret, ohead);
		goto drop_err;
	}
	if (!revalidate_data(ctx, &data, &data_end, &ip4)) {
		ret = DROP_INVALID;
		goto drop_err;
	}

	ret = maybe_add_l2_hdr(ctx, DIRECT_ROUTING_DEV_IFINDEX,
			       &l2_hdr_required);
	if (ret != 0)
		goto drop_err;
	if (!l2_hdr_required)
		goto out_send;
	else if (!revalidate_data_with_eth_hlen(ctx, &data, &data_end, &ip4,
						__ETH_HLEN))
		return DROP_INVALID;

	if (nodeport_lb_hairpin())
		dmac = map_lookup_elem(&NODEPORT_NEIGH4, &ip4->daddr);
	if (dmac) {
		union macaddr mac = NATIVE_DEV_MAC_BY_IFINDEX(fib_params.l.ifindex);

		if (eth_store_daddr_aligned(ctx, dmac->addr, 0) < 0) {
			ret = DROP_WRITE_ERROR;
			goto drop_err;
		}
		if (eth_store_saddr_aligned(ctx, mac.addr, 0) < 0) {
			ret = DROP_WRITE_ERROR;
			goto drop_err;
		}
	} else {
		fib_params.l.ipv4_src = ip4->saddr;
		fib_params.l.ipv4_dst = ip4->daddr;

		ret = fib_lookup(ctx, &fib_params.l, sizeof(fib_params),
				 BPF_FIB_LOOKUP_DIRECT | BPF_FIB_LOOKUP_OUTPUT);
		if (ret != 0) {
			ret = DROP_NO_FIB;
			goto drop_err;
		}
		if (nodeport_lb_hairpin())
			map_update_elem(&NODEPORT_NEIGH4, &ip4->daddr,
					fib_params.l.dmac, 0);
		if (eth_store_daddr(ctx, fib_params.l.dmac, 0) < 0) {
			ret = DROP_WRITE_ERROR;
			goto drop_err;
		}
		if (eth_store_saddr(ctx, fib_params.l.smac, 0) < 0) {
			ret = DROP_WRITE_ERROR;
			goto drop_err;
		}
	}

out_send:
	cilium_capture_out(ctx);
	return ctx_redirect(ctx, fib_params.l.ifindex, 0);
drop_err:
	return send_drop_notify_error(ctx, 0, ret, CTX_ACT_DROP, METRIC_EGRESS);
}


#endif /* ENABLE_DSR */



/* Main node-port entry point for host-external ingressing node-port traffic
 * which handles the case of: i) backend is local EP, ii) backend is remote EP,
 * iii) reply from remote backend EP.
 */
static __always_inline int nodeport_lb4(struct __ctx_buff *ctx,
					__u32 src_identity)
{
	struct ipv4_ct_tuple tuple = {};
	void *data, *data_end;
	struct iphdr *ip4;
	int ret,  l3_off = ETH_HLEN, l4_off;
	struct csum_offset csum_off = {};
	struct lb4_service *svc;
	struct lb4_key key = {};
	struct ct_state ct_state_new = {};
	union macaddr smac, *mac;
	bool backend_local;
	__u32 monitor = 0;

	cilium_capture_in(ctx);

	if (!revalidate_data(ctx, &data, &data_end, &ip4))
		return DROP_INVALID;

	tuple.nexthdr = ip4->protocol;
	tuple.daddr = ip4->daddr;
	tuple.saddr = ip4->saddr;

	l4_off = l3_off + ipv4_hdrlen(ip4);

	ret = lb4_extract_key(ctx, ip4, l4_off, &key, &csum_off, CT_EGRESS);
	if (IS_ERR(ret)) {
		if (ret == DROP_NO_SERVICE)
			goto skip_service_lookup;
		else if (ret == DROP_UNKNOWN_L4)
			return CTX_ACT_OK;
		else
			return ret;
	}

	// 1. 从 cilium_lb4_services_v2 map 里查找 service/backends
	svc = lb4_lookup_service(&key, false);
	if (svc) {
		const bool skip_l3_xlate = DSR_ENCAP_MODE == DSR_ENCAP_IPIP;

		if (!lb4_src_range_ok(svc, ip4->saddr))
			return DROP_NOT_IN_SRC_RANGE;

		ret = lb4_local(get_ct_map4(&tuple), ctx, l3_off, l4_off,
				&csum_off, &key, &tuple, svc, &ct_state_new,
				ip4->saddr, ipv4_has_l4_header(ip4),
				skip_l3_xlate);
		if (IS_ERR(ret))
			return ret;
	}

	if (!svc || !lb4_svc_is_routable(svc)) {
		if (svc)
			return DROP_IS_CLUSTER_IP;

		/* The packet is not destined to a service but it can be a reply
		 * packet from a remote backend, in which case we need to perform
		 * the reverse NAT.
		 */
skip_service_lookup:
		ctx_set_xfer(ctx, XFER_PKT_NO_SVC);

#ifndef ENABLE_MASQUERADE
		if (nodeport_uses_dsr4(&tuple))
			return CTX_ACT_OK;
#endif

		ctx_store_meta(ctx, CB_NAT, NAT_DIR_INGRESS);
		ctx_store_meta(ctx, CB_SRC_IDENTITY, src_identity);
		ep_tail_call(ctx, CILIUM_CALL_IPV4_NODEPORT_NAT);
		return DROP_MISSED_TAIL_CALL;
	}

	backend_local = __lookup_ip4_endpoint(tuple.daddr);
	if (!backend_local && lb4_svc_is_hostport(svc))
		return DROP_INVALID;

	/* Reply from DSR packet is never seen on this node again hence no
	 * need to track in here.
	 */
	// backend_local 不可能为 true，lb 节点不会有 local backend
	if (backend_local || !nodeport_uses_dsr4(&tuple)) {
		struct ct_state ct_state = {};

		ret = ct_lookup4(get_ct_map4(&tuple), &tuple, ctx, l4_off,
				 CT_EGRESS, &ct_state, &monitor);
		switch (ret) {
		case CT_NEW:
redo_all:
#ifdef PRESERVE_WORLD_ID
			ct_state_new.src_sec_id = WORLD_ID;
#else
			ct_state_new.src_sec_id = SECLABEL;
#endif /* PRESERVE_WORLD_ID */
			ct_state_new.node_port = 1;
			ct_state_new.ifindex = NATIVE_DEV_IFINDEX;
			ret = ct_create4(get_ct_map4(&tuple), NULL, &tuple, ctx,
					 CT_EGRESS, &ct_state_new, false);
			if (IS_ERR(ret))
				return ret;
			if (backend_local) {
				ct_flip_tuple_dir4(&tuple);
redo_local:
				/* Reset rev_nat_index, otherwise ipv4_policy()
				 * in bpf_lxc will do invalid xlation.
				 */
				ct_state_new.rev_nat_index = 0;
				ret = ct_create4(get_ct_map4(&tuple), NULL,
						 &tuple, ctx, CT_INGRESS,
						 &ct_state_new, false);
				if (IS_ERR(ret))
					return ret;
			}
			break;
		case CT_REOPENED:
		case CT_ESTABLISHED:
		case CT_REPLY:
			/* Recreate CT entries, as the existing one is stale and
			 * belongs to a flow which target a different svc.
			 */
			if (unlikely(ct_state.rev_nat_index !=
				     svc->rev_nat_index))
				goto redo_all;
			if (backend_local) {
				ct_flip_tuple_dir4(&tuple);
				if (!__ct_entry_keep_alive(get_ct_map4(&tuple),
							   &tuple)) {
#ifdef PRESERVE_WORLD_ID
					ct_state_new.src_sec_id = WORLD_ID;
#else
					ct_state_new.src_sec_id = SECLABEL;
#endif /* PRESERVE_WORLD_ID */
					ct_state_new.node_port = 1;
					ct_state_new.ifindex = NATIVE_DEV_IFINDEX;
					goto redo_local;
				}
			}
			break;
		default:
			return DROP_UNKNOWN_CT;
		}

		if (!revalidate_data(ctx, &data, &data_end, &ip4))
			return DROP_INVALID;
		if (eth_load_saddr(ctx, smac.addr, 0) < 0)
			return DROP_INVALID;

		mac = map_lookup_elem(&NODEPORT_NEIGH4, &ip4->saddr);
		if (!mac || eth_addrcmp(mac, &smac)) {
			ret = map_update_elem(&NODEPORT_NEIGH4, &ip4->saddr,
					      &smac, 0);
			if (ret < 0)
				return ret;
		}
	}

	if (!backend_local) {
		// edt_set_aggregate(ctx, 0);
		if (nodeport_uses_dsr4(&tuple)) {
#if DSR_ENCAP_MODE == DSR_ENCAP_IPIP
			ctx_store_meta(ctx, CB_HINT, ((__u32)tuple.sport << 16) | tuple.dport);
			ctx_store_meta(ctx, CB_ADDR_V4, tuple.daddr);
#elif DSR_ENCAP_MODE == DSR_ENCAP_NONE
			ctx_store_meta(ctx, CB_PORT, key.dport);
			ctx_store_meta(ctx, CB_ADDR_V4, key.address);
#endif /* DSR_ENCAP_MODE */
			ep_tail_call(ctx, CILIUM_CALL_IPV4_NODEPORT_DSR);
		} else {
			ctx_store_meta(ctx, CB_NAT, NAT_DIR_EGRESS);
			ep_tail_call(ctx, CILIUM_CALL_IPV4_NODEPORT_NAT);
		}
		return DROP_MISSED_TAIL_CALL;
	}

	ctx_set_xfer(ctx, XFER_PKT_NO_SVC);

	return CTX_ACT_OK;
}



#endif /* ENABLE_IPV4 */



#endif /* __NODEPORT_H_ */