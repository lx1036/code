# https://github.com/vincenren/nginx/blob/master/CMakeLists.txt
# 需要添加 CMakeLists.txt

cmake_minimum_required(VERSION 3.6)
project(nginx)

#set(CMAKE_C_STANDARD 99)
set(CMAKE_C_FLAGS "-pipe -Wall -Wextra -Wpointer-arith -Wconditional-uninitialized -Wno-unused-parameter
-Wno-deprecated-declarations -Werror -g -DNDK_SET_VAR
-Wno-unused-function -Wno-unused-variable -Wno-undefined-internal -Wno-unused-but-set-variable")

#add_definitions(-DNGX_HTTP_UPSYNC_MODULE=0)

include_directories(src/auto
        src/core
        src/event
        src/http
        src/http/modules
        src/os
        src/os/unix
		src/stream)

set(SOURCE_FILES
    src/core/nginx.c
    src/core/ngx_array.c
    src/core/ngx_buf.c
    src/core/ngx_conf_file.c
    src/core/ngx_connection.c
    src/core/ngx_cpuinfo.c
    src/core/ngx_crc32.c
    src/core/ngx_cycle.c
    src/core/ngx_file.c
    src/core/ngx_hash.c
    src/core/ngx_inet.c
    src/core/ngx_list.c
    src/core/ngx_log.c
    src/core/ngx_module.c
    src/core/ngx_modules.c
    src/core/ngx_palloc.c
    src/core/ngx_parse_time.c
    src/core/ngx_proxy_protocol.c
    src/core/ngx_rbtree.c
    src/core/ngx_regex.c
    src/core/ngx_resolver.c
    src/core/ngx_rwlock.c
    src/core/ngx_shmtx.c
    src/core/ngx_slab.c
    src/core/ngx_string.c
    src/core/ngx_times.c
    src/event/ngx_event.c
    src/event/modules/ngx_kqueue_module.c
    src/event/ngx_event_connect.c
    src/event/ngx_event_openssl.c
    src/event/ngx_event_openssl_stapling.c
    src/event/ngx_event_posted.c
    src/event/ngx_event_tcp.c
    src/event/ngx_event_timer.c
    src/event/ngx_event_udp.c
    src/os/unix/ngx_alloc.c
    src/os/unix/ngx_channel.c
    src/os/unix/ngx_darwin_init.c
    src/os/unix/ngx_darwin_sendfile_chain.c
    src/os/unix/ngx_dlopen.c
    src/os/unix/ngx_errno.c
    src/os/unix/ngx_files.c
    src/os/unix/ngx_pipe.c
    src/os/unix/ngx_posix_init.c
    src/os/unix/ngx_process.c
    src/os/unix/ngx_process_cycle.c
    src/os/unix/ngx_readv_chain.c
    src/os/unix/ngx_recv.c
    src/os/unix/ngx_send.c
    src/os/unix/ngx_setproctitle.c
    src/os/unix/ngx_shmem.c
    src/os/unix/ngx_socket.c
    src/os/unix/ngx_time.c
    src/os/unix/ngx_udp_recv.c
    src/os/unix/ngx_udp_send.c
    src/os/unix/ngx_udp_sendmsg_chain.c
    src/os/unix/ngx_writev_chain.c
    src/stream/ngx_stream.c
    src/stream/ngx_stream_proxy_module.c
    src/stream/ngx_stream_return_module.c
    src/stream/ngx_stream_script.c
    src/stream/ngx_stream_ssl_module.c
    src/stream/ngx_stream_ssl_preread_module.c
    src/stream/ngx_stream_upstream.c
    src/stream/ngx_stream_upstream_round_robin.c
    src/core/ngx_proxy_protocol.h
		src/stream/ngx_stream_log_module.c
        src/http/ngx_http.c
		src/http/ngx_http_request.h
		src/core/ngx_open_file_cache.h)

add_executable(nginx ${SOURCE_FILES})
target_link_libraries(nginx -lpcre -lz)
