

/**
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_map_in_map.c
*/

#include <stddef.h>
#include <linux/bpf.h>
#include <linux/types.h>
#include <bpf/bpf_helpers.h>

// 必须有个模板 map，在 userspace 里会更新掉
struct inner_map_tmp2 {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1);
    __uint(map_flags, 0);
    __uint(key_size, sizeof(__u32)); /** 发现 inner 和 outer key/value 类型必须一致 */
    // __uint(key_size, sizeof(int));
    // __uint(value_size, sizeof(int));
    __uint(value_size, sizeof(__u32));
} inner_map_tmp SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY_OF_MAPS);
    __uint(max_entries, 1);
    __uint(map_flags, 0);
    // BPF_MAP_TYPE_ARRAY_OF_MAPS key 必须是 __u32, @see https://docs.kernel.org/bpf/map_of_maps.html
    __uint(key_size, sizeof(__u32));
    /* must be sizeof(__u32) for map in map */
    __uint(value_size, sizeof(__u32));
    __array(values, struct inner_map_tmp2);
} mim_array SEC(".maps") = {
        .values = {
                [0] = (void *) &inner_map_tmp, // 赋值格式可以指定 index
        }
};

// struct {
//     __uint(type, BPF_MAP_TYPE_HASH_OF_MAPS);
//     __uint(max_entries, 1);
//     __uint(map_flags, 0);
//     // __uint(key_size, sizeof(__u32)); // 注意这里是 int，但是可选择的
//     __uint(key_size, sizeof(__u32)); // 注意这里是 int，但是可选择的
//     /* must be sizeof(__u32) for map in map，保存 inner map 的 map id */
//     __uint(value_size, sizeof(__u32));
//     __array(values, struct inner_map_tmp2);
// } mim_hash SEC(".maps") = {
//     .values = {
//         // [0] = (void *) &inner_map_tmp, // populating map, 赋值格式可以指定 index, 也可以不指定
//         (void *) &inner_map_tmp,
//     }
// };

struct {
    __uint(type, BPF_MAP_TYPE_HASH_OF_MAPS);
    __uint(max_entries, 1);
    __uint(map_flags, 0);
    // __uint(key_size, sizeof(__u32)); // 注意这里是 int，但是可选择的
    __uint(key_size, sizeof(__u32)); // 注意这里是 int，但是可选择的
    /* must be sizeof(__u32) for map in map，保存 inner map 的 map id */
    __uint(value_size, sizeof(__u32));
    __array(values, struct inner_map_tmp2);
} mim_hash SEC(".maps"); // 可以不需要初始化赋值，在 userspace 里赋值


SEC("xdp_mimtest")
int xdp_mimtest0(struct xdp_md *ctx) {
    int value = 123;
    int *value_p;
    int key = 0;

    // 这里指定 struct，否则一直报错
    // "LoadAndAssign err: field XdpMimtest0: program xdp_mimtest0: load program: permission denied: invalid indirect access to stack R3 off=-4 size=8 (51 line(s) omitted)"
    struct inner_map_tmp2 *inner_map;
//    void *inner_map;

    inner_map = bpf_map_lookup_elem(&mim_array, &key);
    if (!inner_map)
        return XDP_DROP;

    // 很奇怪，inner_map 必须指定，不符合直觉，应该是没有定义 inner_map_tmp 导致的。但是最终目的，应该是不需要在 bpf 里指定的才对！！！
    bpf_map_update_elem(inner_map, &key, &value, 0);
    value_p = bpf_map_lookup_elem(inner_map, &key);
    if (!value_p || *value_p != 123)
        return XDP_DROP;

    inner_map = bpf_map_lookup_elem(&mim_hash, &key);
    if (!inner_map)
        return XDP_DROP;

    bpf_map_update_elem(inner_map, &key, &value, 0);

    return XDP_PASS;
}

int _version SEC("version") = 1;
char _license[] SEC("license") = "GPL";
