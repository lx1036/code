


#ifndef __LIB_PCAP_H_
#define __LIB_PCAP_H_



static __always_inline void
cilium_capture_out(struct __ctx_buff *ctx __maybe_unused)
{
#ifdef ENABLE_CAPTURE
    __u32 cap_len;
	__u16 rule_id;

	/* cilium_capture_out() is always paired with cilium_capture_in(), so
	 * we can rely on previous cached result on whether to push the pkt
	 * to the RB or not.
	 */
	if (cilium_capture_cached(ctx, &rule_id, &cap_len))
		__cilium_capture_out(ctx, rule_id, cap_len);
#endif /* ENABLE_CAPTURE */
}



#endif //__LIB_PCAP_H_
