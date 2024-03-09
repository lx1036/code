



#include <ngx_config.h>
#include <ngx_core.h>


static ngx_int_t ngx_parse_inet_url(ngx_pool_t *pool, ngx_url_t *u);
static ngx_int_t ngx_inet_add_addr(ngx_pool_t *pool, ngx_url_t *u, struct sockaddr *sockaddr, socklen_t socklen, ngx_uint_t total);

// 获取地址的字符串形式 sa -> u_char *text
size_t ngx_sock_ntop(struct sockaddr *sa, socklen_t socklen, u_char *text, size_t len, ngx_uint_t port) {
    u_char               *p;
    struct sockaddr_in   *sin;

    switch (sa->sa_family) {
    case AF_INET:
        sin = (struct sockaddr_in *) sa;
        p = (u_char *) &sin->sin_addr;
        if (port) {
            p = ngx_snprintf(text, len, "%ud.%ud.%ud.%ud:%d", p[0], p[1], p[2], p[3], ntohs(sin->sin_port));
        } else {
            p = ngx_snprintf(text, len, "%ud.%ud.%ud.%ud", p[0], p[1], p[2], p[3]);
        }

        return (p - text);

    default:
        return 0;
    }
}

ngx_int_t ngx_cmp_sockaddr(struct sockaddr *sa1, socklen_t slen1,
    struct sockaddr *sa2, socklen_t slen2, ngx_uint_t cmp_port) {
    struct sockaddr_in   *sin1, *sin2;
    if (sa1->sa_family != sa2->sa_family) {
        return NGX_DECLINED;
    }

    switch (sa1->sa_family) {
    default: /* AF_INET */
        sin1 = (struct sockaddr_in *) sa1;
        sin2 = (struct sockaddr_in *) sa2;
        if (cmp_port && sin1->sin_port != sin2->sin_port) {
            return NGX_DECLINED;
        }
        if (sin1->sin_addr.s_addr != sin2->sin_addr.s_addr) {
            return NGX_DECLINED;
        }
        break;
    }

    return NGX_OK;
}


ngx_int_t
ngx_ptocidr(ngx_str_t *text, ngx_cidr_t *cidr)
{
    u_char      *addr, *mask, *last;
    size_t       len;
    ngx_int_t    shift;
    addr = text->data;
    last = addr + text->len;
    mask = ngx_strlchr(addr, last, '/');
    len = (mask ? mask : last) - addr;
    cidr->u.in.addr = ngx_inet_addr(addr, len);
    if (cidr->u.in.addr != INADDR_NONE) {
        cidr->family = AF_INET;
        if (mask == NULL) {
            cidr->u.in.mask = 0xffffffff;
            return NGX_OK;
        }
    } else {
        return NGX_ERROR;
    }

    mask++;
    shift = ngx_atoi(mask, last - mask);
    if (shift == NGX_ERROR) {
        return NGX_ERROR;
    }

    switch (cidr->family) {
    default: /* AF_INET */
        if (shift > 32) {
            return NGX_ERROR;
        }

        if (shift) {
            cidr->u.in.mask = htonl((uint32_t) (0xffffffffu << (32 - shift)));

        } else {
            /* x86 compilers use a shl instruction that shifts by modulo 32 */
            cidr->u.in.mask = 0;
        }

        if (cidr->u.in.addr == (cidr->u.in.addr & cidr->u.in.mask)) {
            return NGX_OK;
        }

        cidr->u.in.addr &= cidr->u.in.mask;

        return NGX_DONE;
    }
}

ngx_int_t
ngx_inet_resolve_host(ngx_pool_t *pool, ngx_url_t *u)
{
    u_char              *host;
    ngx_uint_t           i, n;
    struct hostent      *h;
    struct sockaddr_in   sin;

    /* AF_INET only */
    ngx_memzero(&sin, sizeof(struct sockaddr_in));
    sin.sin_family = AF_INET;
    sin.sin_addr.s_addr = ngx_inet_addr(u->host.data, u->host.len);
    if (sin.sin_addr.s_addr == INADDR_NONE) {
        host = ngx_alloc(u->host.len + 1, pool->log);
        if (host == NULL) {
            return NGX_ERROR;
        }

        (void) ngx_cpystrn(host, u->host.data, u->host.len + 1);
        h = gethostbyname((char *) host);
        ngx_free(host);

        if (h == NULL || h->h_addr_list[0] == NULL) {
            u->err = "host not found";
            return NGX_ERROR;
        }

        for (n = 0; h->h_addr_list[n] != NULL; n++) { /* void */ }

        /* MP: ngx_shared_palloc() */

        for (i = 0; i < n; i++) {
            sin.sin_addr.s_addr = *(in_addr_t *) (h->h_addr_list[i]);

            if (ngx_inet_add_addr(pool, u, (struct sockaddr *) &sin,
                                  sizeof(struct sockaddr_in), n)
                != NGX_OK)
            {
                return NGX_ERROR;
            }
        }
    } else {
        /* MP: ngx_shared_palloc() */
        if (ngx_inet_add_addr(pool, u, (struct sockaddr *) &sin,
                              sizeof(struct sockaddr_in), 1)
            != NGX_OK)
        {
            return NGX_ERROR;
        }
    }

    return NGX_OK;
}

static ngx_int_t ngx_inet_add_addr(ngx_pool_t *pool, ngx_url_t *u, struct sockaddr *sockaddr,
    socklen_t socklen, ngx_uint_t total) {
    u_char           *p;
    size_t            len;
    ngx_uint_t        i, nports;
    ngx_addr_t       *addr;
    struct sockaddr  *sa;

    nports = u->last_port ? u->last_port - u->port + 1 : 1;
    if (u->addrs == NULL) {
        u->addrs = ngx_palloc(pool, total * nports * sizeof(ngx_addr_t));
        if (u->addrs == NULL) {
            return NGX_ERROR;
        }
    }

    for (i = 0; i < nports; i++) {
        sa = ngx_pcalloc(pool, socklen);
        if (sa == NULL) {
            return NGX_ERROR;
        }

        ngx_memcpy(sa, sockaddr, socklen);
        ngx_inet_set_port(sa, u->port + i);
        switch (sa->sa_family) {
        default: /* AF_INET */
            len = NGX_INET_ADDRSTRLEN + sizeof(":65535") - 1;
        }

        p = ngx_pnalloc(pool, len);
        if (p == NULL) {
            return NGX_ERROR;
        }

        len = ngx_sock_ntop(sa, socklen, p, len, 1);

        addr = &u->addrs[u->naddrs++];

        addr->sockaddr = sa;
        addr->socklen = socklen;

        addr->name.len = len;
        addr->name.data = p;
    }

    return NGX_OK;
}

void ngx_inet_set_port(struct sockaddr *sa, in_port_t port) {
    struct sockaddr_in   *sin;
    switch (sa->sa_family) {
    default: /* AF_INET */
        sin = (struct sockaddr_in *) sa;
        sin->sin_port = htons(port);
        break;
    }
}


ngx_int_t ngx_parse_url(ngx_pool_t *pool, ngx_url_t *u) {
    u_char  *p;
    size_t   len;

    p = u->url.data;
    len = u->url.len;

    return ngx_parse_inet_url(pool, u);
}

ngx_uint_t ngx_inet_wildcard(struct sockaddr *sa) {
    struct sockaddr_in   *sin;
    switch (sa->sa_family) {
    case AF_INET:
        sin = (struct sockaddr_in *) sa;
        if (sin->sin_addr.s_addr == INADDR_ANY) {
            return 1;
        }
        break;
    }

    return 0;
}

in_port_t ngx_inet_get_port(struct sockaddr *sa) {
    struct sockaddr_in   *sin;
    switch (sa->sa_family) {
    default: /* AF_INET */
        sin = (struct sockaddr_in *) sa;
        return ntohs(sin->sin_port);
    }
}
// c 语言和 go 还不一样，go 在函数里定义变量然后返回(变量在 stack memory)，但是 c 会报错
// "报错: address of stack memory associated with local variable 'text' returned"
// 所以，c 需要在函数变量，形参指针传进来，也在栈内存.
// @see https://blog.csdn.net/aiwr_/article/details/110431441 栈内存(形参和局部变量)/堆内存(malloc/calloc/free的内存)/静态内存(static变量等)/常量内存(const)/代码存储内存, 总共5块内存区域
// @see ngx_event_accept.c 里 ngx_inet_get_addr() 调用
u_char * ngx_inet_get_addr(struct sockaddr *sa, u_char *text) {
    struct sockaddr_in   *sin;
    u_char               *p;
    // 报错: address of stack memory associated with local variable 'text' returned
    // u_char            text[NGX_SOCKADDR_STRLEN];
    switch (sa->sa_family) {
    default: /* AF_INET */
        sin = (struct sockaddr_in *) sa;
        p = (u_char *) &sin->sin_addr;
        p = ngx_snprintf(text, NGX_SOCKADDR_STRLEN, "%ud.%ud.%ud.%ud", p[0], p[1], p[2], p[3]);
        return p;
    }
}

// "127.0.0.1" -> uint32
in_addr_t ngx_inet_addr(u_char *text, size_t len) {
    u_char      *p, c;
    in_addr_t    addr;
    ngx_uint_t   octet, n;

    addr = 0;
    octet = 0;
    n = 0;
    for (p = text; p < text + len; p++) {
        c = *p;
        if (c >= '0' && c <= '9') {
            octet = octet * 10 + (c - '0');
            if (octet > 255) {
                return INADDR_NONE;
            }

            continue;
        }

        if (c == '.') {
            addr = (addr << 8) + octet;
            octet = 0;
            n++;
            continue;
        }

        return INADDR_NONE;
    }

    if (n == 3) {
        addr = (addr << 8) + octet;
        return htonl(addr);
    }

    return INADDR_NONE;
}
// "127.0.0.1" -> uint32
static ngx_int_t ngx_parse_inet_url(ngx_pool_t *pool, ngx_url_t *u) {
    u_char              *host, *port, *last, *uri, *args, *dash;
    size_t               len;
    ngx_int_t            n;
    struct sockaddr_in  *sin;
    u->socklen = sizeof(struct sockaddr_in);
    sin = (struct sockaddr_in *) &u->sockaddr;
    sin->sin_family = AF_INET;
    u->family = AF_INET;
    host = u->url.data;
    last = host + u->url.len;
    port = ngx_strlchr(host, last, ':');
    uri = ngx_strlchr(host, last, '/');
    args = ngx_strlchr(host, last, '?');
    if (args) {
        if (uri == NULL || args < uri) {
            uri = args;
        }
    }

    if (uri) {
        if (u->listen || !u->uri_part) {
            u->err = "invalid host";
            return NGX_ERROR;
        }

        u->uri.len = last - uri;
        u->uri.data = uri;

        last = uri;

        if (uri < port) {
            port = NULL;
        }
    }

    if (port) {
        port++;
        len = last - port;
        if (u->listen) {
            dash = ngx_strlchr(port, last, '-');
            if (dash) {
                dash++;

                n = ngx_atoi(dash, last - dash);

                if (n < 1 || n > 65535) {
                    u->err = "invalid port";
                    return NGX_ERROR;
                }

                u->last_port = (in_port_t) n;

                len = dash - port - 1;
            }
        }

        n = ngx_atoi(port, len);

        if (n < 1 || n > 65535) {
            u->err = "invalid port";
            return NGX_ERROR;
        }

        if (u->last_port && n > u->last_port) {
            u->err = "invalid port range";
            return NGX_ERROR;
        }

        u->port = (in_port_t) n;
        sin->sin_port = htons((in_port_t) n);
        u->port_text.len = last - port;
        u->port_text.data = port;
        last = port - 1;
    } else {
        if (uri == NULL) {
            if (u->listen) {

                /* test value as port only */

                len = last - host;

                dash = ngx_strlchr(host, last, '-');

                if (dash) {
                    dash++;

                    n = ngx_atoi(dash, last - dash);

                    if (n == NGX_ERROR) {
                        goto no_port;
                    }

                    if (n < 1 || n > 65535) {
                        u->err = "invalid port";

                    } else {
                        u->last_port = (in_port_t) n;
                    }

                    len = dash - host - 1;
                }

                n = ngx_atoi(host, len);

                if (n != NGX_ERROR) {

                    if (u->err) {
                        return NGX_ERROR;
                    }

                    if (n < 1 || n > 65535) {
                        u->err = "invalid port";
                        return NGX_ERROR;
                    }

                    if (u->last_port && n > u->last_port) {
                        u->err = "invalid port range";
                        return NGX_ERROR;
                    }

                    u->port = (in_port_t) n;
                    sin->sin_port = htons((in_port_t) n);
                    sin->sin_addr.s_addr = INADDR_ANY;

                    u->port_text.len = last - host;
                    u->port_text.data = host;

                    u->wildcard = 1;

                    return ngx_inet_add_addr(pool, u, &u->sockaddr.sockaddr,
                                             u->socklen, 1);
                }
            }
        }

no_port:

        u->err = NULL;
        u->no_port = 1;
        u->port = u->default_port;
        sin->sin_port = htons(u->default_port);
        u->last_port = 0;
    }

    len = last - host;

    if (len == 0) {
        u->err = "no host";
        return NGX_ERROR;
    }

    u->host.len = len;
    u->host.data = host;

    if (u->listen && len == 1 && *host == '*') {
        sin->sin_addr.s_addr = INADDR_ANY;
        u->wildcard = 1;
        return ngx_inet_add_addr(pool, u, &u->sockaddr.sockaddr, u->socklen, 1);
    }

    sin->sin_addr.s_addr = ngx_inet_addr(host, len);
    if (sin->sin_addr.s_addr != INADDR_NONE) {
        if (sin->sin_addr.s_addr == INADDR_ANY) {
            u->wildcard = 1;
        }

        return ngx_inet_add_addr(pool, u, &u->sockaddr.sockaddr, u->socklen, 1);
    }

    if (u->no_resolve) {
        return NGX_OK;
    }

    if (ngx_inet_resolve_host(pool, u) != NGX_OK) {
        return NGX_ERROR;
    }

    u->family = u->addrs[0].sockaddr->sa_family;
    u->socklen = u->addrs[0].socklen;
    ngx_memcpy(&u->sockaddr, u->addrs[0].sockaddr, u->addrs[0].socklen);
    u->wildcard = ngx_inet_wildcard(&u->sockaddr.sockaddr);

    return NGX_OK;
}

ngx_int_t
ngx_parse_addr_port(ngx_pool_t *pool, ngx_addr_t *addr, u_char *text,
    size_t len)
{
    u_char     *p, *last;
    size_t      plen;
    ngx_int_t   rc, port;

    rc = ngx_parse_addr(pool, addr, text, len);

    if (rc != NGX_DECLINED) {
        return rc;
    }

    last = text + len;

    {
        p = ngx_strlchr(text, last, ':');

        if (p == NULL) {
            return NGX_DECLINED;
        }
    }

    p++;
    plen = last - p;

    port = ngx_atoi(p, plen);

    if (port < 1 || port > 65535) {
        return NGX_DECLINED;
    }

    len -= plen + 1;

    rc = ngx_parse_addr(pool, addr, text, len);

    if (rc != NGX_OK) {
        return rc;
    }

    ngx_inet_set_port(addr->sockaddr, (in_port_t) port);

    return NGX_OK;
}

ngx_int_t ngx_parse_addr(ngx_pool_t *pool, ngx_addr_t *addr, u_char *text, size_t len) {
    in_addr_t             inaddr;
    ngx_uint_t            family;
    struct sockaddr_in   *sin;
    inaddr = ngx_inet_addr(text, len);

    if (inaddr != INADDR_NONE) {
        family = AF_INET;
        len = sizeof(struct sockaddr_in);
    } else {
        return NGX_DECLINED;
    }

    addr->sockaddr = ngx_pcalloc(pool, len);
    if (addr->sockaddr == NULL) {
        return NGX_ERROR;
    }

    addr->sockaddr->sa_family = (u_char) family;
    addr->socklen = len;

    switch (family) {

    default: /* AF_INET */
        sin = (struct sockaddr_in *) addr->sockaddr;
        sin->sin_addr.s_addr = inaddr;
        break;
    }

    return NGX_OK;
}



