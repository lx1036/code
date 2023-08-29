#define NGX_CONFIGURE " --prefix=./bin --with-stream --with-stream_ssl_module --with-debug --with-pcre --add-module=./modules/ngx_http_echo_module --add-module=./modules/ngx_stream_lua_module_0.0.13 --add-module=./modules/ngx_http_upstream_dyups_module --with-http_lua_module --with-luajit-inc=/usr/local/Cellar/openresty/1.21.4.2_1/luajit/include/luajit-2.1 --with-luajit-lib=/usr/local/Cellar/openresty/1.21.4.2_1/luajit/lib --with-lua-inc=/usr/local/Cellar/lua@5.1/5.1.5_8/include/lua-5.1 --with-lua-lib=/usr/local/Cellar/lua@5.1/5.1.5_8/lib"

#ifndef NGX_DEBUG
#define NGX_DEBUG  1
#endif

#ifndef NGX_ERROR_LOG_STDERR
#define NGX_ERROR_LOG_STDERR  1
#endif

#ifndef NGX_COMPILER
#define NGX_COMPILER  "clang 14.0.3 (clang-1403.0.22.14.1)"
#endif


#ifndef NGX_HAVE_GCC_ATOMIC
#define NGX_HAVE_GCC_ATOMIC  1
#endif


#ifndef NGX_HAVE_C99_VARIADIC_MACROS
#define NGX_HAVE_C99_VARIADIC_MACROS  1
#endif


#ifndef NGX_HAVE_GCC_VARIADIC_MACROS
#define NGX_HAVE_GCC_VARIADIC_MACROS  1
#endif


#ifndef NGX_HAVE_GCC_BSWAP64
#define NGX_HAVE_GCC_BSWAP64  1
#endif


#ifndef NGX_HAVE_KQUEUE
#define NGX_HAVE_KQUEUE  1
#endif


#ifndef NGX_HAVE_CLEAR_EVENT
#define NGX_HAVE_CLEAR_EVENT  1
#endif


#ifndef NGX_HAVE_TIMER_EVENT
#define NGX_HAVE_TIMER_EVENT  1
#endif


#ifndef NGX_HAVE_SENDFILE
#define NGX_HAVE_SENDFILE  1
#endif


#ifndef NGX_DARWIN_ATOMIC
#define NGX_DARWIN_ATOMIC  1
#endif


#ifndef NGX_HAVE_NONALIGNED
#define NGX_HAVE_NONALIGNED  1
#endif


#ifndef NGX_CPU_CACHE_LINE
#define NGX_CPU_CACHE_LINE  64
#endif


#define NGX_KQUEUE_UDATA_T  (void *)


#ifndef NGX_HAVE_F_NOCACHE
#define NGX_HAVE_F_NOCACHE  1
#endif


#ifndef NGX_HAVE_STATFS
#define NGX_HAVE_STATFS  1
#endif


#ifndef NGX_HAVE_STATVFS
#define NGX_HAVE_STATVFS  1
#endif


#ifndef NGX_HAVE_DLOPEN
#define NGX_HAVE_DLOPEN  1
#endif


#ifndef NGX_HAVE_SCHED_YIELD
#define NGX_HAVE_SCHED_YIELD  1
#endif


#ifndef NGX_HAVE_REUSEPORT
#define NGX_HAVE_REUSEPORT  1
#endif


#ifndef NGX_HAVE_IP_RECVDSTADDR
#define NGX_HAVE_IP_RECVDSTADDR  1
#endif


#ifndef NGX_HAVE_IP_PKTINFO
#define NGX_HAVE_IP_PKTINFO  1
#endif


#ifndef NGX_HAVE_IPV6_RECVPKTINFO
#define NGX_HAVE_IPV6_RECVPKTINFO  1
#endif


#ifndef NGX_HAVE_TCP_FASTOPEN
#define NGX_HAVE_TCP_FASTOPEN  1
#endif


#ifndef NGX_HAVE_UNIX_DOMAIN
#define NGX_HAVE_UNIX_DOMAIN  1
#endif


#ifndef NGX_PTR_SIZE
#define NGX_PTR_SIZE  8
#endif


#ifndef NGX_SIG_ATOMIC_T_SIZE
#define NGX_SIG_ATOMIC_T_SIZE  4
#endif


#ifndef NGX_HAVE_LITTLE_ENDIAN
#define NGX_HAVE_LITTLE_ENDIAN  1
#endif


#ifndef NGX_MAX_SIZE_T_VALUE
#define NGX_MAX_SIZE_T_VALUE  9223372036854775807LL
#endif


#ifndef NGX_SIZE_T_LEN
#define NGX_SIZE_T_LEN  (sizeof("-9223372036854775808") - 1)
#endif


#ifndef NGX_MAX_OFF_T_VALUE
#define NGX_MAX_OFF_T_VALUE  9223372036854775807LL
#endif


#ifndef NGX_OFF_T_LEN
#define NGX_OFF_T_LEN  (sizeof("-9223372036854775808") - 1)
#endif


#ifndef NGX_TIME_T_SIZE
#define NGX_TIME_T_SIZE  8
#endif


#ifndef NGX_TIME_T_LEN
#define NGX_TIME_T_LEN  (sizeof("-9223372036854775808") - 1)
#endif


#ifndef NGX_MAX_TIME_T_VALUE
#define NGX_MAX_TIME_T_VALUE  9223372036854775807LL
#endif


#ifndef NGX_HAVE_INET6
#define NGX_HAVE_INET6  1
#endif


#ifndef NGX_HAVE_PREAD
#define NGX_HAVE_PREAD  1
#endif


#ifndef NGX_HAVE_PWRITE
#define NGX_HAVE_PWRITE  1
#endif


#ifndef NGX_HAVE_PWRITEV
#define NGX_HAVE_PWRITEV  1
#endif


#ifndef NGX_SYS_NERR
#define NGX_SYS_NERR  107
#endif


#ifndef NGX_HAVE_LOCALTIME_R
#define NGX_HAVE_LOCALTIME_R  1
#endif


#ifndef NGX_HAVE_CLOCK_MONOTONIC
#define NGX_HAVE_CLOCK_MONOTONIC  1
#endif


#ifndef NGX_HAVE_POSIX_MEMALIGN
#define NGX_HAVE_POSIX_MEMALIGN  1
#endif


#ifndef NGX_HAVE_MAP_ANON
#define NGX_HAVE_MAP_ANON  1
#endif


#ifndef NGX_HAVE_SYSVSHM
#define NGX_HAVE_SYSVSHM  1
#endif


#ifndef NGX_HAVE_MSGHDR_MSG_CONTROL
#define NGX_HAVE_MSGHDR_MSG_CONTROL  1
#endif


#ifndef NGX_HAVE_FIONBIO
#define NGX_HAVE_FIONBIO  1
#endif


#ifndef NGX_HAVE_FIONREAD
#define NGX_HAVE_FIONREAD  1
#endif


#ifndef NGX_HAVE_GMTOFF
#define NGX_HAVE_GMTOFF  1
#endif


#ifndef NGX_HAVE_D_NAMLEN
#define NGX_HAVE_D_NAMLEN  1
#endif


#ifndef NGX_HAVE_D_TYPE
#define NGX_HAVE_D_TYPE  1
#endif


#ifndef NGX_HAVE_SC_NPROCESSORS_ONLN
#define NGX_HAVE_SC_NPROCESSORS_ONLN  1
#endif


#ifndef NGX_HAVE_OPENAT
#define NGX_HAVE_OPENAT  1
#endif


#ifndef NGX_HAVE_GETADDRINFO
#define NGX_HAVE_GETADDRINFO  1
#endif


#ifndef NGX_HAVE_GETLOADAVG
#define NGX_HAVE_GETLOADAVG  1
#endif


#ifndef NGX_RESOLVER_FILE
#define NGX_RESOLVER_FILE  "/etc/resolv.conf"
#endif


#ifndef NGX_PROCS
#define NGX_PROCS  1
#endif


#ifndef NGX_HTTP_CACHE
#define NGX_HTTP_CACHE  1
#endif


#ifndef NGX_HTTP_GZIP
#define NGX_HTTP_GZIP  1
#endif


#ifndef NGX_HTTP_SSI
#define NGX_HTTP_SSI  1
#endif


#ifndef NGX_CRYPT
#define NGX_CRYPT  1
#endif


#ifndef NGX_HTTP_X_FORWARDED_FOR
#define NGX_HTTP_X_FORWARDED_FOR  1
#endif


#ifndef NGX_HTTP_SSL
#define NGX_HTTP_SSL  1
#endif


#ifndef NGX_HTTP_X_FORWARDED_FOR
#define NGX_HTTP_X_FORWARDED_FOR  1
#endif


#ifndef NGX_HTTP_UPSTREAM_ZONE
#define NGX_HTTP_UPSTREAM_ZONE  1
#endif


#ifndef NGX_STAT_STUB
#define NGX_STAT_STUB  1
#endif


#ifndef NGX_HTTP_UPSTREAM_RBTREE
#define NGX_HTTP_UPSTREAM_RBTREE  1
#endif


#ifndef NGX_STREAM_SSL
#define NGX_STREAM_SSL  1
#endif


#ifndef NGX_STREAM_UPSTREAM_ZONE
#define NGX_STREAM_UPSTREAM_ZONE  1
#endif


#ifndef NGX_STREAM_LUA_HAVE_SA_RESTART
#define NGX_STREAM_LUA_HAVE_SA_RESTART  1
#endif


#ifndef NGX_DYUPS
#define NGX_DYUPS  1
#endif


#ifndef NGX_HTTP_LUA_HAVE_SA_RESTART
#define NGX_HTTP_LUA_HAVE_SA_RESTART  1
#endif


#ifndef T_NGX_DNS_RESOLVE_BACKUP
#define T_NGX_DNS_RESOLVE_BACKUP  1
#endif


#ifndef T_NGX_MASTER_ENV
#define T_NGX_MASTER_ENV  1
#endif


#ifndef T_PIPES
#define T_PIPES  1
#endif


#ifndef T_NGX_INPUT_BODY_FILTER
#define T_NGX_INPUT_BODY_FILTER  1
#endif


#ifndef T_NGX_GZIP_CLEAR_ETAG
#define T_NGX_GZIP_CLEAR_ETAG  1
#endif


#ifndef T_NGX_RESOLVER_FILE
#define T_NGX_RESOLVER_FILE  1
#endif


#ifndef T_DEPRECATED
#define T_DEPRECATED  1
#endif


#ifndef T_NGX_VARS
#define T_NGX_VARS  1
#endif


#ifndef T_NGX_HTTP_STUB_STATUS
#define T_NGX_HTTP_STUB_STATUS  1
#endif


#ifndef T_UPSTREAM_TRIES
#define T_UPSTREAM_TRIES  1
#endif


#ifndef T_GEO
#define T_GEO  1
#endif


#ifndef T_NGX_RET_CACHE
#define T_NGX_RET_CACHE  1
#endif


#ifndef T_LIMIT_REQ
#define T_LIMIT_REQ  1
#endif


#ifndef T_LIMIT_REQ_RATE_VAR
#define T_LIMIT_REQ_RATE_VAR  1
#endif


#ifndef T_NGX_SSL_HANDSHAKE_TIME
#define T_NGX_SSL_HANDSHAKE_TIME  1
#endif


#ifndef T_NGX_HTTP_IMPROVED_IF
#define T_NGX_HTTP_IMPROVED_IF  1
#endif


#ifndef T_NGX_HTTP_UPSTREAM_RETRY_CC
#define T_NGX_HTTP_UPSTREAM_RETRY_CC  1
#endif


#ifndef T_NGX_HTTP_SSL_VCE
#define T_NGX_HTTP_SSL_VCE  1
#endif


#ifndef T_NGX_HTTP_UPSTREAM_RANDOM
#define T_NGX_HTTP_UPSTREAM_RANDOM  1
#endif


#ifndef T_NGX_IMPROVED_LIST
#define T_NGX_IMPROVED_LIST  1
#endif


#ifndef T_NGX_SERVER_INFO
#define T_NGX_SERVER_INFO  1
#endif


#ifndef T_NGX_ACCEPT_FILTER
#define T_NGX_ACCEPT_FILTER  1
#endif


#ifndef T_NGX_MODIFY_DEFAULT_VALUE
#define T_NGX_MODIFY_DEFAULT_VALUE  1
#endif


#ifndef T_NGX_HTTP_UPSTREAM_ID
#define T_NGX_HTTP_UPSTREAM_ID  1
#endif


#ifndef T_NGX_HTTP_IMPROVED_REWRITE
#define T_NGX_HTTP_IMPROVED_REWRITE  1
#endif


#ifndef T_NGX_SHOW_INFO
#define T_NGX_SHOW_INFO  1
#endif


#ifndef T_NGX_HTTP_IMAGE_FILTER
#define T_NGX_HTTP_IMAGE_FILTER  1
#endif


#ifndef T_HTTP_HEADER
#define T_HTTP_HEADER  1
#endif


#ifndef T_HTTP_UPSTREAM_TIMEOUT_VAR
#define T_HTTP_UPSTREAM_TIMEOUT_VAR  1
#endif


#ifndef NGX_PCRE2
#define NGX_PCRE2  1
#endif


#ifndef NGX_PCRE
#define NGX_PCRE  1
#endif


#ifndef NGX_OPENSSL
#define NGX_OPENSSL  1
#endif


#ifndef NGX_SSL
#define NGX_SSL  1
#endif


#ifndef T_NGX_HAVE_DTLS
#define T_NGX_HAVE_DTLS  1
#endif


#ifndef NGX_ZLIB
#define NGX_ZLIB  1
#endif


#ifndef NGX_PREFIX
#define NGX_PREFIX  "./bin/"
#endif


#ifndef NGX_CONF_PREFIX
#define NGX_CONF_PREFIX  "conf/"
#endif


#ifndef NGX_SBIN_PATH
#define NGX_SBIN_PATH  "sbin/nginx"
#endif


#ifndef NGX_CONF_PATH
#define NGX_CONF_PATH  "conf/nginx.conf"
#endif


#ifndef NGX_PID_PATH
#define NGX_PID_PATH  "logs/nginx.pid"
#endif


#ifndef NGX_LOCK_PATH
#define NGX_LOCK_PATH  "logs/nginx.lock"
#endif


#ifndef NGX_ERROR_LOG_PATH
#define NGX_ERROR_LOG_PATH  "logs/error.log"
#endif


#ifndef NGX_HTTP_LOG_PATH
#define NGX_HTTP_LOG_PATH  "logs/access.log"
#endif


#ifndef NGX_HTTP_CLIENT_TEMP_PATH
#define NGX_HTTP_CLIENT_TEMP_PATH  "client_body_temp"
#endif


#ifndef NGX_HTTP_PROXY_TEMP_PATH
#define NGX_HTTP_PROXY_TEMP_PATH  "proxy_temp"
#endif


#ifndef NGX_HTTP_FASTCGI_TEMP_PATH
#define NGX_HTTP_FASTCGI_TEMP_PATH  "fastcgi_temp"
#endif


#ifndef NGX_HTTP_UWSGI_TEMP_PATH
#define NGX_HTTP_UWSGI_TEMP_PATH  "uwsgi_temp"
#endif


#ifndef NGX_HTTP_SCGI_TEMP_PATH
#define NGX_HTTP_SCGI_TEMP_PATH  "scgi_temp"
#endif


#ifndef NGX_SUPPRESS_WARN
#define NGX_SUPPRESS_WARN  1
#endif


#ifndef NGX_SMP
#define NGX_SMP  1
#endif


#ifndef NGX_USER
#define NGX_USER  "nobody"
#endif


#ifndef NGX_GROUP
#define NGX_GROUP  "nobody"
#endif

