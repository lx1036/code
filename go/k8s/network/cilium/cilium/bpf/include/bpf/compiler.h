//
// Created by 刘祥 on 7/1/22.
//

#ifndef __BPF_COMPILER_H_
#define __BPF_COMPILER_H_

#ifndef __non_bpf_context
# include "stddef.h"
#endif

#ifndef __section
# define __section(X) __attribute__((section(X), used))
#endif

#ifndef __maybe_unused
# define __maybe_unused __attribute__((__unused__))
#endif


#undef __always_inline /* stddef.h defines its own */
#define __always_inline inline __attribute__((always_inline))

#ifndef __packed
# define __packed		__attribute__((packed))
#endif

/* {READ,WRITE}_ONCE() with verifier workaround via bpf_barrier(). */

#ifndef READ_ONCE
# define READ_ONCE(X)						\
			({ typeof(X) __val = __READ_ONCE(X);	\
			   bpf_barrier();			\
			   __val; })
#endif

#endif //__BPF_COMPILER_H_
