


/*
log_format main '$remote_addr [$time_local] $bytes_received $bytes_sent $protocol $status $connection $session_time';
*/


#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_stream.h>

#if (NGX_ZLIB)
#include <zlib.h>
#endif


typedef struct ngx_stream_log_op_s  ngx_stream_log_op_t;
typedef u_char *(*ngx_stream_log_op_run_pt) (ngx_stream_session_t *s, u_char *buf, ngx_stream_log_op_t *op);
typedef size_t (*ngx_stream_log_op_getlen_pt) (ngx_stream_session_t *s, uintptr_t data);

struct ngx_stream_log_op_s {
    size_t                       len;
    ngx_stream_log_op_getlen_pt  getlen;
    ngx_stream_log_op_run_pt     run;
    uintptr_t                    data;
};


typedef struct {
    ngx_str_t                    name;
    ngx_array_t                 *flushes;
    ngx_array_t                 *ops;        /* array of ngx_stream_log_op_t */
} ngx_stream_log_fmt_t;


typedef struct {
    ngx_array_t                  formats;    /* array of ngx_stream_log_fmt_t */
} ngx_stream_log_main_conf_t;


typedef struct {
    u_char                      *start;
    u_char                      *pos;
    u_char                      *last;

    ngx_event_t                 *event;
    ngx_msec_t                   flush;
    ngx_int_t                    gzip;
} ngx_stream_log_buf_t;


typedef struct {
    ngx_array_t                 *lengths;
    ngx_array_t                 *values;
} ngx_stream_log_script_t;


typedef struct {
    ngx_open_file_t             *file;
    ngx_stream_log_script_t     *script;
    time_t                       disk_full_time;
    time_t                       error_log_time;
//    ngx_syslog_peer_t           *syslog_peer;
    ngx_stream_log_fmt_t        *format;
    ngx_stream_complex_value_t  *filter;
} ngx_stream_log_t;


typedef struct {
    ngx_array_t                 *logs;       /* array of ngx_stream_log_t */

//    ngx_open_file_cache_t       *open_file_cache;
    time_t                       open_file_cache_valid;
    ngx_uint_t                   open_file_cache_min_uses;

    ngx_uint_t                   off;        /* unsigned  off:1 */
} ngx_stream_log_srv_conf_t;


typedef struct {
    ngx_str_t                    name;
    size_t                       len;
    ngx_stream_log_op_run_pt     run;
} ngx_stream_log_var_t;


#define NGX_STREAM_LOG_ESCAPE_DEFAULT  0
#define NGX_STREAM_LOG_ESCAPE_JSON     1
#define NGX_STREAM_LOG_ESCAPE_NONE     2


static void ngx_stream_log_write(ngx_stream_session_t *s, ngx_stream_log_t *log,
                                 u_char *buf, size_t len);
static ssize_t ngx_stream_log_script_write(ngx_stream_session_t *s,
                                           ngx_stream_log_script_t *script, u_char **name, u_char *buf, size_t len);

#if (NGX_ZLIB)
static ssize_t ngx_stream_log_gzip(ngx_fd_t fd, u_char *buf, size_t len,
                                   ngx_int_t level, ngx_log_t *log);

static void *ngx_stream_log_gzip_alloc(void *opaque, u_int items, u_int size);
static void ngx_stream_log_gzip_free(void *opaque, void *address);
#endif

static void ngx_stream_log_flush(ngx_open_file_t *file, ngx_log_t *log);
static void ngx_stream_log_flush_handler(ngx_event_t *ev);

static ngx_int_t ngx_stream_log_variable_compile(ngx_conf_t *cf,
                                                 ngx_stream_log_op_t *op, ngx_str_t *value, ngx_uint_t escape);
static size_t ngx_stream_log_variable_getlen(ngx_stream_session_t *s,
                                             uintptr_t data);
static u_char *ngx_stream_log_variable(ngx_stream_session_t *s, u_char *buf,
                                       ngx_stream_log_op_t *op);
static uintptr_t ngx_stream_log_escape(u_char *dst, u_char *src, size_t size);
static size_t ngx_stream_log_json_variable_getlen(ngx_stream_session_t *s,
                                                  uintptr_t data);
static u_char *ngx_stream_log_json_variable(ngx_stream_session_t *s,
                                            u_char *buf, ngx_stream_log_op_t *op);
static size_t ngx_stream_log_unescaped_variable_getlen(ngx_stream_session_t *s,
                                                       uintptr_t data);
static u_char *ngx_stream_log_unescaped_variable(ngx_stream_session_t *s,
                                                 u_char *buf, ngx_stream_log_op_t *op);

static void *ngx_stream_log_create_main_conf(ngx_conf_t *cf);
static void *ngx_stream_log_create_srv_conf(ngx_conf_t *cf);
static char *ngx_stream_log_merge_srv_conf(ngx_conf_t *cf, void *parent,
                                           void *child);
static char *ngx_stream_log_set_log(ngx_conf_t *cf, ngx_command_t *cmd,
                                    void *conf);
static char *ngx_stream_log_set_format(ngx_conf_t *cf, ngx_command_t *cmd,
                                       void *conf);
static char *ngx_stream_log_compile_format(ngx_conf_t *cf,
                                           ngx_array_t *flushes, ngx_array_t *ops, ngx_array_t *args, ngx_uint_t s);
static char *ngx_stream_log_open_file_cache(ngx_conf_t *cf, ngx_command_t *cmd,
                                            void *conf);
static ngx_int_t ngx_stream_log_init(ngx_conf_t *cf);


static ngx_command_t  ngx_stream_log_commands[] = {
        { ngx_string("log_format"),
          NGX_STREAM_MAIN_CONF|NGX_CONF_2MORE,
          ngx_stream_log_set_format,
          NGX_STREAM_MAIN_CONF_OFFSET,
          0,
          NULL },

        { ngx_string("access_log"),
          NGX_STREAM_MAIN_CONF|NGX_STREAM_SRV_CONF|NGX_CONF_1MORE,
          ngx_stream_log_set_log,
          NGX_STREAM_SRV_CONF_OFFSET,
          0,
          NULL },

//        { ngx_string("open_log_file_cache"),
//          NGX_STREAM_MAIN_CONF|NGX_STREAM_SRV_CONF|NGX_CONF_TAKE1234,
//          ngx_stream_log_open_file_cache,
//          NGX_STREAM_SRV_CONF_OFFSET,
//          0,
//          NULL },

        ngx_null_command
};
static ngx_stream_module_t  ngx_stream_log_module_ctx = {
        NULL,                                  /* preconfiguration */
        ngx_stream_log_init,                   /* postconfiguration */

        ngx_stream_log_create_main_conf,       /* create main configuration */
        NULL,                                  /* init main configuration */

        ngx_stream_log_create_srv_conf,        /* create server configuration */
        ngx_stream_log_merge_srv_conf          /* merge server configuration */
};
ngx_module_t  ngx_stream_log_module = {
        NGX_MODULE_V1,
        &ngx_stream_log_module_ctx,            /* module context */
        ngx_stream_log_commands,               /* module directives */
        NGX_STREAM_MODULE,                     /* module type */
        NULL,                                  /* init master */
        NULL,                                  /* init module */
        NULL,                                  /* init process */
        NULL,                                  /* init thread */
        NULL,                                  /* exit thread */
        NULL,                                  /* exit process */
        NULL,                                  /* exit master */
        NGX_MODULE_V1_PADDING
};
static ngx_int_t ngx_stream_log_handler(ngx_stream_session_t *s) {
    u_char                     *line, *p;
    size_t                      len, size;
    ssize_t                     n;
    ngx_str_t                   val;
    ngx_uint_t                  i, l;
    ngx_stream_log_t           *log;
    ngx_stream_log_op_t        *op;
    ngx_stream_log_buf_t       *buffer;
    ngx_stream_log_srv_conf_t  *lscf;

    ngx_log_error(NGX_LOG_STDERR, s->connection->log, 0, "stream log handler");
    lscf = ngx_stream_get_module_srv_conf(s, ngx_stream_log_module);
    if (lscf->off || lscf->logs == NULL) {
        return NGX_OK;
    }

    log = lscf->logs->elts;
    for (l = 0; l < lscf->logs->nelts; l++) {
        if (log[l].filter) {
            if (ngx_stream_complex_value(s, log[l].filter, &val) != NGX_OK) {
                return NGX_ERROR;
            }

            if (val.len == 0 || (val.len == 1 && val.data[0] == '0')) {
                continue;
            }
        }

        if (ngx_time() == log[l].disk_full_time) {
            /*
             * on FreeBSD writing to a full filesystem with enabled softupdates
             * may block process for much longer time than writing to non-full
             * filesystem, so we skip writing to a log for one second
             */
            continue;
        }

        ngx_stream_script_flush_no_cacheable_variables(s, log[l].format->flushes);
        len = 0;
        op = log[l].format->ops->elts;
        for (i = 0; i < log[l].format->ops->nelts; i++) {
            if (op[i].len == 0) {
                len += op[i].getlen(s, op[i].data);
            } else {
                len += op[i].len;
            }
        }

        len += NGX_LINEFEED_SIZE;
        buffer = log[l].file ? log[l].file->data : NULL;
        if (buffer) {
            if (len > (size_t) (buffer->last - buffer->pos)) {
                ngx_stream_log_write(s, &log[l], buffer->start, buffer->pos - buffer->start);
                buffer->pos = buffer->start;
            }

            if (len <= (size_t) (buffer->last - buffer->pos)) {
                p = buffer->pos;
                if (buffer->event && p == buffer->start) {
                    ngx_add_timer(buffer->event, buffer->flush);
                }

                for (i = 0; i < log[l].format->ops->nelts; i++) {
                    p = op[i].run(s, p, &op[i]);
                }

                ngx_linefeed(p);
                buffer->pos = p;

                continue;
            }

            if (buffer->event && buffer->event->timer_set) {
                ngx_del_timer(buffer->event);
            }
        }

        line = ngx_pnalloc(s->connection->pool, len);
        if (line == NULL) {
            return NGX_ERROR;
        }

        p = line;
        for (i = 0; i < log[l].format->ops->nelts; i++) {
            p = op[i].run(s, p, &op[i]);
        }
        ngx_linefeed(p);
        ngx_stream_log_write(s, &log[l], line, p - line);
    }

    return NGX_OK;
}
static ngx_int_t ngx_stream_log_init(ngx_conf_t *cf) {
    ngx_stream_handler_pt        *h;
    ngx_stream_core_main_conf_t  *cmcf;

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);

    h = ngx_array_push(&cmcf->phases[NGX_STREAM_LOG_PHASE].handlers);
    if (h == NULL) {
        return NGX_ERROR;
    }

    *h = ngx_stream_log_handler;

    return NGX_OK;
}

static void ngx_stream_log_write(ngx_stream_session_t *s, ngx_stream_log_t *log, u_char *buf, size_t len) {
    u_char                *name;
    time_t                 now;
    ssize_t                n;
    ngx_err_t              err;

    if (log->script == NULL) {
        name = log->file->name.data; // /dev/stdout /dev/stderr
        n = ngx_write_fd(log->file->fd, buf, len); // log->file->fd="/dev/stdout"
    } else {
        name = NULL;
        n = ngx_stream_log_script_write(s, log->script, &name, buf, len);
    }

    if (n == (ssize_t) len) {
        return;
    }

    now = ngx_time();

#if (T_PIPES)
    if (name == NULL) {
        name = (u_char *) "log file";
    }
#endif

    if (n == -1) {
        err = ngx_errno;
        if (err == NGX_ENOSPC) {
            log->disk_full_time = now;
        }

        if (now - log->error_log_time > 59) {
            ngx_log_error(NGX_LOG_ALERT, s->connection->log, err, ngx_write_fd_n " to \"%s\" failed", name);
            log->error_log_time = now;
        }

        return;
    }

    if (now - log->error_log_time > 59) {
        ngx_log_error(NGX_LOG_ALERT, s->connection->log, 0, ngx_write_fd_n " to \"%s\" was incomplete: %z of %uz", name, n, len);
        log->error_log_time = now;
    }
}
static ssize_t
ngx_stream_log_script_write(ngx_stream_session_t *s,
                            ngx_stream_log_script_t *script, u_char **name, u_char *buf, size_t len)
{
    ssize_t                     n = 0;
    ngx_str_t                   log;
//    ngx_open_file_info_t        of;
    ngx_stream_log_srv_conf_t  *lscf;

//    if (ngx_stream_script_run(s, &log, script->lengths->elts, 1, script->values->elts) == NULL) {
//        /* simulate successful logging */
//        return len;
//    }

//    log.data[log.len - 1] = '\0';
//    *name = log.data;
//
//    ngx_log_debug1(NGX_LOG_DEBUG_STREAM, s->connection->log, 0,
//                   "stream log \"%s\"", log.data);
//
//    lscf = ngx_stream_get_module_srv_conf(s, ngx_stream_log_module);
//
//    ngx_memzero(&of, sizeof(ngx_open_file_info_t));
//
//    of.log = 1;
//    of.valid = lscf->open_file_cache_valid;
//    of.min_uses = lscf->open_file_cache_min_uses;
//    of.directio = NGX_OPEN_FILE_DIRECTIO_OFF;
//
//    if (ngx_open_cached_file(lscf->open_file_cache, &log, &of,
//                             s->connection->pool)
//        != NGX_OK)
//    {
//        if (of.err == 0) {
//            /* simulate successful logging */
//            return len;
//        }
//
//        ngx_log_error(NGX_LOG_CRIT, s->connection->log, ngx_errno,
//                      "%s \"%s\" failed", of.failed, log.data);
//        /* simulate successful logging */
//        return len;
//    }
//
//    ngx_log_debug1(NGX_LOG_DEBUG_STREAM, s->connection->log, 0,
//                   "stream log #%d", of.fd);
//
//    n = ngx_write_fd(of.fd, buf, len);

    return n;
}

static void * ngx_stream_log_create_main_conf(ngx_conf_t *cf) {
    ngx_stream_log_main_conf_t  *conf;
    conf = ngx_pcalloc(cf->pool, sizeof(ngx_stream_log_main_conf_t));
    if (conf == NULL) {
        return NULL;
    }

    if (ngx_array_init(&conf->formats, cf->pool, 4, sizeof(ngx_stream_log_fmt_t)) != NGX_OK) {
        return NULL;
    }

    return conf;
}
static void * ngx_stream_log_create_srv_conf(ngx_conf_t *cf) {
    ngx_stream_log_srv_conf_t  *conf;
    conf = ngx_pcalloc(cf->pool, sizeof(ngx_stream_log_srv_conf_t));
    if (conf == NULL) {
        return NULL;
    }

    return conf;
}
static char * ngx_stream_log_merge_srv_conf(ngx_conf_t *cf, void *parent, void *child) {
    ngx_stream_log_srv_conf_t *prev = parent;
    ngx_stream_log_srv_conf_t *conf = child;
    if (conf->logs || conf->off) {
        return NGX_CONF_OK;
    }

    conf->logs = prev->logs;
    conf->off = prev->off;

    return NGX_CONF_OK;
}

static char * ngx_stream_log_set_format(ngx_conf_t *cf, ngx_command_t *cmd, void *conf) {
    ngx_stream_log_main_conf_t *lmcf = conf;
    ngx_str_t             *value;
    ngx_uint_t             i;
    ngx_stream_log_fmt_t  *fmt;

    value = cf->args->elts;
    fmt = lmcf->formats.elts;
    for (i = 0; i < lmcf->formats.nelts; i++) {
        if (fmt[i].name.len == value[1].len && ngx_strcmp(fmt[i].name.data, value[1].data) == 0) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "duplicate \"log_format\" name \"%V\"", &value[1]);
            return NGX_CONF_ERROR;
        }
    }

    fmt = ngx_array_push(&lmcf->formats);
    if (fmt == NULL) {
        return NGX_CONF_ERROR;
    }

    fmt->name = value[1]; // "proxy"
    fmt->flushes = ngx_array_create(cf->pool, 4, sizeof(ngx_int_t));
    if (fmt->flushes == NULL) {
        return NGX_CONF_ERROR;
    }

    fmt->ops = ngx_array_create(cf->pool, 16, sizeof(ngx_stream_log_op_t));
    if (fmt->ops == NULL) {
        return NGX_CONF_ERROR;
    }

    return ngx_stream_log_compile_format(cf, fmt->flushes, fmt->ops, cf->args, 2);
}
static u_char * ngx_stream_log_copy_short(ngx_stream_session_t *s, u_char *buf, ngx_stream_log_op_t *op) {
    size_t     len;
    uintptr_t  data;
    len = op->len;
    data = op->data;
    while (len--) {
        *buf++ = (u_char) (data & 0xff);
        data >>= 8;
    }

    return buf;
}
static u_char * ngx_stream_log_copy_long(ngx_stream_session_t *s, u_char *buf, ngx_stream_log_op_t *op) {
    return ngx_cpymem(buf, (u_char *) op->data, op->len);
}
static char * ngx_stream_log_compile_format(ngx_conf_t *cf, ngx_array_t *flushes, ngx_array_t *ops, ngx_array_t *args, ngx_uint_t s) {
    u_char                *data, *p, ch;
    size_t                 i, len;
    ngx_str_t             *value, var;
    ngx_int_t             *flush;
    ngx_uint_t             bracket, escape;
    ngx_stream_log_op_t   *op;
    escape = NGX_STREAM_LOG_ESCAPE_DEFAULT;
    value = args->elts;
    for ( /* void */ ; s < args->nelts; s++) {
        i = 0;
        while (i < value[s].len) {
            op = ngx_array_push(ops);
            if (op == NULL) {
                return NGX_CONF_ERROR;
            }

            data = &value[s].data[i];
            if (value[s].data[i] == '$') {
                if (++i == value[s].len) {
                    goto invalid;
                }

                if (value[s].data[i] == '{') {
                    bracket = 1;
                    if (++i == value[s].len) {
                        goto invalid;
                    }

                    var.data = &value[s].data[i];
                } else {
                    bracket = 0;
                    var.data = &value[s].data[i];
                }

                for (var.len = 0; i < value[s].len; i++, var.len++) {
                    ch = value[s].data[i];
                    if (ch == '}' && bracket) {
                        i++;
                        bracket = 0;
                        break;
                    }
                    if ((ch >= 'A' && ch <= 'Z')
                        || (ch >= 'a' && ch <= 'z')
                        || (ch >= '0' && ch <= '9')
                        || ch == '_') {
                        continue;
                    }

                    break;
                }

                if (bracket) {
                    ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "the closing bracket in \"%V\" variable is missing", &var);
                    return NGX_CONF_ERROR;
                }

                if (var.len == 0) {
                    goto invalid;
                }

                if (ngx_stream_log_variable_compile(cf, op, &var, escape) != NGX_OK) {
                    return NGX_CONF_ERROR;
                }

                if (flushes) {
                    flush = ngx_array_push(flushes);
                    if (flush == NULL) {
                        return NGX_CONF_ERROR;
                    }

                    *flush = op->data; /* variable index */
                }

                continue;
            }

            i++;
            while (i < value[s].len && value[s].data[i] != '$') {
                i++;
            }

            len = &value[s].data[i] - data;
            if (len) {
                op->len = len;
                op->getlen = NULL;
                if (len <= sizeof(uintptr_t)) {
                    op->run = ngx_stream_log_copy_short;
                    op->data = 0;
                    while (len--) {
                        op->data <<= 8;
                        op->data |= data[len];
                    }
                } else {
                    op->run = ngx_stream_log_copy_long;
                    p = ngx_pnalloc(cf->pool, len);
                    if (p == NULL) {
                        return NGX_CONF_ERROR;
                    }

                    ngx_memcpy(p, data, len);
                    op->data = (uintptr_t) p;
                }
            }
        }
    }

    return NGX_CONF_OK;

invalid:
    ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "invalid parameter \"%s\"", data);

    return NGX_CONF_ERROR;
}
static ngx_int_t
ngx_stream_log_variable_compile(ngx_conf_t *cf, ngx_stream_log_op_t *op, ngx_str_t *value, ngx_uint_t escape) {
    ngx_int_t  index;

    index = ngx_stream_get_variable_index(cf, value);
    if (index == NGX_ERROR) {
        return NGX_ERROR;
    }

    op->len = 0;
    switch (escape) {
        case NGX_STREAM_LOG_ESCAPE_JSON:
            op->getlen = ngx_stream_log_json_variable_getlen;
            op->run = ngx_stream_log_json_variable;
            break;

        case NGX_STREAM_LOG_ESCAPE_NONE:
            op->getlen = ngx_stream_log_unescaped_variable_getlen;
            op->run = ngx_stream_log_unescaped_variable;
            break;

        default: /* NGX_STREAM_LOG_ESCAPE_DEFAULT */
            op->getlen = ngx_stream_log_variable_getlen;
            op->run = ngx_stream_log_variable;
    }

    op->data = index;

    return NGX_OK;
}
static size_t ngx_stream_log_variable_getlen(ngx_stream_session_t *s, uintptr_t data) {
    uintptr_t                     len;
    ngx_stream_variable_value_t  *value;
    value = ngx_stream_get_indexed_variable(s, data);
    if (value == NULL || value->not_found) {
        return 1;
    }

    len = ngx_stream_log_escape(NULL, value->data, value->len);
    value->escape = len ? 1 : 0;
    return value->len + len * 3;
}
static u_char * ngx_stream_log_variable(ngx_stream_session_t *s, u_char *buf, ngx_stream_log_op_t *op) {
    ngx_stream_variable_value_t  *value;
    value = ngx_stream_get_indexed_variable(s, op->data);
    if (value == NULL || value->not_found) {
        *buf = '-';
        return buf + 1;
    }

    if (value->escape == 0) {
        return ngx_cpymem(buf, value->data, value->len);
    } else {
        return (u_char *) ngx_stream_log_escape(buf, value->data, value->len);
    }
}
static uintptr_t ngx_stream_log_escape(u_char *dst, u_char *src, size_t size) {
    ngx_uint_t      n;
    static u_char   hex[] = "0123456789ABCDEF";

    static uint32_t   escape[] = {
            0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */

            /* ?>=< ;:98 7654 3210  /.-, +*)( '&%$ #"!  */
            0x00000004, /* 0000 0000 0000 0000  0000 0000 0000 0100 */

            /* _^]\ [ZYX WVUT SRQP  ONML KJIH GFED CBA@ */
            0x10000000, /* 0001 0000 0000 0000  0000 0000 0000 0000 */

            /*  ~}| {zyx wvut srqp  onml kjih gfed cba` */
            0x80000000, /* 1000 0000 0000 0000  0000 0000 0000 0000 */

            0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */
            0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */
            0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */
            0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */
    };

    if (dst == NULL) {
        /* find the number of the characters to be escaped */
        n = 0;
        while (size) {
            if (escape[*src >> 5] & (1U << (*src & 0x1f))) {
                n++;
            }
            src++;
            size--;
        }

        return (uintptr_t) n;
    }

    while (size) {
        if (escape[*src >> 5] & (1U << (*src & 0x1f))) {
            *dst++ = '\\';
            *dst++ = 'x';
            *dst++ = hex[*src >> 4];
            *dst++ = hex[*src & 0xf];
            src++;
        } else {
            *dst++ = *src++;
        }
        size--;
    }

    return (uintptr_t) dst;
}

static char * ngx_stream_log_set_log(ngx_conf_t *cf, ngx_command_t *cmd, void *conf) {
    ngx_stream_log_srv_conf_t *lscf = conf;
    ssize_t                              size;
    ngx_int_t                            gzip;
    ngx_uint_t                           i, n;
    ngx_msec_t                           flush;
    ngx_str_t                           *value, name, s;
    ngx_stream_log_t                    *log;
    ngx_stream_log_buf_t                *buffer;
    ngx_stream_log_fmt_t                *fmt;
    ngx_stream_script_compile_t          sc;
    ngx_stream_log_main_conf_t          *lmcf;
    ngx_stream_compile_complex_value_t   ccv;

    value = cf->args->elts;
    if (ngx_strcmp(value[1].data, "off") == 0) { // "access_log off;"
        lscf->off = 1;
        if (cf->args->nelts == 2) {
            return NGX_CONF_OK;
        }

        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "invalid parameter \"%V\"", &value[2]);
        return NGX_CONF_ERROR;
    }

    if (lscf->logs == NULL) {
        lscf->logs = ngx_array_create(cf->pool, 2, sizeof(ngx_stream_log_t));
        if (lscf->logs == NULL) {
            return NGX_CONF_ERROR;
        }
    }

    lmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_log_module);
    log = ngx_array_push(lscf->logs);
    if (log == NULL) {
        return NGX_CONF_ERROR;
    }

    ngx_memzero(log, sizeof(ngx_stream_log_t));
    n = ngx_stream_script_variables_count(&value[1]);
    if (n == 0) {
        log->file = ngx_conf_open_file(cf->cycle, &value[1]);
        if (log->file == NULL) {
            return NGX_CONF_ERROR;
        }
    } else {
        if (ngx_conf_full_name(cf->cycle, &value[1], 0) != NGX_OK) {
            return NGX_CONF_ERROR;
        }

        log->script = ngx_pcalloc(cf->pool, sizeof(ngx_stream_log_script_t));
        if (log->script == NULL) {
            return NGX_CONF_ERROR;
        }

        ngx_memzero(&sc, sizeof(ngx_stream_script_compile_t));
        sc.cf = cf;
        sc.source = &value[1];
        sc.lengths = &log->script->lengths;
        sc.values = &log->script->values;
        sc.variables = n;
        sc.complete_lengths = 1;
        sc.complete_values = 1;
        if (ngx_stream_script_compile(&sc) != NGX_OK) {
            return NGX_CONF_ERROR;
        }
    }

    if (cf->args->nelts >= 3) {
        name = value[2];
    } else {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "log format is not specified");
        return NGX_CONF_ERROR;
    }

    fmt = lmcf->formats.elts;
    for (i = 0; i < lmcf->formats.nelts; i++) {
        if (fmt[i].name.len == name.len && ngx_strcasecmp(fmt[i].name.data, name.data) == 0) {
            log->format = &fmt[i];
            break;
        }
    }
    if (log->format == NULL) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "unknown log format \"%V\"", &name);
        return NGX_CONF_ERROR;
    }

    size = 0;
    flush = 0;
    gzip = 0;

    for (i = 3; i < cf->args->nelts; i++) {
        if (ngx_strncmp(value[i].data, "buffer=", 7) == 0) {
            s.len = value[i].len - 7;
            s.data = value[i].data + 7;

            size = ngx_parse_size(&s);

            if (size == NGX_ERROR || size == 0) {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                                   "invalid buffer size \"%V\"", &s);
                return NGX_CONF_ERROR;
            }

            continue;
        }

        if (ngx_strncmp(value[i].data, "flush=", 6) == 0) {
            s.len = value[i].len - 6;
            s.data = value[i].data + 6;

            flush = ngx_parse_time(&s, 0);

            if (flush == (ngx_msec_t) NGX_ERROR || flush == 0) {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                                   "invalid flush time \"%V\"", &s);
                return NGX_CONF_ERROR;
            }

            continue;
        }

        if (ngx_strncmp(value[i].data, "if=", 3) == 0) {
            s.len = value[i].len - 3;
            s.data = value[i].data + 3;

            ngx_memzero(&ccv, sizeof(ngx_stream_compile_complex_value_t));

            ccv.cf = cf;
            ccv.value = &s;
            ccv.complex_value = ngx_palloc(cf->pool,
                                           sizeof(ngx_stream_complex_value_t));
            if (ccv.complex_value == NULL) {
                return NGX_CONF_ERROR;
            }

            if (ngx_stream_compile_complex_value(&ccv) != NGX_OK) {
                return NGX_CONF_ERROR;
            }

            log->filter = ccv.complex_value;

            continue;
        }

        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                           "invalid parameter \"%V\"", &value[i]);
        return NGX_CONF_ERROR;
    }

    if (flush && size == 0) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "no buffer is defined for access_log \"%V\"", &value[1]);
        return NGX_CONF_ERROR;
    }

    if (size) {
        if (log->script) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                               "buffered logs cannot have variables in name");
            return NGX_CONF_ERROR;
        }

        if (log->file->data) {
            buffer = log->file->data;

            if (buffer->last - buffer->start != size
                || buffer->flush != flush
                || buffer->gzip != gzip)
            {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                                   "access_log \"%V\" already defined "
                                   "with conflicting parameters",
                                   &value[1]);
                return NGX_CONF_ERROR;
            }

            return NGX_CONF_OK;
        }

        buffer = ngx_pcalloc(cf->pool, sizeof(ngx_stream_log_buf_t));
        if (buffer == NULL) {
            return NGX_CONF_ERROR;
        }

        buffer->start = ngx_pnalloc(cf->pool, size);
        if (buffer->start == NULL) {
            return NGX_CONF_ERROR;
        }

        buffer->pos = buffer->start;
        buffer->last = buffer->start + size;

        if (flush) {
            buffer->event = ngx_pcalloc(cf->pool, sizeof(ngx_event_t));
            if (buffer->event == NULL) {
                return NGX_CONF_ERROR;
            }

            buffer->event->data = log->file;
            buffer->event->handler = ngx_stream_log_flush_handler;
            buffer->event->log = &cf->cycle->new_log;
            buffer->event->cancelable = 1;

            buffer->flush = flush;
        }

        buffer->gzip = gzip;

        log->file->flush = ngx_stream_log_flush;
        log->file->data = buffer;
    }

    return NGX_CONF_OK;
}

static void ngx_stream_log_flush(ngx_open_file_t *file, ngx_log_t *log) {
    size_t                 len;
    ssize_t                n;
    ngx_stream_log_buf_t  *buffer;
    buffer = file->data;
    len = buffer->pos - buffer->start;
    if (len == 0) {
        return;
    }

    n = ngx_write_fd(file->fd, buffer->start, len);
    if (n == -1) {
        ngx_log_error(NGX_LOG_ALERT, log, ngx_errno,
                      ngx_write_fd_n " to \"%s\" failed",
                      file->name.data);

    } else if ((size_t) n != len) {
        ngx_log_error(NGX_LOG_ALERT, log, 0,
                      ngx_write_fd_n " to \"%s\" was incomplete: %z of %uz",
                      file->name.data, n, len);
    }

    buffer->pos = buffer->start;
    if (buffer->event && buffer->event->timer_set) {
        ngx_del_timer(buffer->event);
    }
}


static void ngx_stream_log_flush_handler(ngx_event_t *ev) {
    ngx_log_error(NGX_LOG_DEBUG_EVENT, ev->log, 0, "stream log buffer flush handler");
    ngx_stream_log_flush(ev->data, ev->log);
}


