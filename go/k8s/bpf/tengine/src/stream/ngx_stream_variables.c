



#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_stream.h>
#include <nginx.h>

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

static ngx_int_t ngx_stream_variable_nginx_version(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_hostname(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_pid(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_msec(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_time_iso8601(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_time_local(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);
static ngx_int_t ngx_stream_variable_protocol(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data);

static ngx_uint_t  ngx_stream_variable_depth = 100;


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

    // { ngx_string("nginx_version"), NULL, ngx_stream_variable_nginx_version,
    //   0, 0, 0 },

    { ngx_string("hostname"), NULL, ngx_stream_variable_hostname,
      0, 0, 0 },

    { ngx_string("pid"), NULL, ngx_stream_variable_pid,
      0, 0, 0 },

    { ngx_string("msec"), NULL, ngx_stream_variable_msec,
      0, NGX_STREAM_VAR_NOCACHEABLE, 0 },

    // { ngx_string("time_iso8601"), NULL, ngx_stream_variable_time_iso8601,
    //   0, NGX_STREAM_VAR_NOCACHEABLE, 0 },

    // { ngx_string("time_local"), NULL, ngx_stream_variable_time_local,
    //   0, NGX_STREAM_VAR_NOCACHEABLE, 0 },

    { ngx_string("protocol"), NULL,
      ngx_stream_variable_protocol, 0, 0, 0 },

      ngx_stream_null_variable
};

ngx_int_t ngx_stream_variables_add_core_vars(ngx_conf_t *cf) {
    ngx_stream_variable_t        *cv, *v;
    ngx_stream_core_main_conf_t  *cmcf;

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);

    cmcf->variables_keys = ngx_pcalloc(cf->temp_pool,
                                       sizeof(ngx_hash_keys_arrays_t));
    if (cmcf->variables_keys == NULL) {
        return NGX_ERROR;
    }

    cmcf->variables_keys->pool = cf->pool;
    cmcf->variables_keys->temp_pool = cf->pool;

    if (ngx_hash_keys_array_init(cmcf->variables_keys, NGX_HASH_SMALL)
        != NGX_OK)
    {
        return NGX_ERROR;
    }

    if (ngx_array_init(&cmcf->prefix_variables, cf->pool, 8,
                       sizeof(ngx_stream_variable_t))
        != NGX_OK)
    {
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
        if (name->len != key[i].key.len
            || ngx_strncasecmp(name->data, key[i].key.data, name->len) != 0)
        {
            continue;
        }

        v = key[i].value;

        if (!(v->flags & NGX_STREAM_VAR_CHANGEABLE)) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                               "the duplicate \"%V\" variable", name);
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
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                           "conflicting variable name \"%V\"", name);
        return NULL;
    }

    return v;
}

static ngx_stream_variable_t * ngx_stream_add_prefix_variable(ngx_conf_t *cf, ngx_str_t *name, ngx_uint_t flags) {
    ngx_uint_t                    i;
    ngx_stream_variable_t        *v;
    ngx_stream_core_main_conf_t  *cmcf;

    cmcf = ngx_stream_conf_get_module_main_conf(cf, ngx_stream_core_module);
    v = cmcf->prefix_variables.elts;
    for (i = 0; i < cmcf->prefix_variables.nelts; i++) {
        if (name->len != v[i].name.len
            || ngx_strncasecmp(name->data, v[i].name.data, name->len) != 0) {
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


static ngx_int_t ngx_stream_variable_binary_remote_addr(ngx_stream_session_t *s,
     ngx_stream_variable_value_t *v, uintptr_t data) {
    struct sockaddr_in   *sin;
    switch (s->connection->sockaddr->sa_family) {
#if (NGX_HAVE_UNIX_DOMAIN)
    case AF_UNIX:
        v->len = s->connection->addr_text.len;
        v->valid = 1;
        v->no_cacheable = 0;
        v->not_found = 0;
        v->data = s->connection->addr_text.data;

        break;
#endif

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

static ngx_int_t
ngx_stream_variable_proxy_protocol_addr(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data)
{
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


static ngx_int_t
ngx_stream_variable_proxy_protocol_port(ngx_stream_session_t *s,
    ngx_stream_variable_value_t *v, uintptr_t data)
{
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
            ngx_log_error(NGX_LOG_EMERG, cf->log, 0, "unknown \"%V\" variable", &v[i].name);
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
    if (ngx_hash_init(&hash, cmcf->variables_keys->keys.elts,
                      cmcf->variables_keys->keys.nelts) != NGX_OK) {
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
        ngx_log_error(NGX_LOG_ALERT, s->connection->log, 0,
                      "unknown variable index: %ui", index);
        return NULL;
    }

    if (s->variables[index].not_found || s->variables[index].valid) {
        return &s->variables[index];
    }

    v = cmcf->variables.elts;

    if (ngx_stream_variable_depth == 0) {
        ngx_log_error(NGX_LOG_ERR, s->connection->log, 0,
                      "cycle while evaluating variable \"%V\"",
                      &v[index].name);
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

size_t
ngx_stream_script_copy_var_len_code(ngx_stream_script_engine_t *e)
{
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


void
ngx_stream_script_copy_var_code(ngx_stream_script_engine_t *e)
{
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

            ngx_log_debug2(NGX_LOG_DEBUG_STREAM,
                           e->session->connection->log, 0,
                           "stream script var: \"%*s\"", e->pos - p, p);
        }
    }
}







