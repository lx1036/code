

#include <stddef.h>
#include <stdbool.h>
#include <string.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>
#include <linux/tcp.h>
#include <sys/socket.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>


/**
* bpftool gen skeleton bpf_bpfel.o > bpf_bpfel.skel.h
* readelf -a bpf_bpfel.o
*/
// https://cilium.slack.com/archives/C4XCTGYEM/p1709030644794629?thread_ts=1709025000.815169&cid=C4XCTGYEM
__u32 inherit_cb_flags = 0;
//static volatile __u32 inherit_cb_flags = 0;

struct bpf_test_option {
    __u8 flags;
    __u8 max_delack_ms;
    __u8 rand;
} __attribute__((packed));

// rewrite const in userspace
static volatile const struct bpf_test_option passive_synack_out = {};


SEC("xdp/test_rewrite_const")
int test_rewrite_const(struct xdp_md *ctx) {
    bpf_printk("flags: %d, max_delack_ms:%d, rand:%d", passive_synack_out.flags, passive_synack_out.max_delack_ms,
               passive_synack_out.rand);

    inherit_cb_flags = 100;

    return XDP_PASS;
}


int _version SEC("version") = 1;
char _license[] SEC("license") = "GPL";
