


#include <stddef.h>

#include <linux/if.h>
#include <linux/if_ether.h>
#include <linux/if_tunnel.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/pkt_cls.h>

#include <linux/bpf.h>
#include <linux/types.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>



#define CTRL_ARRAY_SIZE 2
#define CNTRS_ARRAY_SIZE 512

struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(__u32));
	__uint(max_entries, CTRL_ARRAY_SIZE);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
} ctl_array SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(__u64));
	__uint(max_entries, CNTRS_ARRAY_SIZE);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
} cntrs_array SEC(".maps");


SEC("xdp")
int packet_counter(struct xdp_md* ctx) {
    void* data_end = (void*)(long)ctx->data_end;
    void* data = (void*)(long)ctx->data;
  
    __u32 ctl_flag_pos = 0;
    
    __u32* flag = bpf_map_lookup_elem(&ctl_array, &ctl_flag_pos);
    if (!flag || (*flag == 0)) {
        return XDP_PASS;
    };

    __u32 cntr_pos = 0;
    __u64* cntr_val = bpf_map_lookup_elem(&cntrs_array, &cntr_pos);
    if (cntr_val) {
        *cntr_val += 1;
    };
    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
