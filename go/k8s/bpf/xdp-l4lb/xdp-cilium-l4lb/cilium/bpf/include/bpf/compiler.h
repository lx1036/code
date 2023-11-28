







#ifndef __BPF_COMPILER_H_
#define __BPF_COMPILER_H_

#ifndef __non_bpf_context
#include "stddef.h"
#endif

#ifndef __section
#define __section(X) __attribute__((section(X), used))
#endif

#ifndef __maybe_unused
#define __maybe_unused		__attribute__((__unused__))
#endif

#ifndef offsetof
#define offsetof(T, M)		__builtin_offsetof(T, M)
#endif

#ifndef field_sizeof
#define field_sizeof(T, M)	sizeof((((T *)NULL)->M))
#endif

// 是GCC编译器的一个特性，它用于告诉编译器取消结构体/联合体/枚举类型变量在内存中的对齐，而按照实际占用的字节数进行内存布局
/*
 * struct A {
    char a;
    int b;
} __attribute__((packed));
在默认情况下，由于int类型的对齐要求，结构体A的大小可能会为8字节。但使用了__attribute__((packed))后，结构体A的大小就会为5字节，节省了内存空间
在大多数情况下，int类型占用4字节（32位）
 */
#ifndef __packed
#define __packed		__attribute__((packed))
#endif

#ifndef __nobuiltin
#if __clang_major__ >= 10
#define __nobuiltin(X) __attribute__((no_builtin(X)))
#else
#define __nobuiltin(X)
#endif
#endif

#ifndef likely
#define likely(X)		__builtin_expect(!!(X), 1) // !! true == 1
#endif

#ifndef unlikely
#define unlikely(X)		__builtin_expect(!!(X), 0) // !! false == 0
#endif

#ifndef always_succeeds		/* Mainly for documentation purpose. */
#define always_succeeds(X)	likely(X)
#endif

#undef __always_inline		/* stddef.h defines its own */
#define __always_inline		inline __attribute__((always_inline))

#ifndef __stringify
#define __stringify(X)		#X
#endif

#ifndef __fetch
#define __fetch(X)		(__u32)(__u64)(&(X))
#endif

#ifndef __aligned
#define __aligned(X)		__attribute__((aligned(X)))
#endif

#ifndef build_bug_on
#define build_bug_on(E)	((void)sizeof(char[1 - 2*!!(E)]))
#endif

#ifndef __throw_build_bug
#define __throw_build_bug()	__builtin_trap()
#endif

#ifndef __printf
#define __printf(X, Y)		__attribute__((__format__(printf, X, Y)))
#endif

#ifndef barrier
#define barrier() asm volatile("": : :"memory")
#endif

#ifndef barrier_data
# define barrier_data(ptr)	asm volatile("": :"r"(ptr) :"memory")
#endif

static __always_inline void bpf_barrier(void)
{
	/* Workaround to avoid verifier complaint:
	 * "dereference of modified ctx ptr R5 off=48+0, ctx+const is allowed,
	 *        ctx+const+const is not"
	 */
	barrier();
}

#ifndef ARRAY_SIZE
# define ARRAY_SIZE(A)		(sizeof(A) / sizeof((A)[0]))
#endif

#ifndef __READ_ONCE
# define __READ_ONCE(X)		(*(volatile typeof(X) *)&X)
#endif

#ifndef __WRITE_ONCE
# define __WRITE_ONCE(X, V)	(*(volatile typeof(X) *)&X) = (V)
#endif

/* {READ,WRITE}_ONCE() with verifier workaround via bpf_barrier(). */

#ifndef READ_ONCE
# define READ_ONCE(X)						\
			({ typeof(X) __val = __READ_ONCE(X);	\
			   bpf_barrier();			\
			   __val; })
#endif

#ifndef WRITE_ONCE
# define WRITE_ONCE(X, V)					\
				({ typeof(X) __val = (V);	\
				   __WRITE_ONCE(X, __val);	\
				   bpf_barrier();		\
				   __val; })
#endif

#endif /* __BPF_COMPILER_H_ */


