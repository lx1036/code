
# 本地运行 `make && make run` 编译自己的 nginx 二进制

CC =	cc
CFLAGS = -I/usr/local/Cellar/openresty/1.21.4.2_1/luajit/include/luajit-2.1 \
	-I/usr/local/Cellar/openresty/1.21.4.2_1/luajit/include/luajit-2.1 \
	-pipe -O -Wall -Wextra -Wpointer-arith -Wconditional-uninitialized \
	-Wno-unused-parameter -Wno-deprecated-declarations -Werror -g -DNDK_SET_VAR \
	-Wno-unused-function -Wno-unused-variable -Wno-undefined-internal
CPP =	cc -E
LINK =	$(CC)


CORE_DEPS = src/core/nginx.h \
	src/core/ngx_log.h \
	src/core/ngx_array.h \
	src/core/ngx_cycle.h \
	src/core/ngx_buf.h \
	src/core/ngx_string.h \
	src/core/ngx_palloc.h \
	src/core/ngx_times.h \
	src/core/ngx_regex.h \
	src/core/ngx_connection.h \
	src/core/ngx_conf_file.h \
	src/core/ngx_rbtree.h \
	src/core/ngx_module.h \
	src/core/ngx_file.h \
	src/core/ngx_parse.h \
	src/core/ngx_crc32.h \
	src/core/ngx_slab.h \
	src/core/ngx_inet.h \
	src/event/ngx_event.h \
	src/event/ngx_event_timer.h \
	src/event/ngx_event_posted.h \
	src/event/ngx_event_connect.h \
	src/event/ngx_event_pipe.h \
	src/event/ngx_event_udp.h \
	src/event/ngx_event_udpv2.h \
	src/os/unix/ngx_time.h \
	src/os/unix/ngx_errno.h \
	src/os/unix/ngx_alloc.h \
	src/os/unix/ngx_files.h \
	src/os/unix/ngx_channel.h \
	src/os/unix/ngx_shmem.h \
	src/os/unix/ngx_process.h \
	src/os/unix/ngx_setaffinity.h \
	src/os/unix/ngx_setproctitle.h \
	src/os/unix/ngx_atomic.h \
	src/os/unix/ngx_gcc_atomic_x86.h \
	src/os/unix/ngx_thread.h \
	src/os/unix/ngx_socket.h \
	src/os/unix/ngx_os.h \
	src/os/unix/ngx_user.h \
	src/os/unix/ngx_pipe.h \
	src/os/unix/ngx_sysinfo.h \
	src/os/unix/ngx_dlopen.h \
	src/os/unix/ngx_process_cycle.h \
	src/os/unix/ngx_darwin_config.h \
	src/os/unix/ngx_darwin.h \
	src/proc/ngx_proc.h \
	src/event/ngx_event_openssl.h \
	src/core/ngx_auto_config.h

CORE_INCS = -I src/core \
	-I src/event \
	-I src/event/modules \
	-I src/os/unix \
	-I src/proc \
	-I objs

# 执行 `make` 默认执行 `make build`
build: binary

binary:	objs/nginx

objs/nginx: objs/src/core/nginx.o \
	objs/src/core/ngx_log.o \
	objs/src/core/ngx_array.o \
	objs/src/core/ngx_cycle.o \
	objs/src/core/ngx_buf.o \
	objs/src/core/ngx_cpuinfo.o \
	objs/src/core/ngx_string.o \
	objs/src/core/ngx_palloc.o \
	objs/src/core/ngx_times.o \
	objs/src/core/ngx_regex.o \
	objs/src/core/ngx_connection.o \
	objs/src/core/ngx_conf_file.o \
	objs/src/core/ngx_rbtree.o \
	objs/src/core/ngx_module.o \
	objs/src/core/ngx_file.o \
	objs/src/core/ngx_parse.o \
	objs/src/core/ngx_crc32.o \
	objs/src/core/ngx_slab.o \
	objs/src/core/ngx_inet.o \
	objs/src/core/ngx_modules.o \
	objs/src/os/unix/ngx_alloc.o \
	objs/src/os/unix/ngx_process.o \
	objs/src/os/unix/ngx_dlopen.o \
	objs/src/os/unix/ngx_errno.o \
	objs/src/os/unix/ngx_posix_init.o \
	objs/src/os/unix/ngx_time.o \
	objs/src/os/unix/ngx_setproctitle.o \
	objs/src/os/unix/ngx_darwin_init.o \
	objs/src/os/unix/ngx_darwin_sendfile_chain.o \
	objs/src/os/unix/ngx_recv.o \
	objs/src/os/unix/ngx_send.o \
	objs/src/os/unix/ngx_udp_recv.o \
	objs/src/os/unix/ngx_udp_send.o \
	objs/src/os/unix/ngx_udp_sendmsg_chain.o \
	objs/src/os/unix/ngx_writev_chain.o \
	objs/src/os/unix/ngx_readv_chain.o \
	objs/src/os/unix/ngx_process_cycle.o \
	objs/src/event/ngx_event.o \
	objs/src/event/ngx_event_timer.o \

	$(LINK) -o objs/nginx \
	objs/src/core/nginx.o \
	objs/src/core/ngx_log.o \
	objs/src/core/ngx_array.o \
	objs/src/core/ngx_cycle.o \
	objs/src/core/ngx_buf.o \
	objs/src/core/ngx_string.o \
	objs/src/core/ngx_palloc.o \
	objs/src/core/ngx_times.o \
	objs/src/core/ngx_regex.o \
	objs/src/core/ngx_connection.o \
	objs/src/core/ngx_conf_file.o \
	objs/src/core/ngx_rbtree.o \
	objs/src/core/ngx_module.o \
	objs/src/core/ngx_file.o \
	objs/src/core/ngx_parse.o \
	objs/src/core/ngx_cpuinfo.o \
	objs/src/core/ngx_crc32.o \
	objs/src/core/ngx_slab.o \
	objs/src/core/ngx_inet.o \
	objs/src/core/ngx_modules.o \
	objs/src/os/unix/ngx_alloc.o \
	objs/src/os/unix/ngx_process.o \
	objs/src/os/unix/ngx_dlopen.o \
	objs/src/os/unix/ngx_errno.o \
	objs/src/os/unix/ngx_posix_init.o \
	objs/src/os/unix/ngx_time.o \
	objs/src/os/unix/ngx_setproctitle.o \
	objs/src/os/unix/ngx_darwin_init.o \
	objs/src/os/unix/ngx_darwin_sendfile_chain.o \
	objs/src/os/unix/ngx_recv.o \
	objs/src/os/unix/ngx_send.o \
	objs/src/os/unix/ngx_udp_recv.o \
	objs/src/os/unix/ngx_udp_send.o \
	objs/src/os/unix/ngx_udp_sendmsg_chain.o \
	objs/src/os/unix/ngx_writev_chain.o \
	objs/src/os/unix/ngx_readv_chain.o \
	objs/src/os/unix/ngx_process_cycle.o \
	objs/src/event/ngx_event.o \
	objs/src/event/ngx_event_timer.o \
	-L/usr/local/Cellar/openresty/1.21.4.2_1/luajit/lib -L/usr/local/Cellar/openresty/1.21.4.2_1/luajit/lib -Wl,-u,_pcre_version -L/usr/local/Cellar/openresty/1.21.4.2_1/luajit/lib -lluajit-5.1 -lm -ldl -pagezero_size 10000 -image_base 100000000 -L/usr/local/Cellar/pcre/8.45/lib -lpcre -L/usr/local/Cellar/openresty/1.21.4.2_1/luajit/lib -lluajit-5.1 -lm -ldl -pagezero_size 10000 -image_base 100000000 -lpcre2-8 -lssl -lcrypto -lz

# objs/src/event/ngx_event_openssl.o
# objs/src/event/ngx_event_pipe.o
test:
	
	
	objs/src/event/ngx_event_posted.o \
	objs/src/event/ngx_event_accept.o \
	objs/src/event/ngx_event_udp.o \
	objs/src/event/ngx_event_connect.o \
	objs/src/event/ngx_event_pipe.o \
	objs/src/event/ngx_event_udpv2.o \
	
	
	objs/src/os/unix/ngx_files.o \
	objs/src/os/unix/ngx_socket.o \
	
	objs/src/os/unix/ngx_send.o \
	
	objs/src/os/unix/ngx_channel.o \
	objs/src/os/unix/ngx_shmem.o \
	
	objs/src/os/unix/ngx_daemon.o \
	objs/src/os/unix/ngx_setaffinity.o \
	
	
	objs/src/os/unix/ngx_user.o \
	
	objs/src/os/unix/ngx_pipe.o \
	objs/src/os/unix/ngx_sysinfo.o \
	
	
	objs/src/event/modules/ngx_kqueue_module.o \
	
	objs/src/proc/ngx_proc.o \
	objs/src/event/ngx_event_openssl.o \
	objs/src/event/ngx_event_openssl_stapling.o
	
	objs/src/event/ngx_event_posted.o \
	objs/src/event/ngx_event_accept.o \
	objs/src/event/ngx_event_udp.o \
	objs/src/event/ngx_event_connect.o \
	objs/src/event/ngx_event_udpv2.o \
	
	objs/src/os/unix/ngx_files.o \
	objs/src/os/unix/ngx_socket.o \
	
	
	objs/src/os/unix/ngx_channel.o \
	objs/src/os/unix/ngx_shmem.o \
	
	objs/src/os/unix/ngx_daemon.o \
	objs/src/os/unix/ngx_setaffinity.o \
	objs/src/os/unix/ngx_user.o \
	
	objs/src/os/unix/ngx_pipe.o \
	objs/src/os/unix/ngx_sysinfo.o \
	
	objs/src/event/modules/ngx_kqueue_module.o \
	
	
	objs/src/proc/ngx_proc.o \
	objs/src/event/ngx_event_openssl_stapling.o \


objs/src/core/nginx.o:	$(CORE_DEPS) \
	src/core/nginx.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/nginx.o \
		src/core/nginx.c

objs/src/core/ngx_log.o:	$(CORE_DEPS) \
	src/core/ngx_log.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_log.o \
		src/core/ngx_log.c

objs/src/core/ngx_array.o:	$(CORE_DEPS) \
	src/core/ngx_array.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_array.o \
		src/core/ngx_array.c

objs/src/core/ngx_cycle.o:	$(CORE_DEPS) \
	src/core/ngx_cycle.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_cycle.o \
		src/core/ngx_cycle.c

objs/src/core/ngx_cpuinfo.o:	$(CORE_DEPS) \
	src/core/ngx_cpuinfo.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_cpuinfo.o \
		src/core/ngx_cpuinfo.c

objs/src/core/ngx_buf.o:	$(CORE_DEPS) \
	src/core/ngx_buf.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_buf.o \
		src/core/ngx_buf.c

objs/src/core/ngx_string.o:	$(CORE_DEPS) \
	src/core/ngx_string.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_string.o \
		src/core/ngx_string.c

objs/src/core/ngx_palloc.o:	$(CORE_DEPS) \
	src/core/ngx_palloc.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_palloc.o \
		src/core/ngx_palloc.c

objs/src/core/ngx_times.o:	$(CORE_DEPS) \
	src/core/ngx_times.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_times.o \
		src/core/ngx_times.c

objs/src/core/ngx_regex.o:	$(CORE_DEPS) \
	src/core/ngx_regex.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_regex.o \
		src/core/ngx_regex.c
		
objs/src/core/ngx_connection.o:	$(CORE_DEPS) \
	src/core/ngx_connection.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_connection.o \
		src/core/ngx_connection.c

objs/src/core/ngx_conf_file.o:	$(CORE_DEPS) \
	src/core/ngx_conf_file.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_conf_file.o \
		src/core/ngx_conf_file.c

objs/src/core/ngx_rbtree.o:	$(CORE_DEPS) \
	src/core/ngx_rbtree.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_rbtree.o \
		src/core/ngx_rbtree.c

objs/src/core/ngx_module.o:	$(CORE_DEPS) \
	src/core/ngx_module.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_module.o \
		src/core/ngx_module.c

objs/src/core/ngx_file.o:	$(CORE_DEPS) \
	src/core/ngx_file.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_file.o \
		src/core/ngx_file.c

objs/src/core/ngx_parse.o:	$(CORE_DEPS) \
	src/core/ngx_parse.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_parse.o \
		src/core/ngx_parse.c

objs/src/core/ngx_crc32.o:	$(CORE_DEPS) \
	src/core/ngx_crc32.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_crc32.o \
		src/core/ngx_crc32.c

objs/src/core/ngx_slab.o:	$(CORE_DEPS) \
	src/core/ngx_slab.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_slab.o \
		src/core/ngx_slab.c

objs/src/core/ngx_inet.o:	$(CORE_DEPS) \
	src/core/ngx_inet.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_inet.o \
		src/core/ngx_inet.c
		
objs/src/core/ngx_modules.o:	$(CORE_DEPS) \
	src/core/ngx_modules.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/core/ngx_modules.o \
		src/core/ngx_modules.c

################################os/event############################################
# 以下是每一个 c -> o 文件编译, 主要使用 nginx 原有的 event/os/proc c 文件，路径 
objs/src/event/ngx_event.o:	$(CORE_DEPS) \
	src/event/ngx_event.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event.o \
		src/event/ngx_event.c


objs/src/event/ngx_event_timer.o:	$(CORE_DEPS) \
	src/event/ngx_event_timer.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_timer.o \
		src/event/ngx_event_timer.c


objs/src/event/ngx_event_posted.o:	$(CORE_DEPS) \
	src/event/ngx_event_posted.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_posted.o \
		src/event/ngx_event_posted.c


objs/src/event/ngx_event_accept.o:	$(CORE_DEPS) \
	src/event/ngx_event_accept.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_accept.o \
		src/event/ngx_event_accept.c


objs/src/event/ngx_event_udp.o:	$(CORE_DEPS) \
	src/event/ngx_event_udp.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_udp.o \
		src/event/ngx_event_udp.c


objs/src/event/ngx_event_connect.o:	$(CORE_DEPS) \
	src/event/ngx_event_connect.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_connect.o \
		src/event/ngx_event_connect.c


objs/src/event/ngx_event_pipe.o:	$(CORE_DEPS) \
	src/event/ngx_event_pipe.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_pipe.o \
		src/event/ngx_event_pipe.c


objs/src/event/ngx_event_udpv2.o:	$(CORE_DEPS) \
	src/event/ngx_event_udpv2.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_udpv2.o \
		src/event/ngx_event_udpv2.c


objs/src/os/unix/ngx_time.o:	$(CORE_DEPS) \
	src/os/unix/ngx_time.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_time.o \
		src/os/unix/ngx_time.c


objs/src/os/unix/ngx_errno.o:	$(CORE_DEPS) \
	src/os/unix/ngx_errno.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_errno.o \
		src/os/unix/ngx_errno.c


objs/src/os/unix/ngx_alloc.o:	$(CORE_DEPS) \
	src/os/unix/ngx_alloc.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_alloc.o \
		src/os/unix/ngx_alloc.c


objs/src/os/unix/ngx_files.o:	$(CORE_DEPS) \
	src/os/unix/ngx_files.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_files.o \
		src/os/unix/ngx_files.c


objs/src/os/unix/ngx_socket.o:	$(CORE_DEPS) \
	src/os/unix/ngx_socket.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_socket.o \
		src/os/unix/ngx_socket.c


objs/src/os/unix/ngx_recv.o:	$(CORE_DEPS) \
	src/os/unix/ngx_recv.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_recv.o \
		src/os/unix/ngx_recv.c


objs/src/os/unix/ngx_readv_chain.o:	$(CORE_DEPS) \
	src/os/unix/ngx_readv_chain.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_readv_chain.o \
		src/os/unix/ngx_readv_chain.c


objs/src/os/unix/ngx_udp_recv.o:	$(CORE_DEPS) \
	src/os/unix/ngx_udp_recv.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_udp_recv.o \
		src/os/unix/ngx_udp_recv.c


objs/src/os/unix/ngx_send.o:	$(CORE_DEPS) \
	src/os/unix/ngx_send.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_send.o \
		src/os/unix/ngx_send.c


objs/src/os/unix/ngx_writev_chain.o:	$(CORE_DEPS) \
	src/os/unix/ngx_writev_chain.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_writev_chain.o \
		src/os/unix/ngx_writev_chain.c


objs/src/os/unix/ngx_udp_send.o:	$(CORE_DEPS) \
	src/os/unix/ngx_udp_send.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_udp_send.o \
		src/os/unix/ngx_udp_send.c


objs/src/os/unix/ngx_udp_sendmsg_chain.o:	$(CORE_DEPS) \
	src/os/unix/ngx_udp_sendmsg_chain.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_udp_sendmsg_chain.o \
		src/os/unix/ngx_udp_sendmsg_chain.c


objs/src/os/unix/ngx_channel.o:	$(CORE_DEPS) \
	src/os/unix/ngx_channel.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_channel.o \
		src/os/unix/ngx_channel.c


objs/src/os/unix/ngx_shmem.o:	$(CORE_DEPS) \
	src/os/unix/ngx_shmem.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_shmem.o \
		src/os/unix/ngx_shmem.c


objs/src/os/unix/ngx_process.o:	$(CORE_DEPS) \
	src/os/unix/ngx_process.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_process.o \
		src/os/unix/ngx_process.c


objs/src/os/unix/ngx_daemon.o:	$(CORE_DEPS) \
	src/os/unix/ngx_daemon.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_daemon.o \
		src/os/unix/ngx_daemon.c


objs/src/os/unix/ngx_setaffinity.o:	$(CORE_DEPS) \
	src/os/unix/ngx_setaffinity.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_setaffinity.o \
		src/os/unix/ngx_setaffinity.c


objs/src/os/unix/ngx_setproctitle.o:	$(CORE_DEPS) \
	src/os/unix/ngx_setproctitle.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_setproctitle.o \
		src/os/unix/ngx_setproctitle.c


objs/src/os/unix/ngx_posix_init.o:	$(CORE_DEPS) \
	src/os/unix/ngx_posix_init.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_posix_init.o \
		src/os/unix/ngx_posix_init.c


objs/src/os/unix/ngx_user.o:	$(CORE_DEPS) \
	src/os/unix/ngx_user.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_user.o \
		src/os/unix/ngx_user.c


objs/src/os/unix/ngx_dlopen.o:	$(CORE_DEPS) \
	src/os/unix/ngx_dlopen.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_dlopen.o \
		src/os/unix/ngx_dlopen.c


objs/src/os/unix/ngx_pipe.o:	$(CORE_DEPS) \
	src/os/unix/ngx_pipe.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_pipe.o \
		src/os/unix/ngx_pipe.c


objs/src/os/unix/ngx_sysinfo.o:	$(CORE_DEPS) \
	src/os/unix/ngx_sysinfo.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_sysinfo.o \
		src/os/unix/ngx_sysinfo.c


objs/src/os/unix/ngx_process_cycle.o:	$(CORE_DEPS) \
	src/os/unix/ngx_process_cycle.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_process_cycle.o \
		src/os/unix/ngx_process_cycle.c


objs/src/os/unix/ngx_darwin_init.o:	$(CORE_DEPS) \
	src/os/unix/ngx_darwin_init.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_darwin_init.o \
		src/os/unix/ngx_darwin_init.c


objs/src/event/modules/ngx_kqueue_module.o:	$(CORE_DEPS) \
	src/event/modules/ngx_kqueue_module.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/modules/ngx_kqueue_module.o \
		src/event/modules/ngx_kqueue_module.c


objs/src/os/unix/ngx_darwin_sendfile_chain.o:	$(CORE_DEPS) \
	src/os/unix/ngx_darwin_sendfile_chain.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/os/unix/ngx_darwin_sendfile_chain.o \
		src/os/unix/ngx_darwin_sendfile_chain.c


objs/src/proc/ngx_proc.o:	$(CORE_DEPS) \
	src/proc/ngx_proc.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/proc/ngx_proc.o \
		src/proc/ngx_proc.c


objs/src/event/ngx_event_openssl.o:	$(CORE_DEPS) \
	src/event/ngx_event_openssl.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_openssl.o \
		src/event/ngx_event_openssl.c


objs/src/event/ngx_event_openssl_stapling.o:	$(CORE_DEPS) \
	src/event/ngx_event_openssl_stapling.c
	$(CC) -c $(CFLAGS) $(CORE_INCS) \
		-o objs/src/event/ngx_event_openssl_stapling.o \
		src/event/ngx_event_openssl_stapling.c


################################################################
run:
	./objs/nginx -c ./nginx-tcp.conf -e stderr


