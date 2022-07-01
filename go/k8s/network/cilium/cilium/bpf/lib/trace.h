//
// Created by 刘祥 on 7/1/22.
//

#ifndef __LIB_TRACE__
#define __LIB_TRACE__




/* Available observation points. */
enum {
    TRACE_TO_LXC,
    TRACE_TO_PROXY,
    TRACE_TO_HOST,
    TRACE_TO_STACK,
    TRACE_TO_OVERLAY,
    TRACE_FROM_LXC,
    TRACE_FROM_PROXY,
    TRACE_FROM_HOST,
    TRACE_FROM_STACK,
    TRACE_FROM_OVERLAY,
    TRACE_FROM_NETWORK,
};


#endif //__LIB_TRACE__
