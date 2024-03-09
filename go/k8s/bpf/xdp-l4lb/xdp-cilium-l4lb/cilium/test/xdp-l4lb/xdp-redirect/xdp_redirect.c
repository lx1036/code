


#include <linux/bpf.h>

#include <bpf/bpf_helpers.h>


SEC("redirect_to_111")
int xdp_redirect_to_111(struct xdp_md *xdp)
{
    return (int) bpf_redirect(111, 0);
}

SEC("redirect_to_222")
int xdp_redirect_to_222(struct xdp_md *xdp)
{
    return (int) bpf_redirect(222, 0);
}


char _license[] SEC("license") = "GPL";
