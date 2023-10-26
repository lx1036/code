

#include <stddef.h>

#include <linux/bpf.h>
#include <linux/types.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>







SEC("tc")
int healthcheck_encap(struct __sk_buff* skb) {

}

char _license[] SEC("license") = "GPL";