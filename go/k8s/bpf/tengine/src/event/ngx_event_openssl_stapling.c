


#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_event.h>
#include <ngx_event_connect.h>



#if (!defined OPENSSL_NO_OCSP && defined SSL_CTRL_SET_TLSEXT_STATUS_REQ_CB)


typedef struct {
    ngx_str_t                    staple;
    ngx_msec_t                   timeout;

    ngx_resolver_t              *resolver;
    ngx_msec_t                   resolver_timeout;

    ngx_addr_t                  *addrs;
    ngx_uint_t                   naddrs;
    ngx_str_t                    host;
    ngx_str_t                    uri;
    in_port_t                    port;

    SSL_CTX                     *ssl_ctx;

    X509                        *cert;
    X509                        *issuer;
    STACK_OF(X509)              *chain;

    u_char                      *name;

    time_t                       valid;
    time_t                       refresh;

    unsigned                     verify:1;
    unsigned                     loading:1;
} ngx_ssl_stapling_t;


typedef struct {
    ngx_addr_t                  *addrs;
    ngx_uint_t                   naddrs;

    ngx_str_t                    host;
    ngx_str_t                    uri;
    in_port_t                    port;
    ngx_uint_t                   depth;

    ngx_shm_zone_t              *shm_zone;

    ngx_resolver_t              *resolver;
    ngx_msec_t                   resolver_timeout;
} ngx_ssl_ocsp_conf_t;


typedef struct {
    ngx_rbtree_t                 rbtree;
    ngx_rbtree_node_t            sentinel;
    ngx_queue_t                  expire_queue;
} ngx_ssl_ocsp_cache_t;


typedef struct {
    ngx_str_node_t               node;
    ngx_queue_t                  queue;
    int                          status;
    time_t                       valid;
} ngx_ssl_ocsp_cache_node_t;


typedef struct ngx_ssl_ocsp_ctx_s  ngx_ssl_ocsp_ctx_t;


struct ngx_ssl_ocsp_s {
    STACK_OF(X509)              *certs;
    ngx_uint_t                   ncert;

    int                          cert_status;
    ngx_int_t                    status;

    ngx_ssl_ocsp_conf_t         *conf;
    ngx_ssl_ocsp_ctx_t          *ctx;
};


struct ngx_ssl_ocsp_ctx_s {
    SSL_CTX                     *ssl_ctx;

    X509                        *cert;
    X509                        *issuer;
    STACK_OF(X509)              *chain;

    int                          status;
    time_t                       valid;

    u_char                      *name;

    ngx_uint_t                   naddrs;
    ngx_uint_t                   naddr;

    ngx_addr_t                  *addrs;
    ngx_str_t                    host;
    ngx_str_t                    uri;
    in_port_t                    port;

    ngx_resolver_t              *resolver;
    ngx_msec_t                   resolver_timeout;

    ngx_msec_t                   timeout;

    void                       (*handler)(ngx_ssl_ocsp_ctx_t *ctx);
    void                        *data;

    ngx_str_t                    key;
    ngx_buf_t                   *request;
    ngx_buf_t                   *response;
    ngx_peer_connection_t        peer;

    ngx_shm_zone_t              *shm_zone;

    ngx_int_t                  (*process)(ngx_ssl_ocsp_ctx_t *ctx);

    ngx_uint_t                   state;

    ngx_uint_t                   code;
    ngx_uint_t                   count;
    ngx_uint_t                   flags;
    ngx_uint_t                   done;

    u_char                      *header_name_start;
    u_char                      *header_name_end;
    u_char                      *header_start;
    u_char                      *header_end;

    ngx_pool_t                  *pool;
    ngx_log_t                   *log;
};


static ngx_int_t ngx_ssl_stapling_certificate(ngx_conf_t *cf, ngx_ssl_t *ssl,
    X509 *cert, ngx_str_t *file, ngx_str_t *responder, ngx_uint_t verify);
static ngx_int_t ngx_ssl_stapling_file(ngx_conf_t *cf, ngx_ssl_t *ssl,
    ngx_ssl_stapling_t *staple, ngx_str_t *file);
static ngx_int_t ngx_ssl_stapling_issuer(ngx_conf_t *cf, ngx_ssl_t *ssl,
    ngx_ssl_stapling_t *staple);
static ngx_int_t ngx_ssl_stapling_responder(ngx_conf_t *cf, ngx_ssl_t *ssl,
    ngx_ssl_stapling_t *staple, ngx_str_t *responder);

static int ngx_ssl_certificate_status_callback(ngx_ssl_conn_t *ssl_conn,
    void *data);
static void ngx_ssl_stapling_update(ngx_ssl_stapling_t *staple);
static void ngx_ssl_stapling_ocsp_handler(ngx_ssl_ocsp_ctx_t *ctx);

static time_t ngx_ssl_stapling_time(ASN1_GENERALIZEDTIME *asn1time);

static void ngx_ssl_stapling_cleanup(void *data);

static void ngx_ssl_ocsp_validate_next(ngx_connection_t *c);
static void ngx_ssl_ocsp_handler(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_responder(ngx_connection_t *c,
    ngx_ssl_ocsp_ctx_t *ctx);

static ngx_ssl_ocsp_ctx_t *ngx_ssl_ocsp_start(ngx_log_t *log);
static void ngx_ssl_ocsp_done(ngx_ssl_ocsp_ctx_t *ctx);
static void ngx_ssl_ocsp_next(ngx_ssl_ocsp_ctx_t *ctx);
static void ngx_ssl_ocsp_request(ngx_ssl_ocsp_ctx_t *ctx);
static void ngx_ssl_ocsp_resolve_handler(ngx_resolver_ctx_t *resolve);
static void ngx_ssl_ocsp_connect(ngx_ssl_ocsp_ctx_t *ctx);
static void ngx_ssl_ocsp_write_handler(ngx_event_t *wev);
static void ngx_ssl_ocsp_read_handler(ngx_event_t *rev);
static void ngx_ssl_ocsp_dummy_handler(ngx_event_t *ev);

static ngx_int_t ngx_ssl_ocsp_create_request(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_process_status_line(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_parse_status_line(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_process_headers(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_parse_header_line(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_process_body(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_verify(ngx_ssl_ocsp_ctx_t *ctx);

static ngx_int_t ngx_ssl_ocsp_cache_lookup(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_cache_store(ngx_ssl_ocsp_ctx_t *ctx);
static ngx_int_t ngx_ssl_ocsp_create_key(ngx_ssl_ocsp_ctx_t *ctx);

static u_char *ngx_ssl_ocsp_log_error(ngx_log_t *log, u_char *buf, size_t len);




void
ngx_ssl_ocsp_cleanup(ngx_connection_t *c)
{
    ngx_ssl_ocsp_t  *ocsp;

    ocsp = c->ssl->ocsp;
    if (ocsp == NULL) {
        return;
    }

    if (ocsp->ctx) {
        ngx_ssl_ocsp_done(ocsp->ctx);
        ocsp->ctx = NULL;
    }

    if (ocsp->certs) {
        sk_X509_pop_free(ocsp->certs, X509_free);
        ocsp->certs = NULL;
    }
}

static void
ngx_ssl_ocsp_done(ngx_ssl_ocsp_ctx_t *ctx)
{
    ngx_log_debug0(NGX_LOG_DEBUG_EVENT, ctx->log, 0,
                   "ssl ocsp done");

    if (ctx->peer.connection) {
        ngx_close_connection(ctx->peer.connection);
    }

    ngx_destroy_pool(ctx->pool);
}


#endif