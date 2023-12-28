


#include <stdbool.h>
#include <stddef.h>
#include <string.h>


#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#define NO_FLAGS 0

struct {
	__uint(type, BPF_MAP_TYPE_XSKMAP);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(__u32));
	__uint(max_entries, 64);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(map_flags, NO_FLAGS);
} xsks_map SEC(".maps");


struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(__u32));
	__uint(max_entries, 64);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(map_flags, NO_FLAGS);
} xdp_stats_map SEC(".maps");



SEC("xdp")
int xdp_prog(struct xdp_md *ctx)
{
    int index = ctx->rx_queue_index;

    __u32 *packet_count;
    packet_count = bpf_map_lookup_elem(&xdp_stats_map, &index);
    if (packet_count) {
        if ((*packet_count)++ & 1)
            return XDP_PASS;
    }

    // A set entry here means that the correspnding queue_id has an active AF_XDP socket bound to it
    if (bpf_map_lookup_elem(&xsks_map, &index)) {
        return bpf_redirect_map(&xsks_map, index, 0); // 这里 index 不是指针
    }
    
	return XDP_PASS;
}



char _license[] SEC("license") = "GPL";
