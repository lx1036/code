//
// Created by 刘祥 on 7/1/22.
//

#ifndef __BPF_SECTION__
#define __BPF_SECTION__

#include "compiler.h"




#ifndef BPF_LICENSE
# define BPF_LICENSE(NAME)				\
	char ____license[] __section_license = NAME
#endif

#endif //__BPF_SECTION__
