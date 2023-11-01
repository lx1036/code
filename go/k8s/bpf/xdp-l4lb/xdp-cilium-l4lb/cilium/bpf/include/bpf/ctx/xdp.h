



#ifndef __BPF_CTX_XDP_H_
#define __BPF_CTX_XDP_H_


#define __ctx_buff			xdp_md
#define __ctx_is			__ctx_xdp


#include "common.h"
#include "../helpers_xdp.h"





#define CTX_ACT_OK			XDP_PASS
#define CTX_ACT_DROP			XDP_DROP
#define CTX_ACT_TX			XDP_TX	/* hairpin only */

#define CTX_DIRECT_WRITE_OK		1

					/* cb + RECIRC_MARKER + XFER_MARKER */
#define META_PIVOT			((int)(field_sizeof(struct __sk_buff, cb) + \
					       sizeof(__u32) * 2))

/* This must be a mask and all offsets guaranteed to be less than that. */
#define __CTX_OFF_MAX			0xff










#endif /* __BPF_CTX_XDP_H_ */