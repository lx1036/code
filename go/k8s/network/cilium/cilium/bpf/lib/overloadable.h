//
// Created by 刘祥 on 7/1/22.
//

#ifndef __LIB_OVERLOADABLE_H_
#define __LIB_OVERLOADABLE_H_




#if __ctx_is == __ctx_skb
# include "lib/overloadable_skb.h"
#else
# include "lib/overloadable_xdp.h"
#endif


#endif //__LIB_OVERLOADABLE_H_
