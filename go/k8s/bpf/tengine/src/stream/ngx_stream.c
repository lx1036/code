



#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_event.h>
#include <ngx_stream.h>
#include <nginx.h>

static char *ngx_stream_block(ngx_conf_t *cf, ngx_command_t *cmd, void *conf);
static ngx_int_t ngx_stream_init_phases(ngx_conf_t *cf,
    ngx_stream_core_main_conf_t *cmcf);
static ngx_int_t ngx_stream_init_phase_handlers(ngx_conf_t *cf,
    ngx_stream_core_main_conf_t *cmcf);
static ngx_int_t ngx_stream_add_ports(ngx_conf_t *cf, ngx_array_t *ports,
    ngx_stream_listen_t *listen);
static char *ngx_stream_optimize_servers(ngx_conf_t *cf, ngx_array_t *ports);
static ngx_int_t ngx_stream_add_addrs(ngx_conf_t *cf, ngx_stream_port_t *stport,
    ngx_stream_conf_addr_t *addr);
static ngx_int_t ngx_stream_cmp_conf_addrs(const void *one, const void *two);
static void ngx_stream_log_session(ngx_stream_session_t *s);
static void ngx_stream_close_connection(ngx_connection_t *c);
static u_char *ngx_stream_log_error(ngx_log_t *log, u_char *buf, size_t len);
static void ngx_stream_proxy_protocol_handler(ngx_event_t *rev);

static ngx_int_t ngx_stream_write_filter(ngx_stream_session_t *s, ngx_chain_t *in, ngx_uint_t from_upstream);
static ngx_int_t ngx_stream_write_filter_init(ngx_conf_t *cf);

static void *ngx_stream_core_create_main_conf(ngx_conf_t *cf);
static char *ngx_stream_core_init_main_conf(ngx_conf_t *cf, void *conf);
static void *ngx_stream_core_create_srv_conf(ngx_conf_t *cf);
static char *ngx_stream_core_merge_srv_conf(ngx_conf_t *cf, void *parent, void *child);
static char *ngx_stream_core_error_log(ngx_conf_t *cf, ngx_command_t *cmd, void *conf);
static char *ngx_stream_core_server(ngx_conf_t *cf, ngx_command_t *cmd, void *conf);
static char *ngx_stream_core_listen(ngx_conf_t *cf, ngx_command_t *cmd, void *conf);
#if (T_NGX_STREAM_SNI)
static char *ngx_stream_core_server_name(ngx_conf_t *cf, ngx_command_t *cmd, void *conf);
#endif
static ngx_stream_variable_t *ngx_stream_add_prefix_variable(ngx_conf_t *cf,
    ngx_str_t *name, ngx_uint_t flags);
static ngx_int_t ngx_stream_variable_binary_remote_addr(
    ngx_stream_session_t *s, ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_remote_addr(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_remote_port(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_proxy_protocol_addr(
    ngx_stream_session_t *s, ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_proxy_protocol_port(
    ngx_stream_session_t *s, ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_proxy_protocol_tlv(
    ngx_stream_session_t *s, ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_server_addr(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_server_port(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_bytes(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_session_time(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_status(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_connection(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_hostname(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_pid(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_msec(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_protocol(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);



ngx_uint_t  ngx_stream_max_module;
ngx_stream_filter_pt  ngx_stream_top_filter;

static ngx_command_t  ngx_stream_commands[] = {
    { ngx_string("stream"),
      NGX_MAIN_CONF|NGX_CONF_BLOCK|NGX_CONF_NOARGS,
      ngx_stream_block, // ngx_conf_file.c::ngx_conf_handler() 里 rv = cmd->set(cf, cmd, conf) 调用
      0,
      0,
      NULL },

      ngx_null_command
};
static ngx_core_module_t  ngx_stream_module_ctx = {
    ngx_string("stream"),
    NULL,
    NULL
};
ngx_module_t  ngx_stream_module = {
    NGX_MODULE_V1,
    &ngx_stream_module_ctx,                /* module context */
    ngx_stream_commands,                   /* module directives */
    NGX_CORE_MODULE,                       /* module type */
    NULL,                                  /* init master */
    NULL,                                  /* init module */
    NULL,                                  /* init process */
    NULL,                                  /* init thread */
    NULL,                                  /* exit thread */
    NULL,                                  /* exit process */
    NULL,                                  /* exit master */
    NGX_MODULE_V1_PADDING
};

static char * ngx_stream_block(ngx_conf_t *cf, ngx_command_t *cmd, void *conf) {
    char                          *rv;
    ngx_uint_t                     i, m, mi, s;
    ngx_conf_t                     pcf;
    ngx_array_t                    ports;
    ngx_stream_listen_t           *listen;
    ngx_stream_module_t           *module;
    ngx_stream_conf_ctx_t         *ctx;
    ngx_stream_core_srv_conf_t   **cscfp;
    ngx_stream_core_main_conf_t   *cmcf;

    if (*(ngx_stream_conf_ctx_t **) conf) {
        return "is duplicate";
    }

    // 1. 初始化 stream{} 和 server{} 一些变量
    ctx = ngx_pcalloc(cf->pool, sizeof(ngx_stream_conf_ctx_t));
    if (ctx == NULL) {
        return NGX_CONF_ERROR;
    }
    *(ngx_stream_conf_ctx_t **) conf = ctx;
    /* count the number of the stream modules and set up their indices */
    ngx_stream_max_module = ngx_count_modules(cf->cycle, NGX_STREAM_MODULE);
    ctx->main_conf = ngx_pcalloc(cf->pool, sizeof(void *) * ngx_stream_max_module);
    if (ctx->main_conf == NULL) {
        return NGX_CONF_ERROR;
    }
    ctx->srv_conf = ngx_pcalloc(cf->pool, sizeof(void *) * ngx_stream_max_module);
    if (ctx->srv_conf == NULL) {
        return NGX_CONF_ERROR;
    }
    for (m = 0; cf->cycle->modules[m]; m++) {
        if (cf->cycle->modules[m]->type != NGX_STREAM_MODULE) {
            continue;
        }

        module = cf->cycle->modules[m]->ctx;
        mi = cf->cycle->modules[m]->ctx_index;
        if (module->create_main_conf) {
            ctx->main_conf[mi] = module->create_main_conf(cf); // ngx_stream_core_create_main_conf
            if (ctx->main_conf[mi] == NULL) {
                return NGX_CONF_ERROR;
            }
        }
        if (module->create_srv_conf) {
            ctx->srv_conf[mi] = module->create_srv_conf(cf); // ngx_stream_core_create_srv_conf
            if (ctx->srv_conf[mi] == NULL) {
                return NGX_CONF_ERROR;
            }
        }
    }

    // 2. 初始化 stream{} 里的变量 variables, 如 remote_addr 等
    pcf = *cf;
    cf->ctx = ctx;
    for (m = 0; cf->cycle->modules[m]; m++) {
        if (cf->cycle->modules[m]->type != NGX_STREAM_MODULE) {
            continue;
        }
        module = cf->cycle->modules[m]->ctx;
        if (module->preconfiguration) {
            if (module->preconfiguration(cf) != NGX_OK) { // -> ngx_stream_variables_add_core_vars()
                return NGX_CONF_ERROR;
            }
        }
    }

    /* parse inside the stream{} block */
    cf->module_type = NGX_STREAM_MODULE;
    cf->cmd_type = NGX_STREAM_MAIN_CONF;
    rv = ngx_conf_parse(cf, NULL);
    if (rv != NGX_CONF_OK) {
        *cf = pcf;
        return rv;
    }

    // 3. 初始化 conf: ngx_stream_core_init_main_conf(), ngx_stream_core_merge_srv_conf()
    cmcf = ctx->main_conf[ngx_stream_core_module.ctx_index];
    cscfp = cmcf->servers.elts;
    for (m = 0; cf->cycle->modules[m]; m++) {
        if (cf->cycle->modules[m]->type != NGX_STREAM_MODULE) {
            continue;
        }

        module = cf->cycle->modules[m]->ctx;
        mi = cf->cycle->modules[m]->ctx_index;
        /* init stream{} main_conf's */
        cf->ctx = ctx;
        if (module->init_main_conf) {
            rv = module->init_main_conf(cf, ctx->main_conf[mi]); // ngx_stream_core_init_main_conf()
            if (rv != NGX_CONF_OK) {
                *cf = pcf;
                return rv;
            }
        }

        for (s = 0; s < cmcf->servers.nelts; s++) {
            /* merge the server{}s' srv_conf's */
            cf->ctx = cscfp[s]->ctx;
            if (module->merge_srv_conf) {
                rv = module->merge_srv_conf(cf, ctx->srv_conf[mi], cscfp[s]->ctx->srv_conf[mi]); // ngx_stream_core_merge_srv_conf()
                if (rv != NGX_CONF_OK) {
                    *cf = pcf;
                    return rv;
                }
            }
        }
    }

    if (ngx_stream_init_phases(cf, cmcf) != NGX_OK) {
        return NGX_CONF_ERROR;
    }
    // 4. 初始化 write_filter->ngx_stream_write_filter_init(), return_module 和 proxy_module 里调用
    for (m = 0; cf->cycle->modules[m]; m++) {
        if (cf->cycle->modules[m]->type != NGX_STREAM_MODULE) {
            continue;
        }

        module = cf->cycle->modules[m]->ctx;
        if (module->postconfiguration) {
            if (module->postconfiguration(cf) != NGX_OK) { // -> ngx_stream_write_filter_init()
                return NGX_CONF_ERROR;
            }
        }
    }

    if (ngx_stream_variables_init_vars(cf) != NGX_OK) {
        return NGX_CONF_ERROR;
    }

    // 5. 初始化 phase handlers, 总共有 7 个 phase
    *cf = pcf;
    if (ngx_stream_init_phase_handlers(cf, cmcf) != NGX_OK) {
        return NGX_CONF_ERROR;
    }

    // 6. 
    if (ngx_array_init(&ports, cf->temp_pool, 4, sizeof(ngx_stream_conf_port_t)) != NGX_OK) {
        return NGX_CONF_ERROR;
    }
    listen = cmcf->listen.elts;
    for (i = 0; i < cmcf->listen.nelts; i++) {
        if (ngx_stream_add_ports(cf, &ports, &listen[i]) != NGX_OK) {
            return NGX_CONF_ERROR;
        }
    }

    return ngx_stream_optimize_servers(cf, &ports);
}
static ngx_int_t ngx_stream_init_phases(ngx_conf_t *cf, ngx_stream_core_main_conf_t *cmcf) {
    if (ngx_array_init(&cmcf->phases[NGX_STREAM_POST_ACCEPT_PHASE].handlers, cf->pool, 1, sizeof(ngx_stream_handler_pt)) != NGX_OK) {
        return NGX_ERROR;
    }

    if (ngx_array_init(&cmcf->phases[NGX_STREAM_PREACCESS_PHASE].handlers, cf->pool, 1, sizeof(ngx_stream_handler_pt)) != NGX_OK) {
        return NGX_ERROR;
    }

    if (ngx_array_init(&cmcf->phases[NGX_STREAM_ACCESS_PHASE].handlers, cf->pool, 1, sizeof(ngx_stream_handler_pt)) != NGX_OK) {
        return NGX_ERROR;
    }

    if (ngx_array_init(&cmcf->phases[NGX_STREAM_SSL_PHASE].handlers, cf->pool, 1, sizeof(ngx_stream_handler_pt)) != NGX_OK) {
        return NGX_ERROR;
    }

    if (ngx_array_init(&cmcf->phases[NGX_STREAM_PREREAD_PHASE].handlers, cf->pool, 1, sizeof(ngx_stream_handler_pt)) != NGX_OK) {
        return NGX_ERROR;
    }

    if (ngx_array_init(&cmcf->phases[NGX_STREAM_LOG_PHASE].handlers, cf->pool, 1, sizeof(ngx_stream_handler_pt)) != NGX_OK) {
        return NGX_ERROR;
    }

    return NGX_OK;
}
static ngx_int_t ngx_stream_init_phase_handlers(ngx_conf_t *cf, ngx_stream_core_main_conf_t *cmcf) {
    ngx_int_t                     j;
    ngx_uint_t                    i, n;
    ngx_stream_handler_pt        *h;
    ngx_stream_phase_handler_t   *ph;
    ngx_stream_phase_handler_pt   checker;

    n = 1 /* content phase */;
    for (i = 0; i < NGX_STREAM_LOG_PHASE; i++) {
        n += cmcf->phases[i].handlers.nelts;
    }

    ph = ngx_pcalloc(cf->pool, n * sizeof(ngx_stream_phase_handler_t) + sizeof(void *));
    if (ph == NULL) {
        return NGX_ERROR;
    }

    cmcf->phase_engine.handlers = ph;
    n = 0;
    for (i = 0; i < NGX_STREAM_LOG_PHASE; i++) {
        h = cmcf->phases[i].handlers.elts;
        switch (i) {
        case NGX_STREAM_PREREAD_PHASE:
            checker = ngx_stream_core_preread_phase;
            break;

        case NGX_STREAM_CONTENT_PHASE:
            ph->checker = ngx_stream_core_content_phase;
            n++;
            ph++;
            continue;

        default:
            checker = ngx_stream_core_generic_phase;
        }

        n += cmcf->phases[i].handlers.nelts;
        for (j = cmcf->phases[i].handlers.nelts - 1; j >= 0; j--) {
            ph->checker = checker;
            ph->handler = h[j];
            ph->next = n;
            ph++;
        }
    }

    return NGX_OK;
}

static ngx_int_t ngx_stream_add_ports(ngx_conf_t *cf, ngx_array_t *ports, ngx_stream_listen_t *listen){
    in_port_t                p;
    ngx_uint_t               i;
    struct sockaddr         *sa;
    ngx_stream_conf_port_t  *port;
    ngx_stream_conf_addr_t  *addr;
#if (T_NGX_STREAM_SNI)
    ngx_stream_core_srv_conf_t *cscf;
#endif

    sa = listen->sockaddr;
    p = ngx_inet_get_port(sa);
    port = ports->elts;
    for (i = 0; i < ports->nelts; i++) {
        if (p == port[i].port && listen->type == port[i].type && sa->sa_family == port[i].family) {
            /* a port is already in the port list */
            port = &port[i];
            goto found;
        }
    }

    /* add a port to the port list */
    port = ngx_array_push(ports); // 真 low, port 指针指向数组的元素，然后赋值 port，从而实现: 赋值 port; array_push(ports, port)
    if (port == NULL) {
        return NGX_ERROR;
    }
    port->family = sa->sa_family;
    port->type = listen->type;
    port->port = p;
    if (ngx_array_init(&port->addrs, cf->temp_pool, 2, sizeof(ngx_stream_conf_addr_t)) != NGX_OK) {
        return NGX_ERROR;
    }

found:
#if (T_NGX_STREAM_SNI)
    cscf = listen->ctx->srv_conf[ngx_stream_core_module.ctx_index];
    addr = port->addrs.elts;
    for (i = 0; i < port->addrs.nelts; i++) {
        if (ngx_cmp_sockaddr(listen->sockaddr, listen->socklen,
            addr[i].opt.sockaddr,
            addr[i].opt.socklen, 0)
            != NGX_OK) {
            continue;
        }

        /* the address is already in the address list */
        if (ngx_stream_add_server(cf, cscf, &addr[i]) != NGX_OK) {
            return NGX_ERROR;
        }

        if (listen->default_server) {
            if (addr[i].opt.default_server) {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                        "a duplicate default server for");
                return NGX_ERROR;
            }
            addr[i].default_server = cscf;
            addr[i].opt.default_server = 1;
        }
        return NGX_OK;
    }

    addr = ngx_array_push(&port->addrs);
    if (addr == NULL) {
        return NGX_ERROR;
    }

    ngx_memset(addr, 0, sizeof(ngx_stream_conf_addr_t));
    addr->opt = *listen;
    addr->default_server = cscf;
    return ngx_stream_add_server(cf, cscf, addr);
#else

    addr = ngx_array_push(&port->addrs);
    if (addr == NULL) {
        return NGX_ERROR;
    }
    addr->opt = *listen;
    return NGX_OK;
#endif
}

// 这里会去创建 listener, 根据配置 "listen 4001 reuseport;"
static char * ngx_stream_optimize_servers(ngx_conf_t *cf, ngx_array_t *ports) {
    ngx_uint_t                   i, p, last, bind_wildcard;
    ngx_listening_t             *ls;
    ngx_stream_port_t           *stport;
    ngx_stream_conf_port_t      *port;
    ngx_stream_conf_addr_t      *addr;
    ngx_stream_core_srv_conf_t  *cscf;

#if (T_NGX_STREAM_SNI)
    ngx_stream_core_main_conf_t *cmcf;
#endif

    port = ports->elts;
    for (p = 0; p < ports->nelts; p++) {
        ngx_sort(port[p].addrs.elts, (size_t) port[p].addrs.nelts, sizeof(ngx_stream_conf_addr_t), ngx_stream_cmp_conf_addrs);
        addr = port[p].addrs.elts;
        last = port[p].addrs.nelts;
        /*
         * if there is the binding to the "*:port" then we need to bind()
         * to the "*:port" only and ignore the other bindings
         */
        if (addr[last - 1].opt.wildcard) {
            addr[last - 1].opt.bind = 1;
            bind_wildcard = 1;
        } else {
            bind_wildcard = 0;
        }

        i = 0;
        while (i < last) {
            if (bind_wildcard && !addr[i].opt.bind) {
                i++;
                continue;
            }
            // 这里会去创建 listener, 根据配置 "listen 4001 reuseport;"
            ls = ngx_create_listening(cf, addr[i].opt.sockaddr, addr[i].opt.socklen);
            if (ls == NULL) {
                return NGX_CONF_ERROR;
            }

            ls->addr_ntop = 1;
            // 这里注册的 listener 的 handler
            // ngx_event_tcp.c::ngx_event_accept() 会调用 ls->handler(c), ngx_event_udp.c::ngx_event_recvmsg() 会调用 ls->handler(c)
            ls->handler = ngx_stream_init_connection;
            ls->pool_size = 256;
            ls->type = addr[i].opt.type;
            cscf = addr->opt.ctx->srv_conf[ngx_stream_core_module.ctx_index];
            ls->logp = cscf->error_log;
            ls->log.data = &ls->addr_text;
            ls->log.handler = ngx_accept_log_error;
            ls->backlog = addr[i].opt.backlog;
            ls->rcvbuf = addr[i].opt.rcvbuf;
            ls->sndbuf = addr[i].opt.sndbuf;
            ls->wildcard = addr[i].opt.wildcard;
            ls->keepalive = addr[i].opt.so_keepalive;
#if (NGX_HAVE_KEEPALIVE_TUNABLE)
            ls->keepidle = addr[i].opt.tcp_keepidle;
            ls->keepintvl = addr[i].opt.tcp_keepintvl;
            ls->keepcnt = addr[i].opt.tcp_keepcnt;
#endif

#if (NGX_HAVE_TCP_FASTOPEN)
            ls->fastopen = addr[i].opt.fastopen; // -1
#endif

#if (NGX_HAVE_REUSEPORT)
            ls->reuseport = addr[i].opt.reuseport;
#endif

            stport = ngx_palloc(cf->pool, sizeof(ngx_stream_port_t));
            if (stport == NULL) {
                return NGX_CONF_ERROR;
            }
            ls->servers = stport;
            stport->naddrs = i + 1;

#if (T_NGX_STREAM_SNI)
            cmcf = addr->opt.ctx->main_conf[ngx_stream_core_module.ctx_index];
            /*Because of ssl_sni_force we have to do this even one server*/
            if (addr[i].servers.nelts >= 1) {
                if (ngx_stream_server_names(cf, cmcf, &addr[i]) != NGX_OK) {
                    return NGX_CONF_ERROR;
                }
            }
#endif

            switch (ls->sockaddr->sa_family) {
            default: /* AF_INET */
                if (ngx_stream_add_addrs(cf, stport, addr) != NGX_OK) { // addr -> "0.0.0.0:5001" -> ls->servers[]
                    return NGX_CONF_ERROR;
                }
                break;
            }

            addr++;
            last--;
        }
    }

    return NGX_CONF_OK;
}

static ngx_int_t ngx_stream_cmp_conf_addrs(const void *one, const void *two) {
    ngx_stream_conf_addr_t  *first, *second;
    first = (ngx_stream_conf_addr_t *) one;
    second = (ngx_stream_conf_addr_t *) two;
    if (first->opt.wildcard) {
        /* a wildcard must be the last resort, shift it to the end */
        return 1;
    }

    if (second->opt.wildcard) {
        /* a wildcard must be the last resort, shift it to the end */
        return -1;
    }

    if (first->opt.bind && !second->opt.bind) {
        /* shift explicit bind()ed addresses to the start */
        return -1;
    }

    if (!first->opt.bind && second->opt.bind) {
        /* shift explicit bind()ed addresses to the start */
        return 1;
    }

    /* do not sort by default */

    return 0;
}
// 添加一些server的配置
static ngx_int_t ngx_stream_add_addrs(ngx_conf_t *cf, ngx_stream_port_t *stport, ngx_stream_conf_addr_t *addr){
    ngx_uint_t             i;
    struct sockaddr_in    *sin;
    ngx_stream_in_addr_t  *addrs;
#if (T_NGX_STREAM_SNI)
    ngx_stream_virtual_names_t  *vn;
#endif

    stport->addrs = ngx_pcalloc(cf->pool, stport->naddrs * sizeof(ngx_stream_in_addr_t));
    if (stport->addrs == NULL) {
        return NGX_ERROR;
    }

    addrs = stport->addrs;
    for (i = 0; i < stport->naddrs; i++) {
        sin = (struct sockaddr_in *) addr[i].opt.sockaddr;
        addrs[i].addr = sin->sin_addr.s_addr;
        addrs[i].conf.ctx = addr[i].opt.ctx;
#if (NGX_STREAM_SSL)
        addrs[i].conf.ssl = addr[i].opt.ssl;
#endif
        // 这里初始化 proxy_protocol 配置
        addrs[i].conf.proxy_protocol = addr[i].opt.proxy_protocol;
        addrs[i].conf.addr_text = addr[i].opt.addr_text;

#if (T_NGX_STREAM_SNI)
        addrs[i].conf.default_server = addr[i].default_server;

        if (addr[i].hash.buckets == NULL
            && (addr[i].wc_head == NULL
                || addr[i].wc_head->hash.buckets == NULL)
            && (addr[i].wc_tail == NULL
                || addr[i].wc_tail->hash.buckets == NULL)
            )
        {
            continue;
        }

        vn = ngx_palloc(cf->pool, sizeof(ngx_stream_virtual_names_t));
        if (vn == NULL) {
            return NGX_ERROR;
        }

        addrs[i].conf.virtual_names = vn;

        vn->names.hash = addr[i].hash;
        vn->names.wc_head = addr[i].wc_head;
        vn->names.wc_tail = addr[i].wc_tail;
#endif
    }

    return NGX_OK;
}

void ngx_stream_session_handler(ngx_event_t *rev) {
    ngx_connection_t      *c;
    ngx_stream_session_t  *s;

    c = rev->data;
    s = c->data;

    ngx_stream_core_run_phases(s);
}
// 剥离 proxy-protocol header
static void ngx_stream_proxy_protocol_handler(ngx_event_t *rev) {
    u_char                      *p, buf[NGX_PROXY_PROTOCOL_MAX_HEADER];
    size_t                       size;
    ssize_t                      n;
    ngx_err_t                    err;
    ngx_connection_t            *c;
    ngx_stream_session_t        *s;
    ngx_stream_core_srv_conf_t  *cscf;

    c = rev->data;
    s = c->data;

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "stream PROXY protocol handler");
    if (rev->timedout) {
        ngx_log_error(NGX_LOG_STDERR, c->log, NGX_ETIMEDOUT, "client timed out");
        ngx_stream_finalize_session(s, NGX_STREAM_OK);
        return;
    }

    n = recv(c->fd, (char *) buf, sizeof(buf), MSG_PEEK);
    err = ngx_socket_errno;
    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "recv(): %z", n);
    if (n == -1) {
        if (err == NGX_EAGAIN) {
            rev->ready = 0;
            if (!rev->timer_set) {
                cscf = ngx_stream_get_module_srv_conf(s, ngx_stream_core_module);
                ngx_add_timer(rev, cscf->proxy_protocol_timeout);
            }

            if (ngx_handle_read_event(rev, 0) != NGX_OK) {
                ngx_stream_finalize_session(s, NGX_STREAM_INTERNAL_SERVER_ERROR);
            }

            return;
        }

        ngx_connection_error(c, err, "recv() failed");
        ngx_stream_finalize_session(s, NGX_STREAM_OK);
        return;
    }

    if (rev->timer_set) {
        ngx_del_timer(rev);
    }

    p = ngx_proxy_protocol_read(c, buf, buf + n); // buf: "PROXY TCP4 127.0.0.1 127.0.0.1 12345 5003\r\nhello world"
    if (p == NULL) {
        ngx_stream_finalize_session(s, NGX_STREAM_BAD_REQUEST);
        return;
    }

    size = p - buf;
    if (c->recv(c, buf, size) != (ssize_t) size) { // recv -> ngx_recv.c::ngx_unix_recv()
        ngx_stream_finalize_session(s, NGX_STREAM_INTERNAL_SERVER_ERROR);
        return;
    }

    c->log->action = "initializing session";

    ngx_stream_session_handler(rev);
}
// ngx_stream_session_handler() 处理 read event, 即处理客户端的请求报文
void ngx_stream_init_connection(ngx_connection_t *c) {
    u_char                        text[NGX_SOCKADDR_STRLEN];
    size_t                        len;
    ngx_uint_t                    i;
    ngx_time_t                   *tp;
    ngx_event_t                  *rev;
    struct sockaddr              *sa;
    ngx_stream_port_t            *port;
    struct sockaddr_in           *sin;
    ngx_stream_in_addr_t         *addr;
    ngx_stream_session_t         *s;
    ngx_stream_addr_conf_t       *addr_conf;
    ngx_stream_core_srv_conf_t   *cscf;
    ngx_stream_core_main_conf_t  *cmcf;

    /* find the server configuration for the address:port */

    port = c->listening->servers;
    if (port->naddrs > 1) {
        /*
         * There are several addresses on this port and one of them
         * is the "*:port" wildcard so getsockname() is needed to determine
         * the server address.
         *
         * AcceptEx() and recvmsg() already gave this address.
         */

        if (ngx_connection_local_sockaddr(c, NULL, 0) != NGX_OK) {
            ngx_stream_close_connection(c);
            return;
        }

        sa = c->local_sockaddr;
        switch (sa->sa_family) {
        default: /* AF_INET */
            sin = (struct sockaddr_in *) sa;
            addr = port->addrs;

            /* the last address is "*" */

            for (i = 0; i < port->naddrs - 1; i++) {
                if (addr[i].addr == sin->sin_addr.s_addr) {
                    break;
                }
            }

            addr_conf = &addr[i].conf;
            break;
        }

    } else {
        switch (c->local_sockaddr->sa_family) {
        default: /* AF_INET */
            addr = port->addrs;
            addr_conf = &addr[0].conf; // "0.0.0.0:5001", "0.0.0.0:5002"
            break;
        }
    }

    // 这里需要每一个 udp 报文必须加上 pp header
    // 这里有个奇怪的点: c->buffer 直接就拿到了请求报文数据，而且还会变化，待验证???
    if (c->type == SOCK_DGRAM && addr_conf->proxy_protocol && c->buffer != NULL) {
        c->log->action = "reading PROXY protocol for UDP";
        u_char *buf_pos = c->buffer->pos;
        u_char *buf_last = c->buffer->last;
        size_t len = buf_last - buf_pos;
        u_char *p;
        p = ngx_proxy_protocol_read(c, buf_pos, buf_last); // p 已经截断 pp header 后的报文
        if (p == NULL) {
            ngx_stream_session_t *s;
            s = c->data;
            ngx_stream_finalize_session(s, NGX_STREAM_BAD_REQUEST);
            return;
        }
        size_t pp_len = p - buf_pos; // 43="PROXY TCP4 127.0.0.1 127.0.0.1 12345 5006\r\n"
        c->buffer->pos += pp_len;
    }

    s = ngx_pcalloc(c->pool, sizeof(ngx_stream_session_t));
    if (s == NULL) {
        ngx_stream_close_connection(c);
        return;
    }

    s->signature = NGX_STREAM_MODULE;
    s->main_conf = addr_conf->ctx->main_conf;
    s->srv_conf = addr_conf->ctx->srv_conf;

#if (T_NGX_STREAM_SNI)
    s->addr_conf = addr_conf;
    s->main_conf = ((ngx_stream_core_srv_conf_t*)addr_conf->default_server)->ctx->main_conf;
    s->srv_conf = ((ngx_stream_core_srv_conf_t*)addr_conf->default_server)->ctx->srv_conf;
#endif

#if (NGX_STREAM_SSL)
    s->ssl = addr_conf->ssl;
#endif

    if (c->buffer) {
        s->received += c->buffer->last - c->buffer->pos;
    }

    s->connection = c;
    c->data = s;
    cscf = ngx_stream_get_module_srv_conf(s, ngx_stream_core_module);
    ngx_set_connection_log(c, cscf->error_log);
    len = ngx_sock_ntop(c->sockaddr, c->socklen, text, NGX_SOCKADDR_STRLEN, 1);
    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "*%uA %sclient %*s connected to %V",
                  c->number, c->type == SOCK_DGRAM ? "udp " : "tcp",
                  len, text, &addr_conf->addr_text);

    c->log->connection = c->number;
    c->log->handler = ngx_stream_log_error;
    c->log->data = s;
    c->log->action = "initializing session";
    c->log_error = NGX_ERROR_INFO;
    s->ctx = ngx_pcalloc(c->pool, sizeof(void *) * ngx_stream_max_module);
    if (s->ctx == NULL) {
        ngx_stream_close_connection(c);
        return;
    }

    cmcf = ngx_stream_get_module_main_conf(s, ngx_stream_core_module);
    s->variables = ngx_pcalloc(s->connection->pool, cmcf->variables.nelts * sizeof(ngx_stream_variable_value_t));

    if (s->variables == NULL) {
        ngx_stream_close_connection(c);
        return;
    }

    tp = ngx_timeofday();
    s->start_sec = tp->sec;
    s->start_msec = tp->msec;
    rev = c->read;
    rev->handler = ngx_stream_session_handler;

    if (c->type == SOCK_STREAM && addr_conf->proxy_protocol) {
        c->log->action = "reading PROXY protocol";
        rev->handler = ngx_stream_proxy_protocol_handler;
        if (!rev->ready) {
            ngx_add_timer(rev, cscf->proxy_protocol_timeout);
            if (ngx_handle_read_event(rev, 0) != NGX_OK) {
                ngx_stream_finalize_session(s, NGX_STREAM_INTERNAL_SERVER_ERROR);
            }

            return;
        }
    }

    if (ngx_use_accept_mutex) {
        ngx_post_event(rev, &ngx_posted_events);
        return;
    }

    rev->handler(rev); // -> ngx_stream_session_handler()
}

void ngx_stream_finalize_session(ngx_stream_session_t *s, ngx_uint_t rc) {
    ngx_log_error(NGX_LOG_STDERR, s->connection->log, 0, "finalize stream session: %i", rc);
    s->status = rc;
    ngx_stream_log_session(s);
    ngx_stream_close_connection(s->connection);
}

static u_char * ngx_stream_log_error(ngx_log_t *log, u_char *buf, size_t len) {
    u_char                *p;
    ngx_stream_session_t  *s;

    if (log->action) {
        p = ngx_snprintf(buf, len, " while %s", log->action);
        len -= p - buf;
        buf = p;
    }

    s = log->data;
    p = ngx_snprintf(buf, len, ", %sclient: %V, server: %V",
                     s->connection->type == SOCK_DGRAM ? "udp " : "",
                     &s->connection->addr_text,
                     &s->connection->listening->addr_text);
    len -= p - buf;
    buf = p;

    if (s->log_handler) {
        p = s->log_handler(log, buf, len);
    }

    return p;
}

static void ngx_stream_log_session(ngx_stream_session_t *s) {
    ngx_uint_t                    i, n;
    ngx_stream_handler_pt        *log_handler;
    ngx_stream_core_main_conf_t  *cmcf;

    cmcf = ngx_stream_get_module_main_conf(s, ngx_stream_core_module);
    log_handler = cmcf->phases[NGX_STREAM_LOG_PHASE].handlers.elts;
    n = cmcf->phases[NGX_STREAM_LOG_PHASE].handlers.nelts;

    for (i = 0; i < n; i++) {
        log_handler[i](s);
    }
}

static void ngx_stream_close_connection(ngx_connection_t *c) {
    ngx_pool_t  *pool;
    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "close stream connection: %d", c->fd);

#if (NGX_STREAM_SSL)
    if (c->ssl) {
        if (ngx_ssl_shutdown(c) == NGX_AGAIN) {
            c->ssl->handler = ngx_stream_close_connection;
            return;
        }
    }
#endif

#if (NGX_STAT_STUB)
    (void) ngx_atomic_fetch_add(ngx_stat_active, -1);
#endif

    pool = c->pool;
    ngx_close_connection(c);
    ngx_destroy_pool(pool);
}


static ngx_uint_t  ngx_stream_variable_depth = 100;
static ngx_command_t  ngx_stream_core_commands[] = {
    { ngx_string("variables_hash_max_size"),
      NGX_STREAM_MAIN_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_num_slot,
      NGX_STREAM_MAIN_CONF_OFFSET,
      offsetof(ngx_stream_core_main_conf_t, variables_hash_max_size),
      NULL },

    { ngx_string("variables_hash_bucket_size"),
      NGX_STREAM_MAIN_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_num_slot,
      NGX_STREAM_MAIN_CONF_OFFSET,
      offsetof(ngx_stream_core_main_conf_t, variables_hash_bucket_size),
      NULL },

    { ngx_string("server"),
      NGX_STREAM_MAIN_CONF|NGX_CONF_BLOCK|NGX_CONF_NOARGS,
      ngx_stream_core_server,
      0,
      0,
      NULL },

    { ngx_string("listen"),
      NGX_STREAM_SRV_CONF|NGX_CONF_1MORE,
      ngx_stream_core_listen,
      NGX_STREAM_SRV_CONF_OFFSET,
      0,
      NULL },

    { ngx_string("error_log"),
      NGX_STREAM_MAIN_CONF|NGX_STREAM_SRV_CONF|NGX_CONF_1MORE,
      ngx_stream_core_error_log,
      NGX_STREAM_SRV_CONF_OFFSET,
      0,
      NULL },

    { ngx_string("proxy_protocol_timeout"),
      NGX_STREAM_MAIN_CONF|NGX_STREAM_SRV_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_msec_slot,
      NGX_STREAM_SRV_CONF_OFFSET,
      offsetof(ngx_stream_core_srv_conf_t, proxy_protocol_timeout),
      NULL },

    { ngx_string("tcp_nodelay"),
      NGX_STREAM_MAIN_CONF|NGX_STREAM_SRV_CONF|NGX_CONF_FLAG,
      ngx_conf_set_flag_slot,
      NGX_STREAM_SRV_CONF_OFFSET,
      offsetof(ngx_stream_core_srv_conf_t, tcp_nodelay),
      NULL },

    { ngx_string("preread_buffer_size"),
      NGX_STREAM_MAIN_CONF|NGX_STREAM_SRV_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_size_slot,
      NGX_STREAM_SRV_CONF_OFFSET,
      offsetof(ngx_stream_core_srv_conf_t, preread_buffer_size),
      NULL },

    { ngx_string("preread_timeout"),
      NGX_STREAM_MAIN_CONF|NGX_STREAM_SRV_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_msec_slot,
      NGX_STREAM_SRV_CONF_OFFSET,
      offsetof(ngx_stream_core_srv_conf_t, preread_timeout),
      NULL },

#if (T_NGX_STREAM_SNI)
    { ngx_string("server_name"),
      NGX_STREAM_SRV_CONF|NGX_CONF_1MORE,
      ngx_stream_core_server_name,
      NGX_STREAM_SRV_CONF_OFFSET,
      0,
      NULL },

    { ngx_string("server_names_hash_max_size"),
      NGX_STREAM_MAIN_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_num_slot,
      NGX_STREAM_MAIN_CONF_OFFSET,
      offsetof(ngx_stream_core_main_conf_t, server_names_hash_max_size),
      NULL },

    { ngx_string("server_names_hash_bucket_size"),
      NGX_STREAM_MAIN_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_num_slot,
      NGX_STREAM_MAIN_CONF_OFFSET,
      offsetof(ngx_stream_core_main_conf_t, server_names_hash_bucket_size),
      NULL },

#endif
      ngx_null_command
};
static ngx_stream_module_t  ngx_stream_core_module_ctx = {
    ngx_stream_variables_add_core_vars,    /* preconfiguration */
    NULL,                                  /* postconfiguration */

    ngx_stream_core_create_main_conf,      /* create main configuration */
    ngx_stream_core_init_main_conf,        /* init main configuration */

    ngx_stream_core_create_srv_conf,       /* create server configuration */
    ngx_stream_core_merge_srv_conf         /* merge server configuration */
};
static void * ngx_stream_core_create_main_conf(ngx_conf_t *cf) {
    ngx_stream_core_main_conf_t  *cmcf;
    cmcf = ngx_pcalloc(cf->pool, sizeof(ngx_stream_core_main_conf_t));
    if (cmcf == NULL) {
        return NULL;
    }

    if (ngx_array_init(&cmcf->servers, cf->pool, 4, sizeof(ngx_stream_core_srv_conf_t *)) != NGX_OK) {
        return NULL;
    }

    if (ngx_array_init(&cmcf->listen, cf->pool, 4, sizeof(ngx_stream_listen_t)) != NGX_OK) {
        return NULL;
    }

    cmcf->variables_hash_max_size = NGX_CONF_UNSET_UINT;
    cmcf->variables_hash_bucket_size = NGX_CONF_UNSET_UINT;

#if (T_NGX_STREAM_SNI)
    cmcf->server_names_hash_max_size = NGX_CONF_UNSET_UINT;
    cmcf->server_names_hash_bucket_size = NGX_CONF_UNSET_UINT;
#endif

    return cmcf;
}
static char * ngx_stream_core_init_main_conf(ngx_conf_t *cf, void *conf) {
    ngx_stream_core_main_conf_t *cmcf = conf;
    ngx_conf_init_uint_value(cmcf->variables_hash_max_size, 1024);
    ngx_conf_init_uint_value(cmcf->variables_hash_bucket_size, 64);

#if (T_NGX_STREAM_SNI)
    ngx_conf_init_uint_value(cmcf->server_names_hash_max_size, 512);
    ngx_conf_init_uint_value(cmcf->server_names_hash_bucket_size, ngx_cacheline_size);
#endif

    cmcf->variables_hash_bucket_size = ngx_align(cmcf->variables_hash_bucket_size, ngx_cacheline_size);

    if (cmcf->ncaptures) {
        cmcf->ncaptures = (cmcf->ncaptures + 1) * 3;
    }

    return NGX_CONF_OK;
}
static void * ngx_stream_core_create_srv_conf(ngx_conf_t *cf){
    ngx_stream_core_srv_conf_t  *cscf;
    cscf = ngx_pcalloc(cf->pool, sizeof(ngx_stream_core_srv_conf_t));
    if (cscf == NULL) {
        return NULL;
    }

    /*
     * set by ngx_pcalloc():
     *
     *     cscf->handler = NULL;
     *     cscf->error_log = NULL;
     */

    cscf->file_name = cf->conf_file->file.name.data;
    cscf->line = cf->conf_file->line;
    cscf->resolver_timeout = NGX_CONF_UNSET_MSEC;
    cscf->proxy_protocol_timeout = NGX_CONF_UNSET_MSEC;
    cscf->tcp_nodelay = NGX_CONF_UNSET;
    cscf->preread_buffer_size = NGX_CONF_UNSET_SIZE;
    cscf->preread_timeout = NGX_CONF_UNSET_MSEC;

#if (T_NGX_STREAM_SNI)
    if (ngx_array_init(&cscf->server_names, cf->temp_pool, 4, sizeof(ngx_stream_server_name_t)) != NGX_OK) {
        return NULL;
    }
#endif

    return cscf;
}
static char * ngx_stream_core_merge_srv_conf(ngx_conf_t *cf, void *parent, void *child) {
    ngx_stream_core_srv_conf_t *prev = parent;
    ngx_stream_core_srv_conf_t *conf = child;
#if (T_NGX_STREAM_SNI)
    ngx_str_t                   name;
    ngx_stream_server_name_t   *sn;
#endif

    ngx_conf_merge_msec_value(conf->resolver_timeout, prev->resolver_timeout, 30000);
    if (conf->resolver == NULL) {
        if (prev->resolver == NULL) {
            /*
             * create dummy resolver in stream {} context
             * to inherit it in all servers
             */
            prev->resolver = ngx_resolver_create(cf, NULL, 0);
            if (prev->resolver == NULL) {
                return NGX_CONF_ERROR;
            }
        }

        conf->resolver = prev->resolver;
    }

    if (conf->handler == NULL) {
        ngx_log_error(NGX_LOG_STDERR, cf->log, 0, "no handler for server in %s:%ui", conf->file_name, conf->line);
        return NGX_CONF_ERROR;
    }

    if (conf->error_log == NULL) {
        if (prev->error_log) {
            conf->error_log = prev->error_log;
        } else {
            conf->error_log = &cf->cycle->new_log;
        }
    }

    ngx_conf_merge_msec_value(conf->proxy_protocol_timeout, prev->proxy_protocol_timeout, 30000);
    ngx_conf_merge_value(conf->tcp_nodelay, prev->tcp_nodelay, 1);
    ngx_conf_merge_size_value(conf->preread_buffer_size, prev->preread_buffer_size, 16384);
    ngx_conf_merge_msec_value(conf->preread_timeout, prev->preread_timeout, 30000);

#if (T_NGX_STREAM_SNI)
    if (conf->server_names.nelts == 0) {
        /* the array has 4 empty preallocated elements, so push cannot fail */
        sn = ngx_array_push(&conf->server_names);
        sn->server = conf;
        ngx_str_set(&sn->name, "");
    }

    sn = conf->server_names.elts;
    name = sn[0].name;

    if (name.data[0] == '.') {
        name.len--;
        name.data++;
    }

    conf->server_name.len = name.len;
    conf->server_name.data = ngx_pstrdup(cf->pool, &name);
    if (conf->server_name.data == NULL) {
        return NGX_CONF_ERROR;
    }
#endif

    return NGX_CONF_OK;
}
ngx_module_t  ngx_stream_core_module = {
    NGX_MODULE_V1,
    &ngx_stream_core_module_ctx,           /* module context */
    ngx_stream_core_commands,              /* module directives */
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
static ngx_stream_variable_t  ngx_stream_core_variables[] = {

    { ngx_string("binary_remote_addr"), NULL,
      ngx_stream_variable_binary_remote_addr, 0, 0, 0 },

    { ngx_string("remote_addr"), NULL,
      ngx_stream_variable_remote_addr, 0, 0, 0 },

    { ngx_string("remote_port"), NULL,
      ngx_stream_variable_remote_port, 0, 0, 0 },

    { ngx_string("proxy_protocol_addr"), NULL,
      ngx_stream_variable_proxy_protocol_addr,
      offsetof(ngx_proxy_protocol_t, src_addr), 0, 0 },

    { ngx_string("proxy_protocol_port"), NULL,
      ngx_stream_variable_proxy_protocol_port,
      offsetof(ngx_proxy_protocol_t, src_port), 0, 0 },

    { ngx_string("proxy_protocol_server_addr"), NULL,
      ngx_stream_variable_proxy_protocol_addr,
      offsetof(ngx_proxy_protocol_t, dst_addr), 0, 0 },

    { ngx_string("proxy_protocol_server_port"), NULL,
      ngx_stream_variable_proxy_protocol_port,
      offsetof(ngx_proxy_protocol_t, dst_port), 0, 0 },

    { ngx_string("proxy_protocol_tlv_"), NULL,
      ngx_stream_variable_proxy_protocol_tlv,
      0, NGX_STREAM_VAR_PREFIX, 0 },

    { ngx_string("server_addr"), NULL,
      ngx_stream_variable_server_addr, 0, 0, 0 },

    { ngx_string("server_port"), NULL,
      ngx_stream_variable_server_port, 0, 0, 0 },

    { ngx_string("bytes_sent"), NULL, ngx_stream_variable_bytes,
      0, 0, 0 },

    { ngx_string("bytes_received"), NULL, ngx_stream_variable_bytes,
      1, 0, 0 },

    { ngx_string("session_time"), NULL, ngx_stream_variable_session_time,
      0, NGX_STREAM_VAR_NOCACHEABLE, 0 },

    { ngx_string("status"), NULL, ngx_stream_variable_status,
      0, NGX_STREAM_VAR_NOCACHEABLE, 0 },

    { ngx_string("connection"), NULL,
      ngx_stream_variable_connection, 0, 0, 0 },

    { ngx_string("hostname"), NULL, ngx_stream_variable_hostname,
      0, 0, 0 },

    { ngx_string("pid"), NULL, ngx_stream_variable_pid,
      0, 0, 0 },

    { ngx_string("msec"), NULL, ngx_stream_variable_msec,
      0, NGX_STREAM_VAR_NOCACHEABLE, 0 },

    { ngx_string("protocol"), NULL,
      ngx_stream_variable_protocol, 0, 0, 0 },

      ngx_stream_null_variable
};
ngx_int_t ngx_stream_variables_add_core_vars(ngx_conf_t *cf) {
    ngx_stream_variable_t        *cv, *v;
    ngx_stream_core_main_conf_t  *cmcf;

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);
    cmcf->variables_keys = ngx_pcalloc(cf->temp_pool, sizeof(ngx_hash_keys_arrays_t));
    if (cmcf->variables_keys == NULL) {
        return NGX_ERROR;
    }

    cmcf->variables_keys->pool = cf->pool;
    cmcf->variables_keys->temp_pool = cf->pool;
    if (ngx_hash_keys_array_init(cmcf->variables_keys, NGX_HASH_SMALL) != NGX_OK) {
        return NGX_ERROR;
    }

    if (ngx_array_init(&cmcf->prefix_variables, cf->pool, 8, sizeof(ngx_stream_variable_t)) != NGX_OK) {
        return NGX_ERROR;
    }

    for (cv = ngx_stream_core_variables; cv->name.len; cv++) {
        v = ngx_stream_add_variable(cf, &cv->name, cv->flags);
        if (v == NULL) {
            return NGX_ERROR;
        }

        *v = *cv;
    }

    return NGX_OK;
}
ngx_stream_variable_t * ngx_stream_add_variable(ngx_conf_t *cf, ngx_str_t *name, ngx_uint_t flags) {
    ngx_int_t                     rc;
    ngx_uint_t                    i;
    ngx_hash_key_t               *key;
    ngx_stream_variable_t        *v;
    ngx_stream_core_main_conf_t  *cmcf;

    if (name->len == 0) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "invalid variable name \"$\"");
        return NULL;
    }

    if (flags & NGX_STREAM_VAR_PREFIX) {
        return ngx_stream_add_prefix_variable(cf, name, flags);
    }

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);
    key = cmcf->variables_keys->keys.elts;
    for (i = 0; i < cmcf->variables_keys->keys.nelts; i++) {
        if (name->len != key[i].key.len || ngx_strncasecmp(name->data, key[i].key.data, name->len) != 0) {
            continue;
        }

        v = key[i].value;
        if (!(v->flags & NGX_STREAM_VAR_CHANGEABLE)) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "the duplicate \"%V\" variable", name);
            return NULL;
        }

        if (!(flags & NGX_STREAM_VAR_WEAK)) {
            v->flags &= ~NGX_STREAM_VAR_WEAK;
        }

        return v;
    }

    v = ngx_palloc(cf->pool, sizeof(ngx_stream_variable_t));
    if (v == NULL) {
        return NULL;
    }

    v->name.len = name->len;
    v->name.data = ngx_pnalloc(cf->pool, name->len);
    if (v->name.data == NULL) {
        return NULL;
    }

    ngx_strlow(v->name.data, name->data, name->len);
    v->set_handler = NULL;
    v->get_handler = NULL;
    v->data = 0;
    v->flags = flags;
    v->index = 0;
    rc = ngx_hash_add_key(cmcf->variables_keys, &v->name, v, 0);
    if (rc == NGX_ERROR) {
        return NULL;
    }
    if (rc == NGX_BUSY) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "conflicting variable name \"%V\"", name);
        return NULL;
    }

    return v;
}
// "proxy_protocol_tlv_xxx" 变量使用
static ngx_stream_variable_t * ngx_stream_add_prefix_variable(ngx_conf_t *cf, ngx_str_t *name, ngx_uint_t flags) {
    ngx_uint_t                    i;
    ngx_stream_variable_t        *v;
    ngx_stream_core_main_conf_t  *cmcf;

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);
    v = cmcf->prefix_variables.elts;
    for (i = 0; i < cmcf->prefix_variables.nelts; i++) {
        if (name->len != v[i].name.len || ngx_strncasecmp(name->data, v[i].name.data, name->len) != 0) {
            continue;
        }

        v = &v[i];
        if (!(v->flags & NGX_STREAM_VAR_CHANGEABLE)) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "the duplicate \"%V\" variable", name);
            return NULL;
        }

        if (!(flags & NGX_STREAM_VAR_WEAK)) {
            v->flags &= ~NGX_STREAM_VAR_WEAK;
        }

        return v;
    }

    v = ngx_array_push(&cmcf->prefix_variables);
    if (v == NULL) {
        return NULL;
    }

    v->name.len = name->len;
    v->name.data = ngx_pnalloc(cf->pool, name->len);
    if (v->name.data == NULL) {
        return NULL;
    }

    ngx_strlow(v->name.data, name->data, name->len);
    v->set_handler = NULL;
    v->get_handler = NULL;
    v->data = 0;
    v->flags = flags;
    v->index = 0;

    return v;
}
static ngx_int_t ngx_stream_variable_binary_remote_addr(ngx_stream_session_t *s, ngx_stream_variable_value_t *v, uintptr_t data) {
    struct sockaddr_in   *sin;
    switch (s->connection->sockaddr->sa_family) {
    default: /* AF_INET */
        sin = (struct sockaddr_in *) s->connection->sockaddr; // socket addr
        v->len = sizeof(in_addr_t);
        v->valid = 1;
        v->no_cacheable = 0;
        v->not_found = 0;
        v->data = (u_char *) &sin->sin_addr; // source address

        break;
    }

    return NGX_OK;
}

static ngx_int_t ngx_stream_variable_remote_addr(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    v->len = s->connection->addr_text.len;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = s->connection->addr_text.data;

    return NGX_OK;
}

static ngx_int_t ngx_stream_variable_remote_port(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    ngx_uint_t  port;
    v->len = 0;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;

    v->data = ngx_pnalloc(s->connection->pool, sizeof("65535") - 1);
    if (v->data == NULL) {
        return NGX_ERROR;
    }

    port = ngx_inet_get_port(s->connection->sockaddr);
    if (port > 0 && port < 65536) {
        v->len = ngx_sprintf(v->data, "%ui", port) - v->data;
    }

    return NGX_OK;
}

// https://docs.nginx.com/nginx/admin-guide/load-balancer/using-proxy-protocol/#configuring-nginx-to-accept-the-proxy-protocol
static ngx_int_t ngx_stream_variable_proxy_protocol_addr(ngx_stream_session_t *s, ngx_stream_variable_value_t *v, uintptr_t data) {
    ngx_str_t             *addr;
    ngx_proxy_protocol_t  *pp;

    pp = s->connection->proxy_protocol;
    if (pp == NULL) {
        v->not_found = 1;
        return NGX_OK;
    }

    addr = (ngx_str_t *) ((char *) pp + data);
    v->len = addr->len;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = addr->data;

    return NGX_OK;
}
static ngx_int_t ngx_stream_variable_proxy_protocol_port(ngx_stream_session_t *s, ngx_stream_variable_value_t *v, uintptr_t data) {
    ngx_uint_t             port;
    ngx_proxy_protocol_t  *pp;

    pp = s->connection->proxy_protocol;
    if (pp == NULL) {
        v->not_found = 1;
        return NGX_OK;
    }

    v->len = 0;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = ngx_pnalloc(s->connection->pool, sizeof("65535") - 1);
    if (v->data == NULL) {
        return NGX_ERROR;
    }

    port = *(in_port_t *) ((char *) pp + data);
    if (port > 0 && port < 65536) {
        v->len = ngx_sprintf(v->data, "%ui", port) - v->data;
    }

    return NGX_OK;
}

static ngx_int_t
ngx_stream_variable_proxy_protocol_tlv(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data)
{
    ngx_str_t *name = (ngx_str_t *) data;

    ngx_int_t  rc;
    ngx_str_t  tlv, value;

    tlv.len = name->len - (sizeof("proxy_protocol_tlv_") - 1);
    tlv.data = name->data + sizeof("proxy_protocol_tlv_") - 1;

    rc = ngx_proxy_protocol_get_tlv(s->connection, &tlv, &value);
    if (rc == NGX_ERROR) {
        return NGX_ERROR;
    }

    if (rc == NGX_DECLINED) {
        v->not_found = 1;
        return NGX_OK;
    }

    v->len = value.len;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = value.data;

    return NGX_OK;
}


static ngx_int_t
ngx_stream_variable_server_addr(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data)
{
    ngx_str_t  str;
    u_char     addr[NGX_SOCKADDR_STRLEN];

    str.len = NGX_SOCKADDR_STRLEN;
    str.data = addr;

    if (ngx_connection_local_sockaddr(s->connection, &str, 0) != NGX_OK) {
        return NGX_ERROR;
    }

    str.data = ngx_pnalloc(s->connection->pool, str.len);
    if (str.data == NULL) {
        return NGX_ERROR;
    }

    ngx_memcpy(str.data, addr, str.len);

    v->len = str.len;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = str.data;

    return NGX_OK;
}


static ngx_int_t
ngx_stream_variable_server_port(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    ngx_uint_t  port;

    v->len = 0;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;

    if (ngx_connection_local_sockaddr(s->connection, NULL, 0) != NGX_OK) {
        return NGX_ERROR;
    }

    v->data = ngx_pnalloc(s->connection->pool, sizeof("65535") - 1);
    if (v->data == NULL) {
        return NGX_ERROR;
    }

    port = ngx_inet_get_port(s->connection->local_sockaddr);
    if (port > 0 && port < 65536) {
        v->len = ngx_sprintf(v->data, "%ui", port) - v->data;
    }

    return NGX_OK;
}

static ngx_int_t
ngx_stream_variable_bytes(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data)
{
    u_char  *p;

    p = ngx_pnalloc(s->connection->pool, NGX_OFF_T_LEN);
    if (p == NULL) {
        return NGX_ERROR;
    }

    if (data == 1) {
        v->len = ngx_sprintf(p, "%O", s->received) - p;
    } else {
        v->len = ngx_sprintf(p, "%O", s->connection->sent) - p;
    }

    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = p;

    return NGX_OK;
}

static ngx_int_t ngx_stream_variable_session_time(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    u_char          *p;
    ngx_time_t      *tp;
    ngx_msec_int_t   ms;

    p = ngx_pnalloc(s->connection->pool, NGX_TIME_T_LEN + 4);
    if (p == NULL) {
        return NGX_ERROR;
    }

    tp = ngx_timeofday();
    ms = (ngx_msec_int_t)
             ((tp->sec - s->start_sec) * 1000 + (tp->msec - s->start_msec));
    ms = ngx_max(ms, 0);
    v->len = ngx_sprintf(p, "%T.%03M", (time_t) ms / 1000, ms % 1000) - p;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = p;

    return NGX_OK;
}

static ngx_int_t
ngx_stream_variable_status(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data)
{
    v->data = ngx_pnalloc(s->connection->pool, NGX_INT_T_LEN);
    if (v->data == NULL) {
        return NGX_ERROR;
    }

    v->len = ngx_sprintf(v->data, "%03ui", s->status) - v->data;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;

    return NGX_OK;
}

static ngx_int_t ngx_stream_variable_connection(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    u_char  *p;
    p = ngx_pnalloc(s->connection->pool, NGX_ATOMIC_T_LEN);
    if (p == NULL) {
        return NGX_ERROR;
    }

    v->len = ngx_sprintf(p, "%uA", s->connection->number) - p;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = p;

    return NGX_OK;
}

static ngx_int_t ngx_stream_variable_hostname(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    v->len = ngx_cycle->hostname.len;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = ngx_cycle->hostname.data;

    return NGX_OK;
}


static ngx_int_t ngx_stream_variable_pid(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    u_char  *p;
    p = ngx_pnalloc(s->connection->pool, NGX_INT64_LEN);
    if (p == NULL) {
        return NGX_ERROR;
    }

    v->len = ngx_sprintf(p, "%P", ngx_pid) - p;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = p;

    return NGX_OK;
}

static ngx_int_t ngx_stream_variable_msec(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    u_char      *p;
    ngx_time_t  *tp;

    p = ngx_pnalloc(s->connection->pool, NGX_TIME_T_LEN + 4);
    if (p == NULL) {
        return NGX_ERROR;
    }

    tp = ngx_timeofday();
    v->len = ngx_sprintf(p, "%T.%03M", tp->sec, tp->msec) - p;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = p;

    return NGX_OK;
}

static ngx_int_t ngx_stream_variable_protocol(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data) {
    v->len = 3;
    v->valid = 1;
    v->no_cacheable = 0;
    v->not_found = 0;
    v->data = (u_char *) (s->connection->type == SOCK_DGRAM ? "UDP" : "TCP");

    return NGX_OK;
}

ngx_int_t ngx_stream_variables_init_vars(ngx_conf_t *cf) {
    size_t                        len;
    ngx_uint_t                    i, n;
    ngx_hash_key_t               *key;
    ngx_hash_init_t               hash;
    ngx_stream_variable_t        *v, *av, *pv;
    ngx_stream_core_main_conf_t  *cmcf;

    /* set the handlers for the indexed stream variables */

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);
    v = cmcf->variables.elts;
    pv = cmcf->prefix_variables.elts;
    key = cmcf->variables_keys->keys.elts;

    for (i = 0; i < cmcf->variables.nelts; i++) {
        for (n = 0; n < cmcf->variables_keys->keys.nelts; n++) {
            av = key[n].value;
            if (v[i].name.len == key[n].key.len
                && ngx_strncmp(v[i].name.data, key[n].key.data, v[i].name.len)
                   == 0) {
                v[i].get_handler = av->get_handler;
                v[i].data = av->data;
                av->flags |= NGX_STREAM_VAR_INDEXED;
                v[i].flags = av->flags;
                av->index = i;
                if (av->get_handler == NULL
                    || (av->flags & NGX_STREAM_VAR_WEAK)) {
                    break;
                }

                goto next;
            }
        }

        len = 0;
        av = NULL;
        for (n = 0; n < cmcf->prefix_variables.nelts; n++) {
            if (v[i].name.len >= pv[n].name.len && v[i].name.len > len
                && ngx_strncmp(v[i].name.data, pv[n].name.data, pv[n].name.len)
                   == 0) {
                av = &pv[n];
                len = pv[n].name.len;
            }
        }

        if (av) {
            v[i].get_handler = av->get_handler;
            v[i].data = (uintptr_t) &v[i].name;
            v[i].flags = av->flags;

            goto next;
         }

        if (v[i].get_handler == NULL) {
            ngx_log_error(NGX_LOG_STDERR, cf->log, 0, "unknown \"%V\" variable", &v[i].name);
            return NGX_ERROR;
        }

    next:
        continue;
    }

    for (n = 0; n < cmcf->variables_keys->keys.nelts; n++) {
        av = key[n].value;
        if (av->flags & NGX_STREAM_VAR_NOHASH) {
            key[n].key.data = NULL;
        }
    }

    hash.hash = &cmcf->variables_hash;
    hash.key = ngx_hash_key;
    hash.max_size = cmcf->variables_hash_max_size;
    hash.bucket_size = cmcf->variables_hash_bucket_size;
    hash.name = "variables_hash";
    hash.pool = cf->pool;
    hash.temp_pool = NULL;
    if (ngx_hash_init(&hash, cmcf->variables_keys->keys.elts, cmcf->variables_keys->keys.nelts) != NGX_OK) {
        return NGX_ERROR;
    }

    cmcf->variables_keys = NULL;

    return NGX_OK;
}

ngx_int_t
ngx_stream_get_variable_index(ngx_conf_t *cf, ngx_str_t *name)
{
    ngx_uint_t                    i;
    ngx_stream_variable_t        *v;
    ngx_stream_core_main_conf_t  *cmcf;

    if (name->len == 0) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                           "invalid variable name \"$\"");
        return NGX_ERROR;
    }

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);

    v = cmcf->variables.elts;

    if (v == NULL) {
        if (ngx_array_init(&cmcf->variables, cf->pool, 4,
                           sizeof(ngx_stream_variable_t))
            != NGX_OK)
        {
            return NGX_ERROR;
        }

    } else {
        for (i = 0; i < cmcf->variables.nelts; i++) {
            if (name->len != v[i].name.len
                || ngx_strncasecmp(name->data, v[i].name.data, name->len) != 0)
            {
                continue;
            }

            return i;
        }
    }

    v = ngx_array_push(&cmcf->variables);
    if (v == NULL) {
        return NGX_ERROR;
    }

    v->name.len = name->len;
    v->name.data = ngx_pnalloc(cf->pool, name->len);
    if (v->name.data == NULL) {
        return NGX_ERROR;
    }

    ngx_strlow(v->name.data, name->data, name->len);

    v->set_handler = NULL;
    v->get_handler = NULL;
    v->data = 0;
    v->flags = 0;
    v->index = cmcf->variables.nelts - 1;

    return v->index;
}

ngx_stream_variable_value_t *
ngx_stream_get_indexed_variable(ngx_stream_session_t *s, ngx_uint_t index)
{
    ngx_stream_variable_t        *v;
    ngx_stream_core_main_conf_t  *cmcf;

    cmcf = ngx_stream_get_module_main_conf(s, ngx_stream_core_module);
    if (cmcf->variables.nelts <= index) {
        ngx_log_error(NGX_LOG_STDERR, s->connection->log, 0, "unknown variable index: %ui", index);
        return NULL;
    }

    if (s->variables[index].not_found || s->variables[index].valid) {
        return &s->variables[index];
    }

    v = cmcf->variables.elts;

    if (ngx_stream_variable_depth == 0) {
        ngx_log_error(NGX_LOG_ERR, s->connection->log, 0, "cycle while evaluating variable \"%V\"", &v[index].name);
        return NULL;
    }

    ngx_stream_variable_depth--;

    if (v[index].get_handler(s, &s->variables[index], v[index].data)
        == NGX_OK)
    {
        ngx_stream_variable_depth++;

        if (v[index].flags & NGX_STREAM_VAR_NOCACHEABLE) {
            s->variables[index].no_cacheable = 1;
        }

        return &s->variables[index];
    }

    ngx_stream_variable_depth++;

    s->variables[index].valid = 0;
    s->variables[index].not_found = 1;

    return NULL;
}

ngx_stream_variable_value_t *
ngx_stream_get_flushed_variable(ngx_stream_session_t *s, ngx_uint_t index)
{
    ngx_stream_variable_value_t  *v;

    v = &s->variables[index];

    if (v->valid || v->not_found) {
        if (!v->no_cacheable) {
            return v;
        }

        v->valid = 0;
        v->not_found = 0;
    }

    return ngx_stream_get_indexed_variable(s, index);
}

size_t ngx_stream_script_copy_var_len_code(ngx_stream_script_engine_t *e) {
    ngx_stream_variable_value_t   *value;
    ngx_stream_script_var_code_t  *code;

    code = (ngx_stream_script_var_code_t *) e->ip;

    e->ip += sizeof(ngx_stream_script_var_code_t);

    if (e->flushed) {
        value = ngx_stream_get_indexed_variable(e->session, code->index);

    } else {
        value = ngx_stream_get_flushed_variable(e->session, code->index);
    }

    if (value && !value->not_found) {
        return value->len;
    }

    return 0;
}

void ngx_stream_script_copy_var_code(ngx_stream_script_engine_t *e) {
    u_char                        *p;
    ngx_stream_variable_value_t   *value;
    ngx_stream_script_var_code_t  *code;

    code = (ngx_stream_script_var_code_t *) e->ip;

    e->ip += sizeof(ngx_stream_script_var_code_t);

    if (!e->skip) {

        if (e->flushed) {
            value = ngx_stream_get_indexed_variable(e->session, code->index);

        } else {
            value = ngx_stream_get_flushed_variable(e->session, code->index);
        }

        if (value && !value->not_found) {
            p = e->pos;
            e->pos = ngx_copy(p, value->data, value->len);
            ngx_log_error(NGX_LOG_STDERR, e->session->connection->log, 0, "stream script var: \"%*s\"", e->pos - p, p);
        }
    }
}


static char * ngx_stream_core_server(ngx_conf_t *cf, ngx_command_t *cmd, void *conf) {
    char                         *rv;
    void                         *mconf;
    ngx_uint_t                    m;
    ngx_conf_t                    pcf;
    ngx_stream_module_t          *module;
    ngx_stream_conf_ctx_t        *ctx, *stream_ctx;
    ngx_stream_core_srv_conf_t   *cscf, **cscfp;
    ngx_stream_core_main_conf_t  *cmcf;

    ctx = ngx_pcalloc(cf->pool, sizeof(ngx_stream_conf_ctx_t));
    if (ctx == NULL) {
        return NGX_CONF_ERROR;
    }

    stream_ctx = cf->ctx;
    ctx->main_conf = stream_ctx->main_conf;

    /* the server{}'s srv_conf */
    ctx->srv_conf = ngx_pcalloc(cf->pool, sizeof(void *) * ngx_stream_max_module);
    if (ctx->srv_conf == NULL) {
        return NGX_CONF_ERROR;
    }

    for (m = 0; cf->cycle->modules[m]; m++) {
        if (cf->cycle->modules[m]->type != NGX_STREAM_MODULE) {
            continue;
        }

        module = cf->cycle->modules[m]->ctx;
        if (module->create_srv_conf) {
            mconf = module->create_srv_conf(cf); // ngx_stream_core_create_srv_conf
            if (mconf == NULL) {
                return NGX_CONF_ERROR;
            }

            ctx->srv_conf[cf->cycle->modules[m]->ctx_index] = mconf;
        }
    }

    /* the server configuration context */
    cscf = ctx->srv_conf[ngx_stream_core_module.ctx_index];
    cscf->ctx = ctx;
    cmcf = ctx->main_conf[ngx_stream_core_module.ctx_index];
    cscfp = ngx_array_push(&cmcf->servers);
    if (cscfp == NULL) {
        return NGX_CONF_ERROR;
    }

    *cscfp = cscf;
    /* parse inside server{} */
    pcf = *cf;
    cf->ctx = ctx;
    cf->cmd_type = NGX_STREAM_SRV_CONF;
    rv = ngx_conf_parse(cf, NULL); // parse server{} 里的 listen 等指令，会跳转到 ngx_stream_core_listen
    *cf = pcf;
    if (rv == NGX_CONF_OK && !cscf->listen) {
        ngx_log_error(NGX_LOG_STDERR, cf->log, 0, "no \"listen\" is defined for server in %s:%ui", cscf->file_name, cscf->line);
        return NGX_CONF_ERROR;
    }

    return rv;
}
static char * ngx_stream_core_listen(ngx_conf_t *cf, ngx_command_t *cmd, void *conf) {
    ngx_stream_core_srv_conf_t  *cscf = conf;
    ngx_str_t                    *value, size;
    ngx_url_t                     u;
    ngx_uint_t                    i, n, backlog;
    ngx_stream_listen_t          *ls, *als, *nls;
    ngx_stream_core_main_conf_t  *cmcf;

    cscf->listen = 1;
    value = cf->args->elts;
    ngx_memzero(&u, sizeof(ngx_url_t));
    u.url = value[1]; // 4001
    u.listen = 1;
    if (ngx_parse_url(cf->pool, &u) != NGX_OK) {
        if (u.err) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                               "%s in \"%V\" of the \"listen\" directive",
                               u.err, &u.url);
        }

        return NGX_CONF_ERROR;
    }

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);
    ls = ngx_array_push(&cmcf->listen);
    if (ls == NULL) {
        return NGX_CONF_ERROR;
    }

    ngx_memzero(ls, sizeof(ngx_stream_listen_t));
    ls->backlog = NGX_LISTEN_BACKLOG;
    ls->rcvbuf = -1;
    ls->sndbuf = -1;
    ls->type = SOCK_STREAM;
    ls->ctx = cf->ctx;

#if (NGX_HAVE_TCP_FASTOPEN)
    ls->fastopen = -1;
#endif

    backlog = 0;
    for (i = 2; i < cf->args->nelts; i++) {
        if (ngx_strcmp(value[i].data, "udp") == 0) {
            ls->type = SOCK_DGRAM;
            continue;
        }

#if (T_NGX_STREAM_SNI)
        if (ngx_strcmp(value[i].data, "default_server") == 0
            || ngx_strcmp(value[i].data, "default") == 0) {
            ls->default_server = 1;
            continue;
        }
#endif

        if (ngx_strcmp(value[i].data, "bind") == 0) {
            ls->bind = 1;
            continue;
        }

#if (NGX_HAVE_TCP_FASTOPEN)
        if (ngx_strncmp(value[i].data, "fastopen=", 9) == 0) {
            ls->fastopen = ngx_atoi(value[i].data + 9, value[i].len - 9);
            ls->bind = 1;
            if (ls->fastopen == NGX_ERROR) {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                                   "invalid fastopen \"%V\"", &value[i]);
                return NGX_CONF_ERROR;
            }

            continue;
        }
#endif

        if (ngx_strncmp(value[i].data, "backlog=", 8) == 0) {
            ls->backlog = ngx_atoi(value[i].data + 8, value[i].len - 8);
            ls->bind = 1;

            if (ls->backlog == NGX_ERROR || ls->backlog == 0) {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                                   "invalid backlog \"%V\"", &value[i]);
                return NGX_CONF_ERROR;
            }

            backlog = 1;

            continue;
        }

        if (ngx_strncmp(value[i].data, "rcvbuf=", 7) == 0) {
            size.len = value[i].len - 7;
            size.data = value[i].data + 7;

            ls->rcvbuf = ngx_parse_size(&size);
            ls->bind = 1;

            if (ls->rcvbuf == NGX_ERROR) {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                                   "invalid rcvbuf \"%V\"", &value[i]);
                return NGX_CONF_ERROR;
            }

            continue;
        }

        if (ngx_strncmp(value[i].data, "sndbuf=", 7) == 0) {
            size.len = value[i].len - 7;
            size.data = value[i].data + 7;

            ls->sndbuf = ngx_parse_size(&size);
            ls->bind = 1;

            if (ls->sndbuf == NGX_ERROR) {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                                   "invalid sndbuf \"%V\"", &value[i]);
                return NGX_CONF_ERROR;
            }

            continue;
        }

        if (ngx_strcmp(value[i].data, "reuseport") == 0) {
#if (NGX_HAVE_REUSEPORT)
            ls->reuseport = 1;
            ls->bind = 1;
#else
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                               "reuseport is not supported "
                               "on this platform, ignored");
#endif
            continue;
        }

        if (ngx_strcmp(value[i].data, "ssl") == 0) {
#if (NGX_STREAM_SSL)
            ngx_stream_ssl_conf_t  *sslcf;
            sslcf = ngx_stream_conf_get_module_srv_conf(cf, ngx_stream_ssl_module);
            sslcf->listen = 1;
            sslcf->file = cf->conf_file->file.name.data;
            sslcf->line = cf->conf_file->line;
            ls->ssl = 1;
            continue;
#else
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                               "the \"ssl\" parameter requires "
                               "ngx_stream_ssl_module");
            return NGX_CONF_ERROR;
#endif
        }

        if (ngx_strncmp(value[i].data, "so_keepalive=", 13) == 0) {
            if (ngx_strcmp(&value[i].data[13], "on") == 0) {
                ls->so_keepalive = 1;
            } else if (ngx_strcmp(&value[i].data[13], "off") == 0) {
                ls->so_keepalive = 2;
            } else {

#if (NGX_HAVE_KEEPALIVE_TUNABLE)
                u_char     *p, *end;
                ngx_str_t   s;

                end = value[i].data + value[i].len;
                s.data = value[i].data + 13;

                p = ngx_strlchr(s.data, end, ':');
                if (p == NULL) {
                    p = end;
                }

                if (p > s.data) {
                    s.len = p - s.data;

                    ls->tcp_keepidle = ngx_parse_time(&s, 1);
                    if (ls->tcp_keepidle == (time_t) NGX_ERROR) {
                        goto invalid_so_keepalive;
                    }
                }

                s.data = (p < end) ? (p + 1) : end;

                p = ngx_strlchr(s.data, end, ':');
                if (p == NULL) {
                    p = end;
                }

                if (p > s.data) {
                    s.len = p - s.data;

                    ls->tcp_keepintvl = ngx_parse_time(&s, 1);
                    if (ls->tcp_keepintvl == (time_t) NGX_ERROR) {
                        goto invalid_so_keepalive;
                    }
                }

                s.data = (p < end) ? (p + 1) : end;

                if (s.data < end) {
                    s.len = end - s.data;

                    ls->tcp_keepcnt = ngx_atoi(s.data, s.len);
                    if (ls->tcp_keepcnt == NGX_ERROR) {
                        goto invalid_so_keepalive;
                    }
                }

                if (ls->tcp_keepidle == 0 && ls->tcp_keepintvl == 0
                    && ls->tcp_keepcnt == 0)
                {
                    goto invalid_so_keepalive;
                }

                ls->so_keepalive = 1;

#else
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                                   "the \"so_keepalive\" parameter accepts "
                                   "only \"on\" or \"off\" on this platform");
                return NGX_CONF_ERROR;
#endif
            }

            ls->bind = 1;
            continue;

#if (NGX_HAVE_KEEPALIVE_TUNABLE)
        invalid_so_keepalive:
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                               "invalid so_keepalive value: \"%s\"",
                               &value[i].data[13]);
            return NGX_CONF_ERROR;
#endif
        }

        if (ngx_strcmp(value[i].data, "proxy_protocol") == 0) {
            ls->proxy_protocol = 1;
            continue;
        }
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "the invalid \"%V\" parameter", &value[i]);
        return NGX_CONF_ERROR;
    }

    if (ls->type == SOCK_DGRAM) {
        if (backlog) {
            return "\"backlog\" parameter is incompatible with \"udp\"";
        }

#if (NGX_STREAM_SSL)
#if !(T_NGX_HAVE_DTLS)
        if (ls->ssl) {
            return "\"ssl\" parameter is incompatible with \"udp\"";
        }
#endif
#endif

        if (ls->so_keepalive) {
            return "\"so_keepalive\" parameter is incompatible with \"udp\"";
        }

        // if (ls->proxy_protocol) {
        //     return "\"proxy_protocol\" parameter is incompatible with \"udp\"";
        // }

#if (NGX_HAVE_TCP_FASTOPEN)
        if (ls->fastopen != -1) {
            return "\"fastopen\" parameter is incompatible with \"udp\"";
        }
#endif
    }

    for (n = 0; n < u.naddrs; n++) {
        for (i = 0; i < n; i++) {
            if (ngx_cmp_sockaddr(u.addrs[n].sockaddr, u.addrs[n].socklen, u.addrs[i].sockaddr, u.addrs[i].socklen, 1) == NGX_OK) {
                goto next;
            }
        }

        if (n != 0) {
            nls = ngx_array_push(&cmcf->listen);
            if (nls == NULL) {
                return NGX_CONF_ERROR;
            }
            *nls = *ls;
        } else {
            nls = ls;
        }

        nls->sockaddr = u.addrs[n].sockaddr; // 4001=0x0fa1
        nls->socklen = u.addrs[n].socklen;
        nls->addr_text = u.addrs[n].name; // 0.0.0.0:4001
        nls->wildcard = ngx_inet_wildcard(nls->sockaddr);
        als = cmcf->listen.elts;

#if (T_NGX_STREAM_SNI)
        continue;
#endif

        for (i = 0; i < cmcf->listen.nelts - 1; i++) {
            if (nls->type != als[i].type) {
                continue;
            }

            if (ngx_cmp_sockaddr(als[i].sockaddr, als[i].socklen, nls->sockaddr, nls->socklen, 1) != NGX_OK) {
                continue;
            }

            ngx_conf_log_error(NGX_LOG_ERR, cf, 0, "duplicate \"%V\" address and port pair", &nls->addr_text);
            return NGX_CONF_ERROR;
        }

    next:
        continue;
    }

    return NGX_CONF_OK;
}

static char * ngx_stream_core_error_log(ngx_conf_t *cf, ngx_command_t *cmd, void *conf) {
    ngx_stream_core_srv_conf_t  *cscf = conf;
    return ngx_log_set_log(cf, &cscf->error_log);
}
// ngx_stream_core_preread_phase()->ngx_stream_ssl_preread_handler()
ngx_int_t ngx_stream_core_preread_phase(ngx_stream_session_t *s, ngx_stream_phase_handler_t *ph) {
    size_t                       size;
    ssize_t                      n;
    ngx_int_t                    rc;
    ngx_connection_t            *c;
    ngx_stream_core_srv_conf_t  *cscf;

    c = s->connection;

    c->log->action = "prereading client data";

    cscf = ngx_stream_get_module_srv_conf(s, ngx_stream_core_module);

    if (c->read->timedout) {
        rc = NGX_STREAM_OK;

    } else if (c->read->timer_set) {
        rc = NGX_AGAIN;

    } else {
        rc = ph->handler(s); // (nginx`ngx_stream_ssl_preread_handler at ngx_stream_ssl_preread_module.c:220)
    }

    while (rc == NGX_AGAIN) {

        if (c->buffer == NULL) {
            c->buffer = ngx_create_temp_buf(c->pool, cscf->preread_buffer_size);
            if (c->buffer == NULL) {
                rc = NGX_ERROR;
                break;
            }
        }

        size = c->buffer->end - c->buffer->last;

        if (size == 0) {
            ngx_log_error(NGX_LOG_ERR, c->log, 0, "preread buffer full");
            rc = NGX_STREAM_BAD_REQUEST;
            break;
        }

        if (c->read->eof) {
            rc = NGX_STREAM_OK;
            break;
        }

        if (!c->read->ready) {
            break;
        }

        n = c->recv(c, c->buffer->last, size);

        if (n == NGX_ERROR || n == 0) {
            rc = NGX_STREAM_OK;
            break;
        }

        if (n == NGX_AGAIN) {
            break;
        }

        c->buffer->last += n;

        rc = ph->handler(s);
    }

    if (rc == NGX_AGAIN) {
        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            ngx_stream_finalize_session(s, NGX_STREAM_INTERNAL_SERVER_ERROR);
            return NGX_OK;
        }

        if (!c->read->timer_set) {
            ngx_add_timer(c->read, cscf->preread_timeout);
        }

        c->read->handler = ngx_stream_session_handler;

        return NGX_OK;
    }

    if (c->read->timer_set) {
        ngx_del_timer(c->read);
    }

    if (rc == NGX_OK) {
        s->phase_handler = ph->next;
        return NGX_AGAIN;
    }

    if (rc == NGX_DECLINED) {
        s->phase_handler++;
        return NGX_AGAIN;
    }

    if (rc == NGX_DONE) {
        return NGX_OK;
    }

    if (rc == NGX_ERROR) {
        rc = NGX_STREAM_INTERNAL_SERVER_ERROR;
    }

    ngx_stream_finalize_session(s, rc);

    return NGX_OK;
}

ngx_int_t ngx_stream_core_content_phase(ngx_stream_session_t *s, ngx_stream_phase_handler_t *ph) {
    ngx_connection_t            *c;
    ngx_stream_core_srv_conf_t  *cscf;

    c = s->connection;
    c->log->action = NULL;
    cscf = ngx_stream_get_module_srv_conf(s, ngx_stream_core_module);
    if (c->type == SOCK_STREAM
        && cscf->tcp_nodelay
        && ngx_tcp_nodelay(c) != NGX_OK) {
        ngx_stream_finalize_session(s, NGX_STREAM_INTERNAL_SERVER_ERROR);
        return NGX_OK;
    }

    cscf->handler(s); // ngx_stream_proxy_handler(),ngx_stream_return_handler()

    return NGX_OK;
}

ngx_int_t ngx_stream_core_generic_phase(ngx_stream_session_t *s, ngx_stream_phase_handler_t *ph) {
    ngx_int_t  rc;

    /*
     * generic phase checker,
     * used by all phases, except for preread and content
     */
    ngx_log_error(NGX_LOG_STDERR, s->connection->log, 0, "generic phase: %ui", s->phase_handler);
    /* [
        "NGX_STREAM_PREACCESS_PHASE": ngx_stream_set_handler(), 
        "NGX_STREAM_PREACCESS_PHASE": ngx_stream_limit_conn_handler(), 
        "NGX_STREAM_ACCESS_PHASE": ngx_stream_access_handler(), 
        "NGX_STREAM_SSL_PHASE": ngx_stream_ssl_handler(),
        
        ]
    */
    rc = ph->handler(s);
    if (rc == NGX_OK) {
        s->phase_handler = ph->next;
        return NGX_AGAIN;
    }

    if (rc == NGX_DECLINED) {
        s->phase_handler++;
        return NGX_AGAIN;
    }

    if (rc == NGX_AGAIN || rc == NGX_DONE) {
        return NGX_OK;
    }

    if (rc == NGX_ERROR) {
        rc = NGX_STREAM_INTERNAL_SERVER_ERROR;
    }

    ngx_stream_finalize_session(s, rc);

    return NGX_OK;
}

void ngx_stream_core_run_phases(ngx_stream_session_t *s) {
    ngx_int_t                     rc;
    ngx_stream_phase_handler_t   *ph;
    ngx_stream_core_main_conf_t  *cmcf;
    // 1. checker=ngx_stream_core_generic_phase,handler=ngx_stream_ssl_handler
    // 2. ngx_stream_core_preread_phase,ngx_stream_ssl_preread_handler
    // 3. ngx_stream_core_content_phase,ngx_stream_return_handler
    cmcf = ngx_stream_get_module_main_conf(s, ngx_stream_core_module);
    ph = cmcf->phase_engine.handlers;
    while (ph[s->phase_handler].checker) {
        rc = ph[s->phase_handler].checker(s, &ph[s->phase_handler]);
        if (rc == NGX_OK) {
            return;
        }
    }
}


typedef struct {
    ngx_chain_t  *from_upstream;
    ngx_chain_t  *from_downstream;
} ngx_stream_write_filter_ctx_t;
static ngx_stream_module_t  ngx_stream_write_filter_module_ctx = {
    NULL,                                  /* preconfiguration */
    ngx_stream_write_filter_init,          /* postconfiguration */

    NULL,                                  /* create main configuration */
    NULL,                                  /* init main configuration */

    NULL,                                  /* create server configuration */
    NULL                                   /* merge server configuration */
};
ngx_module_t  ngx_stream_write_filter_module = {
    NGX_MODULE_V1,
    &ngx_stream_write_filter_module_ctx,   /* module context */
    NULL,                                  /* module directives */
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
// 这里初始化 ngx_stream_top_filter，这个函数很重要, ngx_stream_return_module 和 ngx_stream_proxy_module 都会调用!!!
static ngx_int_t ngx_stream_write_filter_init(ngx_conf_t *cf) {
    ngx_stream_top_filter = ngx_stream_write_filter;
    return NGX_OK;
}
static ngx_int_t ngx_stream_write_filter(ngx_stream_session_t *s, ngx_chain_t *in, ngx_uint_t from_upstream) {
    off_t                           size;
    ngx_uint_t                      last, flush, sync;
    ngx_chain_t                    *cl, *ln, **ll, **out, *chain;
    ngx_connection_t               *c;
    ngx_stream_write_filter_ctx_t  *ctx;

    ctx = ngx_stream_get_module_ctx(s, ngx_stream_write_filter_module);
    if (ctx == NULL) {
        ctx = ngx_pcalloc(s->connection->pool, sizeof(ngx_stream_write_filter_ctx_t));
        if (ctx == NULL) {
            return NGX_ERROR;
        }

        ngx_stream_set_ctx(s, ctx, ngx_stream_write_filter_module);
    }

    if (from_upstream) {
        c = s->connection;
        out = &ctx->from_upstream;
    } else {
        c = s->upstream->peer.connection;
        out = &ctx->from_downstream;
    }

    if (c->error) {
        return NGX_ERROR;
    }

    size = 0;
    flush = 0;
    sync = 0;
    last = 0;
    ll = out;
    /* find the size, the flush point and the last link of the saved chain */
    for (cl = *out; cl; cl = cl->next) {
        ll = &cl->next;
        ngx_log_error(NGX_LOG_STDERR, c->log, 0, 
                       "write old buf t:%d f:%d %p, pos %p, size: %z "
                       "file: %O, size: %O",
                       cl->buf->temporary, cl->buf->in_file,
                       cl->buf->start, cl->buf->pos,
                       cl->buf->last - cl->buf->pos,
                       cl->buf->file_pos,
                       cl->buf->file_last - cl->buf->file_pos);

        if (ngx_buf_size(cl->buf) == 0 && !ngx_buf_special(cl->buf)) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                          "zero size buf in writer "
                          "t:%d r:%d f:%d %p %p-%p %p %O-%O",
                          cl->buf->temporary,
                          cl->buf->recycled,
                          cl->buf->in_file,
                          cl->buf->start,
                          cl->buf->pos,
                          cl->buf->last,
                          cl->buf->file,
                          cl->buf->file_pos,
                          cl->buf->file_last);

            ngx_debug_point();
            return NGX_ERROR;
        }

        if (ngx_buf_size(cl->buf) < 0) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                          "negative size buf in writer "
                          "t:%d r:%d f:%d %p %p-%p %p %O-%O",
                          cl->buf->temporary,
                          cl->buf->recycled,
                          cl->buf->in_file,
                          cl->buf->start,
                          cl->buf->pos,
                          cl->buf->last,
                          cl->buf->file,
                          cl->buf->file_pos,
                          cl->buf->file_last);

            ngx_debug_point();
            return NGX_ERROR;
        }

        size += ngx_buf_size(cl->buf);
        if (cl->buf->flush || cl->buf->recycled) {
            flush = 1;
        }
        if (cl->buf->sync) {
            sync = 1;
        }
        if (cl->buf->last_buf) {
            last = 1;
        }
    }

    /* add the new chain to the existent one */
    for (ln = in; ln; ln = ln->next) {
        cl = ngx_alloc_chain_link(c->pool);
        if (cl == NULL) {
            return NGX_ERROR;
        }

        cl->buf = ln->buf;
        *ll = cl;
        ll = &cl->next;
        ngx_log_error(NGX_LOG_STDERR, c->log, 0, 
                       "write new buf t:%d f:%d %p, pos %p, size: %z "
                       "file: %O, size: %O",
                       cl->buf->temporary, cl->buf->in_file,
                       cl->buf->start, cl->buf->pos,
                       cl->buf->last - cl->buf->pos,
                       cl->buf->file_pos,
                       cl->buf->file_last - cl->buf->file_pos);

        if (ngx_buf_size(cl->buf) == 0 && !ngx_buf_special(cl->buf)) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                          "zero size buf in writer "
                          "t:%d r:%d f:%d %p %p-%p %p %O-%O",
                          cl->buf->temporary,
                          cl->buf->recycled,
                          cl->buf->in_file,
                          cl->buf->start,
                          cl->buf->pos,
                          cl->buf->last,
                          cl->buf->file,
                          cl->buf->file_pos,
                          cl->buf->file_last);

            ngx_debug_point();
            return NGX_ERROR;
        }

        if (ngx_buf_size(cl->buf) < 0) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                          "negative size buf in writer "
                          "t:%d r:%d f:%d %p %p-%p %p %O-%O",
                          cl->buf->temporary,
                          cl->buf->recycled,
                          cl->buf->in_file,
                          cl->buf->start,
                          cl->buf->pos,
                          cl->buf->last,
                          cl->buf->file,
                          cl->buf->file_pos,
                          cl->buf->file_last);

            ngx_debug_point();
            return NGX_ERROR;
        }

        size += ngx_buf_size(cl->buf);
        if (cl->buf->flush || cl->buf->recycled) {
            flush = 1;
        }
        if (cl->buf->sync) {
            sync = 1;
        }
        if (cl->buf->last_buf) {
            last = 1;
        }
    }

    *ll = NULL;
    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "stream write filter: l:%ui f:%ui s:%O", last, flush, size);
    if (size == 0
        && !(c->buffered & NGX_LOWLEVEL_BUFFERED)
        && !(last && c->need_last_buf)
        && !(flush && c->need_flush_buf)) {
        if (last || flush || sync) {
            for (cl = *out; cl; /* void */) {
                ln = cl;
                cl = cl->next;
                ngx_free_chain(c->pool, ln);
            }

            *out = NULL;
            c->buffered &= ~NGX_STREAM_WRITE_BUFFERED;

            return NGX_OK;
        }

        ngx_log_error(NGX_LOG_STDERR, c->log, 0, "the stream output chain is empty");
        ngx_debug_point();
        return NGX_ERROR;
    }
    // udp: -> ngx_udp_unix_sendmsg_chain() -> 很重要, ngx_sendmsg()
    // send_chain -> ngx_darwin_sendfile_chain() -> 很重要，在这里会返回客户端 "hello world", writev(connection_fd, text)
    chain = c->send_chain(c, *out, 0);
    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "stream write filter %p", chain);
    if (chain == NGX_CHAIN_ERROR) {
        c->error = 1;
        return NGX_ERROR;
    }

    for (cl = *out; cl && cl != chain; /* void */) {
        ln = cl;
        cl = cl->next;
        ngx_free_chain(c->pool, ln);
    }

    *out = chain;
    if (chain) {
        if (c->shared) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0, "shared connection is busy");
            return NGX_ERROR;
        }

        c->buffered |= NGX_STREAM_WRITE_BUFFERED;
        return NGX_AGAIN;
    }

    c->buffered &= ~NGX_STREAM_WRITE_BUFFERED;
    if (c->buffered & NGX_LOWLEVEL_BUFFERED) {
        return NGX_AGAIN;
    }

    return NGX_OK;
}

