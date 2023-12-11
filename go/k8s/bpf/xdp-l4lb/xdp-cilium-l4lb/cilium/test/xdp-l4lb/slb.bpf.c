

#include "vmlinux.h"

// /root/linux-5.10.142/tools/lib/bpf/bpf_tracing.h
#include <bpf/bpf_tracing.h>

// /root/linux-5.10.142/tools/lib/bpf/bpf_endian.h
#include <bpf/bpf_endian.h>

// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
// 它定义了一些用于 BPF（Berkeley Packet Filter）程序的辅助函数，例如用于操作 BPF maps、发送事件到用户空间、获取当前时间戳等
#include <bpf/bpf_helpers.h>


/**
 * __attribute__((always_inline)) 是 GNU C 的一个特性，用来告诉编译器尽可能将某个函数内联
 * 在 C 语言中，函数调用通常会有一些额外的开销，比如参数传递、栈帧管理等。为了优化这些开销，
 * 编译器有时候会选择将小的、调用频繁的函数"内联"，即直接将函数体插入到调用它的地方，以减少函数调用的开销。
 */
#undef __always_inline
#define __always_inline inline __attribute__((always_inline))



__always_inline
static void l4_ingress() {

}




