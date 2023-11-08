



#ifndef __BPF_BUILTINS__
#define __BPF_BUILTINS__

#include "compiler.h"

#ifndef __non_bpf_context

#ifndef lock_xadd
# define lock_xadd(P, V)	((void) __sync_fetch_and_add((P), (V)))
#endif


/* Memory iterators used below. */
#define __it_bwd(x, op) (x -= sizeof(__u##op))
#define __it_fwd(x, op) (x += sizeof(__u##op))

/* Memory operators used below. */
#define __it_set(a, op) (*(__u##op *)__it_bwd(a, op)) = 0
#define __it_xor(a, b, r, op) r |= (*(__u##op *)__it_bwd(a, op)) ^ (*(__u##op *)__it_bwd(b, op))
#define __it_mob(a, b, op) (*(__u##op *)__it_bwd(a, op)) = (*(__u##op *)__it_bwd(b, op))
#define __it_mof(a, b, op)				\
	do {						\
		*(__u##op *)a = *(__u##op *)b;		\
		__it_fwd(a, op); __it_fwd(b, op);	\
	} while (0)



static __always_inline void __bpf_memcpy(void *d, const void *s, __u64 len)
{
#if __clang_major__ >= 10
	if (!__builtin_constant_p(len))
		__throw_build_bug();

	d += len;
	s += len;

	switch (len) {
	case 96:         __it_mob(d, s, 64);
	case 88: jmp_88: __it_mob(d, s, 64);
	case 80: jmp_80: __it_mob(d, s, 64);
	case 72: jmp_72: __it_mob(d, s, 64);
	case 64: jmp_64: __it_mob(d, s, 64);
	case 56: jmp_56: __it_mob(d, s, 64);
	case 48: jmp_48: __it_mob(d, s, 64);
	case 40: jmp_40: __it_mob(d, s, 64);
	case 32: jmp_32: __it_mob(d, s, 64);
	case 24: jmp_24: __it_mob(d, s, 64);
	case 16: jmp_16: __it_mob(d, s, 64);
	case  8: jmp_8:  __it_mob(d, s, 64);
		break;

	case 94: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_88;
	case 86: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_80;
	case 78: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_72;
	case 70: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_64;
	case 62: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_56;
	case 54: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_48;
	case 46: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_40;
	case 38: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_32;
	case 30: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_24;
	case 22: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_16;
	case 14: __it_mob(d, s, 16); __it_mob(d, s, 32); goto jmp_8;
	case  6: __it_mob(d, s, 16); __it_mob(d, s, 32);
		break;

	case 92: __it_mob(d, s, 32); goto jmp_88;
	case 84: __it_mob(d, s, 32); goto jmp_80;
	case 76: __it_mob(d, s, 32); goto jmp_72;
	case 68: __it_mob(d, s, 32); goto jmp_64;
	case 60: __it_mob(d, s, 32); goto jmp_56;
	case 52: __it_mob(d, s, 32); goto jmp_48;
	case 44: __it_mob(d, s, 32); goto jmp_40;
	case 36: __it_mob(d, s, 32); goto jmp_32;
	case 28: __it_mob(d, s, 32); goto jmp_24;
	case 20: __it_mob(d, s, 32); goto jmp_16;
	case 12: __it_mob(d, s, 32); goto jmp_8;
	case  4: __it_mob(d, s, 32);
		break;

	case 90: __it_mob(d, s, 16); goto jmp_88;
	case 82: __it_mob(d, s, 16); goto jmp_80;
	case 74: __it_mob(d, s, 16); goto jmp_72;
	case 66: __it_mob(d, s, 16); goto jmp_64;
	case 58: __it_mob(d, s, 16); goto jmp_56;
	case 50: __it_mob(d, s, 16); goto jmp_48;
	case 42: __it_mob(d, s, 16); goto jmp_40;
	case 34: __it_mob(d, s, 16); goto jmp_32;
	case 26: __it_mob(d, s, 16); goto jmp_24;
	case 18: __it_mob(d, s, 16); goto jmp_16;
	case 10: __it_mob(d, s, 16); goto jmp_8;
	case  2: __it_mob(d, s, 16);
		break;

	case  1: __it_mob(d, s, 8);
		break;

	default:
		/* __builtin_memcpy() is crappy slow since it cannot
		 * make any assumptions about alignment & underlying
		 * efficient unaligned access on the target we're
		 * running.
		 */
		__throw_build_bug();
	}
#else
	__bpf_memcpy_builtin(d, s, len);
#endif
}


static __always_inline __maybe_unused void
__bpf_no_builtin_memcpy(void *d __maybe_unused, const void *s __maybe_unused, __u64 len __maybe_unused)
{
	__throw_build_bug();
}

/* Redirect any direct use in our code to throw an error. */
#define __builtin_memcpy	__bpf_no_builtin_memcpy

static __always_inline __nobuiltin("memcpy") void memcpy(void *d, const void *s, __u64 len)
{
	return __bpf_memcpy(d, s, len);
}






#endif /* __non_bpf_context */

#endif /* __BPF_BUILTINS__ */