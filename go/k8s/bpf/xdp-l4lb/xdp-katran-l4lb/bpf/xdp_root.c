
#include <stddef.h>

#include <linux/bpf.h>
#include <linux/types.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>



#define ROOT_ARRAY_SIZE 3

struct {
	__uint(type, BPF_MAP_TYPE_PROG_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(__u32));
	__uint(max_entries, ROOT_ARRAY_SIZE);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
} root_array SEC(".maps");


SEC("xdp")
int xdp_root(struct xdp_md* ctx) {
    __u32* fd;

// Clang 编译器的一个指令，用于指示编译器在循环展开时完全展开循环
#pragma clang loop unroll(full)
    for (__u32 i = 0; i < ROOT_ARRAY_SIZE; i++) {
        bpf_tail_call(ctx, &root_array, i);
    }
    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
