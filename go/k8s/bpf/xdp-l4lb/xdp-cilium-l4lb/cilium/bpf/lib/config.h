

#ifndef XDP_CILIUM_L4LB_CONFIG_H
#define XDP_CILIUM_L4LB_CONFIG_H


/* Subset of kernel's include/linux/kconfig.h */

#define __ARG_PLACEHOLDER_1 0,
#define __take_second_arg(__ignored, val, ...) val
#define ____is_defined(arg1_or_junk) __take_second_arg(arg1_or_junk 1, 0)
#define ___is_defined(val)           ____is_defined(__ARG_PLACEHOLDER_##val)
#define __is_defined(x)              ___is_defined(x)
#define is_defined(option)           __is_defined(option)


#endif //XDP_CILIUM_L4LB_CONFIG_H
