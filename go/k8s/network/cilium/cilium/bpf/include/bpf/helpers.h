//
// Created by 刘祥 on 7/1/22.
//

#ifndef BPF_HELPERS_H
#define BPF_HELPERS_H




#ifndef BPF_FUNC
# define BPF_FUNC(NAME, ...)						\
	(* NAME)(__VA_ARGS__) __maybe_unused = (void *)BPF_FUNC_##NAME
#endif





#endif //BPF_HELPERS_H
