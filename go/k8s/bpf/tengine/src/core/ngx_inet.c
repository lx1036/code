



#include <ngx_config.h>
#include <ngx_core.h>


static ngx_int_t ngx_parse_unix_domain_url(ngx_pool_t *pool, ngx_url_t *u);
static ngx_int_t ngx_parse_inet_url(ngx_pool_t *pool, ngx_url_t *u);
static ngx_int_t ngx_parse_inet6_url(ngx_pool_t *pool, ngx_url_t *u);
static ngx_int_t ngx_inet_add_addr(ngx_pool_t *pool, ngx_url_t *u,
    struct sockaddr *sockaddr, socklen_t socklen, ngx_uint_t total);





size_t
ngx_sock_ntop(struct sockaddr *sa, socklen_t socklen, u_char *text, size_t len,
    ngx_uint_t port)
{
    u_char               *p;
#if (NGX_HAVE_INET6 || NGX_HAVE_UNIX_DOMAIN)
    size_t                n;
#endif
    struct sockaddr_in   *sin;
#if (NGX_HAVE_UNIX_DOMAIN)
    struct sockaddr_un   *saun;
#endif

    switch (sa->sa_family) {
    case AF_INET:
        sin = (struct sockaddr_in *) sa;
        p = (u_char *) &sin->sin_addr;

        if (port) {
            p = ngx_snprintf(text, len, "%ud.%ud.%ud.%ud:%d",
                             p[0], p[1], p[2], p[3], ntohs(sin->sin_port));
        } else {
            p = ngx_snprintf(text, len, "%ud.%ud.%ud.%ud",
                             p[0], p[1], p[2], p[3]);
        }

        return (p - text);


#if (NGX_HAVE_UNIX_DOMAIN)

    case AF_UNIX:
        saun = (struct sockaddr_un *) sa;

        /* on Linux sockaddr might not include sun_path at all */

        if (socklen <= (socklen_t) offsetof(struct sockaddr_un, sun_path)) {
            p = ngx_snprintf(text, len, "unix:%Z");

        } else {
            n = ngx_strnlen((u_char *) saun->sun_path,
                            socklen - offsetof(struct sockaddr_un, sun_path));
            p = ngx_snprintf(text, len, "unix:%*s%Z", n, saun->sun_path);
        }

        /* we do not include trailing zero in address length */

        return (p - text - 1);

#endif

    default:
        return 0;
    }
}
