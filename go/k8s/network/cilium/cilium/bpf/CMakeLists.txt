cmake_minimum_required(VERSION 3.21)
project(bpf C)

set(CMAKE_C_STANDARD 11)

include_directories(include)
include_directories(include/bpf)
include_directories(include/bpf/ctx)
include_directories(include/linux)
include_directories(lib)

add_executable(bpf
        include/bpf/ctx/ctx.h
        include/bpf/ctx/skb.h
        include/bpf/api.h
        include/linux/bpf.h
        lib/overloadable.h
        lib/overloadable_skb.h
        lib/overloadable_xdp.h
        bpf_lxc.c)
