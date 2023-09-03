#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_stream.h>


static ngx_int_t ngx_stream_upstream_add_variables(ngx_conf_t *cf);
static ngx_int_t ngx_stream_upstream_addr_variable(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_upstream_response_time_variable(
    ngx_stream_session_t *s, ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_upstream_bytes_variable(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);

static char *ngx_stream_upstream(ngx_conf_t *cf, ngx_command_t *cmd,
    void *dummy);
static char *ngx_stream_upstream_server(ngx_conf_t *cf, ngx_command_t *cmd,
    void *conf);
static void *ngx_stream_upstream_create_main_conf(ngx_conf_t *cf);
static char *ngx_stream_upstream_init_main_conf(ngx_conf_t *cf, void *conf);


static ngx_command_t  ngx_stream_upstream_commands[] = {

    // { ngx_string("upstream"),
    //   NGX_STREAM_MAIN_CONF|NGX_CONF_BLOCK|NGX_CONF_TAKE1,
    //   ngx_stream_upstream,
    //   0,
    //   0,
    //   NULL },

    // { ngx_string("server"),
    //   NGX_STREAM_UPS_CONF|NGX_CONF_1MORE,
    //   ngx_stream_upstream_server,
    //   NGX_STREAM_SRV_CONF_OFFSET,
    //   0,
    //   NULL },

      ngx_null_command
};
static ngx_stream_module_t  ngx_stream_upstream_module_ctx = {
    ngx_stream_upstream_add_variables,     /* preconfiguration */
    NULL,                                  /* postconfiguration */

    ngx_stream_upstream_create_main_conf,  /* create main configuration */
    ngx_stream_upstream_init_main_conf,    /* init main configuration */

    NULL,                                  /* create server configuration */
    NULL                                   /* merge server configuration */
};
ngx_module_t  ngx_stream_upstream_module = {
    NGX_MODULE_V1,
    &ngx_stream_upstream_module_ctx,       /* module context */
    ngx_stream_upstream_commands,          /* module directives */
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
static ngx_stream_variable_t  ngx_stream_upstream_vars[] = {

    // { ngx_string("upstream_addr"), NULL,
    //   ngx_stream_upstream_addr_variable, 0,
    //   NGX_STREAM_VAR_NOCACHEABLE, 0 },

    // { ngx_string("upstream_bytes_sent"), NULL,
    //   ngx_stream_upstream_bytes_variable, 0,
    //   NGX_STREAM_VAR_NOCACHEABLE, 0 },

    // { ngx_string("upstream_connect_time"), NULL,
    //   ngx_stream_upstream_response_time_variable, 2,
    //   NGX_STREAM_VAR_NOCACHEABLE, 0 },

    // { ngx_string("upstream_first_byte_time"), NULL,
    //   ngx_stream_upstream_response_time_variable, 1,
    //   NGX_STREAM_VAR_NOCACHEABLE, 0 },

    // { ngx_string("upstream_session_time"), NULL,
    //   ngx_stream_upstream_response_time_variable, 0,
    //   NGX_STREAM_VAR_NOCACHEABLE, 0 },

    // { ngx_string("upstream_bytes_received"), NULL,
    //   ngx_stream_upstream_bytes_variable, 1,
    //   NGX_STREAM_VAR_NOCACHEABLE, 0 },

      ngx_stream_null_variable
};

ngx_stream_upstream_srv_conf_t * ngx_stream_upstream_add(ngx_conf_t *cf, ngx_url_t *u, ngx_uint_t flags) {
    ngx_uint_t                        i;
    ngx_stream_upstream_server_t     *us;
    ngx_stream_upstream_srv_conf_t   *uscf, **uscfp;
    ngx_stream_upstream_main_conf_t  *umcf;

    if (!(flags & NGX_STREAM_UPSTREAM_CREATE)) {
        if (ngx_parse_url(cf->pool, u) != NGX_OK) {
            if (u->err) {
                ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "%s in upstream \"%V\"", u->err, &u->url);
            }
            return NULL;
        }
    }

    umcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_upstream_module);
    uscfp = umcf->upstreams.elts;
    for (i = 0; i < umcf->upstreams.nelts; i++) {
        if (uscfp[i]->host.len != u->host.len || ngx_strncasecmp(uscfp[i]->host.data, u->host.data, u->host.len) != 0) {
            continue;
        }

        if ((flags & NGX_STREAM_UPSTREAM_CREATE)
             && (uscfp[i]->flags & NGX_STREAM_UPSTREAM_CREATE)) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "duplicate upstream \"%V\"", &u->host);
            return NULL;
        }

        if ((uscfp[i]->flags & NGX_STREAM_UPSTREAM_CREATE) && !u->no_port) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0, "upstream \"%V\" may not have port %d", &u->host, u->port);
            return NULL;
        }

        if ((flags & NGX_STREAM_UPSTREAM_CREATE) && !uscfp[i]->no_port) {
            ngx_log_error(NGX_LOG_EMERG, cf->log, 0, "upstream \"%V\" may not have port %d in %s:%ui",
                          &u->host, uscfp[i]->port,
                          uscfp[i]->file_name, uscfp[i]->line);
            return NULL;
        }

        if (uscfp[i]->port != u->port) {
            continue;
        }

        if (flags & NGX_STREAM_UPSTREAM_CREATE) {
            uscfp[i]->flags = flags;
        }

        return uscfp[i];
    }

    uscf = ngx_pcalloc(cf->pool, sizeof(ngx_stream_upstream_srv_conf_t));
    if (uscf == NULL) {
        return NULL;
    }

    uscf->flags = flags;
    uscf->host = u->host;
    uscf->file_name = cf->conf_file->file.name.data;
    uscf->line = cf->conf_file->line;
    uscf->port = u->port;
    uscf->no_port = u->no_port;
    if (u->naddrs == 1 && (u->port || u->family == AF_UNIX)) {
        uscf->servers = ngx_array_create(cf->pool, 1, sizeof(ngx_stream_upstream_server_t));
        if (uscf->servers == NULL) {
            return NULL;
        }

        us = ngx_array_push(uscf->servers);
        if (us == NULL) {
            return NULL;
        }

        ngx_memzero(us, sizeof(ngx_stream_upstream_server_t));
        us->addrs = u->addrs;
        us->naddrs = 1;
    }

    uscfp = ngx_array_push(&umcf->upstreams);
    if (uscfp == NULL) {
        return NULL;
    }
    *uscfp = uscf;

    return uscf;
}

static ngx_int_t
ngx_stream_upstream_add_variables(ngx_conf_t *cf)
{
    ngx_stream_variable_t  *var, *v;

    for (v = ngx_stream_upstream_vars; v->name.len; v++) {
        var = ngx_stream_add_variable(cf, &v->name, v->flags);
        if (var == NULL) {
            return NGX_ERROR;
        }

        var->get_handler = v->get_handler;
        var->data = v->data;
    }

    return NGX_OK;
}


static void *
ngx_stream_upstream_create_main_conf(ngx_conf_t *cf)
{
    ngx_stream_upstream_main_conf_t  *umcf;

    umcf = ngx_pcalloc(cf->pool, sizeof(ngx_stream_upstream_main_conf_t));
    if (umcf == NULL) {
        return NULL;
    }

    if (ngx_array_init(&umcf->upstreams, cf->pool, 4,
                       sizeof(ngx_stream_upstream_srv_conf_t *))
        != NGX_OK)
    {
        return NULL;
    }

    return umcf;
}


static char *
ngx_stream_upstream_init_main_conf(ngx_conf_t *cf, void *conf)
{
    ngx_stream_upstream_main_conf_t *umcf = conf;

    ngx_uint_t                        i;
    ngx_stream_upstream_init_pt       init;
    ngx_stream_upstream_srv_conf_t  **uscfp;

    uscfp = umcf->upstreams.elts;

    for (i = 0; i < umcf->upstreams.nelts; i++) {

        init = uscfp[i]->peer.init_upstream
                                         ? uscfp[i]->peer.init_upstream
                                         : ngx_stream_upstream_init_round_robin;

        if (init(cf, uscfp[i]) != NGX_OK) {
            return NGX_CONF_ERROR;
        }
    }

    return NGX_CONF_OK;
}

