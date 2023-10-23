





#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>


SEC("xdp")
int  xdp_pass_func(struct xdp_md *ctx)
{
	return XDP_PASS;
}

SEC("xdp")
int  xdp_drop_func(struct xdp_md *ctx)
{
	return XDP_DROP;
}



char _license[] SEC("license") = "GPL";