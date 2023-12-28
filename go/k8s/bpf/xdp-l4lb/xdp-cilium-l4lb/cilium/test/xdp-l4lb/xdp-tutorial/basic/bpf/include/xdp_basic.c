





#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include <bpf.h>
#include <xdp_basic.h>


SEC("xdp")
int xdp_pass_func(struct xdp_md *ctx) {
	return XDP_PASS;
}

SEC("xdp")
int xdp_drop_func(struct xdp_md *ctx) {
	return XDP_DROP;
}




struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct datarec));
	__uint(max_entries, XDP_ACTION_MAX);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(map_flags, NO_FLAGS);
} xdp_stats_map SEC(".maps");

/* LLVM maps __sync_fetch_and_add() as a built-in function to the BPF atomic add
 * instruction (that is BPF_STX | BPF_XADD | BPF_W for word sizes)
 */
#ifndef lock_xadd
#define lock_xadd(ptr, val)	((void) __sync_fetch_and_add(ptr, val))
#endif

SEC("xdp")
int xdp_stats1_func(struct xdp_md *ctx) {
	__u32 key = XDP_PASS;
	struct datarec *value;
	value = bpf_map_lookup_elem(&xdp_stats_map, &key);
	if (!value) {
		return XDP_ABORTED;
	}

	lock_xadd(&value->rx_packets, 1);

	return XDP_PASS;
}







char _license[] SEC("license") = "GPL";