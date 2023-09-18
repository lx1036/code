
#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_event.h>


#define NGX_SSL_PASSWORD_BUFFER_SIZE  4096


typedef struct {
    ngx_uint_t  engine;   /* unsigned  engine:1; */
} ngx_openssl_conf_t;


static X509 *ngx_ssl_load_certificate(ngx_pool_t *pool, char **err,
    ngx_str_t *cert, STACK_OF(X509) **chain);
static EVP_PKEY *ngx_ssl_load_certificate_key(ngx_pool_t *pool, char **err,
    ngx_str_t *key, ngx_array_t *passwords);
static int ngx_ssl_password_callback(char *buf, int size, int rwflag,
    void *userdata);
static int ngx_ssl_verify_callback(int ok, X509_STORE_CTX *x509_store);
static void ngx_ssl_info_callback(const ngx_ssl_conn_t *ssl_conn, int where,
    int ret);
static void ngx_ssl_passwords_cleanup(void *data);
static int ngx_ssl_new_client_session(ngx_ssl_conn_t *ssl_conn,
    ngx_ssl_session_t *sess);
#ifdef SSL_READ_EARLY_DATA_SUCCESS
static ngx_int_t ngx_ssl_try_early_data(ngx_connection_t *c);
#endif
#if (NGX_DEBUG)
static void ngx_ssl_handshake_log(ngx_connection_t *c);
#endif
static void ngx_ssl_handshake_handler(ngx_event_t *ev);
#ifdef SSL_READ_EARLY_DATA_SUCCESS
static ssize_t ngx_ssl_recv_early(ngx_connection_t *c, u_char *buf,
    size_t size);
#endif
static ngx_int_t ngx_ssl_handle_recv(ngx_connection_t *c, int n);
static void ngx_ssl_write_handler(ngx_event_t *wev);
#ifdef SSL_READ_EARLY_DATA_SUCCESS
static ssize_t ngx_ssl_write_early(ngx_connection_t *c, u_char *data,
    size_t size);
#endif
static ssize_t ngx_ssl_sendfile(ngx_connection_t *c, ngx_buf_t *file,
    size_t size);
static void ngx_ssl_read_handler(ngx_event_t *rev);
static void ngx_ssl_shutdown_handler(ngx_event_t *ev);
static void ngx_ssl_connection_error(ngx_connection_t *c, int sslerr,
    ngx_err_t err, char *text);
static void ngx_ssl_clear_error(ngx_log_t *log);

static ngx_int_t ngx_ssl_session_id_context(ngx_ssl_t *ssl,
    ngx_str_t *sess_ctx, ngx_array_t *certificates);
static int ngx_ssl_new_session(ngx_ssl_conn_t *ssl_conn,
    ngx_ssl_session_t *sess);
static ngx_ssl_session_t *ngx_ssl_get_cached_session(ngx_ssl_conn_t *ssl_conn,
#if OPENSSL_VERSION_NUMBER >= 0x10100003L
    const
#endif
    u_char *id, int len, int *copy);
static void ngx_ssl_remove_session(SSL_CTX *ssl, ngx_ssl_session_t *sess);
static void ngx_ssl_expire_sessions(ngx_ssl_session_cache_t *cache,
    ngx_slab_pool_t *shpool, ngx_uint_t n);
static void ngx_ssl_session_rbtree_insert_value(ngx_rbtree_node_t *temp,
    ngx_rbtree_node_t *node, ngx_rbtree_node_t *sentinel);

#ifdef SSL_CTRL_SET_TLSEXT_TICKET_KEY_CB
static int ngx_ssl_ticket_key_callback(ngx_ssl_conn_t *ssl_conn,
    unsigned char *name, unsigned char *iv, EVP_CIPHER_CTX *ectx,
    HMAC_CTX *hctx, int enc);
static ngx_int_t ngx_ssl_rotate_ticket_keys(SSL_CTX *ssl_ctx, ngx_log_t *log);
static void ngx_ssl_ticket_keys_cleanup(void *data);
#endif

#ifndef X509_CHECK_FLAG_ALWAYS_CHECK_SUBJECT
static ngx_int_t ngx_ssl_check_name(ngx_str_t *name, ASN1_STRING *str);
#endif

static time_t ngx_ssl_parse_time(
#if OPENSSL_VERSION_NUMBER > 0x10100000L
    const
#endif
    ASN1_TIME *asn1time, ngx_log_t *log);

static void *ngx_openssl_create_conf(ngx_cycle_t *cycle);
static char *ngx_openssl_engine(ngx_conf_t *cf, ngx_command_t *cmd, void *conf);
static void ngx_openssl_exit(ngx_cycle_t *cycle);

#if defined(T_NGX_HAVE_DTLS)
#define NGX_DTLS_MAX_RETRANSMIT_TIME 3

static void ngx_dtls_retransmit(ngx_event_t *ev);
static ngx_int_t ngx_dtls_handshake(ngx_connection_t *c);

static int ngx_dtls_client_hmac(SSL *ssl, u_char res[EVP_MAX_MD_SIZE],
    unsigned int *rlen);

static int ngx_dtls_generate_cookie_cb(SSL *ssl, unsigned char *cookie,
    unsigned int *cookie_len);

static int ngx_dtls_verify_cookie_cb(SSL *ssl,
#if OPENSSL_VERSION_NUMBER >= 0x10100000L
    const
#endif
unsigned char *cookie, unsigned int cookie_len);
#endif

int  ngx_ssl_connection_index;
int  ngx_ssl_server_conf_index;
int  ngx_ssl_session_cache_index;
int  ngx_ssl_ticket_keys_index;
int  ngx_ssl_ocsp_index;
int  ngx_ssl_certificate_index;
int  ngx_ssl_next_certificate_index;
int  ngx_ssl_certificate_name_index;
int  ngx_ssl_stapling_index;

ngx_int_t ngx_ssl_shutdown(ngx_connection_t *c) {
    int         n, sslerr, mode;
    ngx_int_t   rc;
    ngx_err_t   err;
    ngx_uint_t  tries;

#if (T_NGX_HAVE_DTLS)
    if (c->ssl->retrans && c->ssl->retrans->timer_set) {
        ngx_del_timer(c->ssl->retrans);
    }
#endif

    rc = NGX_OK;

    ngx_ssl_ocsp_cleanup(c);

    if (SSL_in_init(c->ssl->connection)) {
        /*
         * OpenSSL 1.0.2f complains if SSL_shutdown() is called during
         * an SSL handshake, while previous versions always return 0.
         * Avoid calling SSL_shutdown() if handshake wasn't completed.
         */

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable) {
            /* Check if there is inflight request */
            if (SSL_want_async(c->ssl->connection) && !c->timedout) {
                c->async->handler = ngx_ssl_shutdown_async_handler;
                ngx_ssl_async_process_fds(c);
                ngx_add_timer(c->async, 300);
                return NGX_AGAIN;
            }

            /* Ignore errors from ngx_ssl_async_process_fds as
               we want to carry on and close the SSL connection
               anyway. */
            ngx_ssl_async_process_fds(c);
            if (ngx_del_async_conn) {
                if (c->num_async_fds) {
                    ngx_del_async_conn(c, NGX_DISABLE_EVENT);
                    c->num_async_fds--;
                }
            }
            ngx_del_conn(c, NGX_DISABLE_EVENT);
        }
#endif

        goto done;
    }

    if (c->timedout || c->error || c->buffered) {
        mode = SSL_RECEIVED_SHUTDOWN|SSL_SENT_SHUTDOWN;
        SSL_set_quiet_shutdown(c->ssl->connection, 1);

    } else {
        mode = SSL_get_shutdown(c->ssl->connection);

        if (c->ssl->no_wait_shutdown) {
            mode |= SSL_RECEIVED_SHUTDOWN;
        }

        if (c->ssl->no_send_shutdown) {
            mode |= SSL_SENT_SHUTDOWN;
        }

        if (c->ssl->no_wait_shutdown && c->ssl->no_send_shutdown) {
            SSL_set_quiet_shutdown(c->ssl->connection, 1);
        }
    }

    SSL_set_shutdown(c->ssl->connection, mode);

    ngx_ssl_clear_error(c->log);

    tries = 2;

    for ( ;; ) {

        /*
         * For bidirectional shutdown, SSL_shutdown() needs to be called
         * twice: first call sends the "close notify" alert and returns 0,
         * second call waits for the peer's "close notify" alert.
         */

        n = SSL_shutdown(c->ssl->connection);

        ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_shutdown: %d", n);

        if (n == 1) {
#if (NGX_SSL && NGX_SSL_ASYNC)
            if (c->async_enable) {
                /* Ignore errors from ngx_ssl_async_process_fds as
                    we want to carry on and close the SSL connection
                    anyway. */
                ngx_ssl_async_process_fds(c);
                if (ngx_del_async_conn) {
                    if (c->num_async_fds) {
                        ngx_del_async_conn(c, NGX_DISABLE_EVENT);
                        c->num_async_fds--;
                    }
                }
                ngx_del_conn(c, NGX_DISABLE_EVENT);
            }
#endif
            goto done;
        }

        if (n == 0 && tries-- > 1) {
            continue;
        }

        /* before 0.9.8m SSL_shutdown() returned 0 instead of -1 on errors */

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif
        sslerr = SSL_get_error(c->ssl->connection, n);

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL_get_error: %d", sslerr);

        if (sslerr == SSL_ERROR_WANT_READ || sslerr == SSL_ERROR_WANT_WRITE) {
#if (NGX_SSL && NGX_SSL_ASYNC)
            if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
                return NGX_ERROR;
            }
#endif
            c->read->handler = ngx_ssl_shutdown_handler;
            c->write->handler = ngx_ssl_shutdown_handler;

            if (sslerr == SSL_ERROR_WANT_READ) {
                c->read->ready = 0;

            } else {
                c->write->ready = 0;
            }

            if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
                goto failed;
            }

            if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
                goto failed;
            }

            ngx_add_timer(c->read, 3000);

            return NGX_AGAIN;
        }

#if (NGX_SSL && NGX_SSL_ASYNC)
    if (c->async_enable) {
        if (sslerr == SSL_ERROR_WANT_ASYNC) {
            c->async->handler = ngx_ssl_shutdown_async_handler;
            c->read->saved_handler = ngx_ssl_shutdown_handler;
            c->read->handler = ngx_ssl_empty_handler;
            c->write->handler = ngx_ssl_shutdown_handler;

            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                           "SSL ASYNC WANT recieved: \"%s\"", __func__);

            /* Ignore errors from ngx_ssl_async_process_fds as
               we want to carry on anyway */
            ngx_ssl_async_process_fds(c);
            return NGX_AGAIN;
        }

        /* Ignore errors from ngx_ssl_async_process_fds as
           we want to carry on and close the SSL connection
           anyway. */
        ngx_ssl_async_process_fds(c);
        if (ngx_del_async_conn) {
            if (c->num_async_fds) {
                ngx_del_async_conn(c, NGX_DISABLE_EVENT);
                c->num_async_fds--;
            }
        }
        ngx_del_conn(c, NGX_DISABLE_EVENT);
    }
#endif

        if (sslerr == SSL_ERROR_ZERO_RETURN || ERR_peek_error() == 0) {
            goto done;
        }

        err = (sslerr == SSL_ERROR_SYSCALL) ? ngx_errno : 0;

        ngx_ssl_connection_error(c, sslerr, err, "SSL_shutdown() failed");

        break;
    }

failed:

    rc = NGX_ERROR;

done:

    if (c->ssl->shutdown_without_free) {
        c->ssl->shutdown_without_free = 0;
        c->recv = ngx_recv;
        return rc;
    }

    SSL_free(c->ssl->connection);
    c->ssl = NULL;
    c->recv = ngx_recv;

    return rc;
}

static void
ngx_ssl_shutdown_handler(ngx_event_t *ev)
{
    ngx_connection_t           *c;
    ngx_connection_handler_pt   handler;

    c = ev->data;
    handler = c->ssl->handler;

    if (ev->timedout) {
        c->timedout = 1;
    }

    ngx_log_error(NGX_LOG_STDERR, ev->log, 0, "SSL shutdown handler");

    if (ngx_ssl_shutdown(c) == NGX_AGAIN) {
        return;
    }

#if (NGX_SSL && NGX_SSL_ASYNC)
    /*
     * empty the handler of async event to avoid
     * going back to previous ssl shutdown state
     */
    c->async->handler = ngx_ssl_empty_handler;
#endif
    handler(c);
}

static void
ngx_ssl_clear_error(ngx_log_t *log)
{
    while (ERR_peek_error()) {
        ngx_ssl_error(NGX_LOG_ALERT, log, 0, "ignoring stale global SSL error");
    }

    ERR_clear_error();
}

static void ngx_ssl_connection_error(ngx_connection_t *c, int sslerr, ngx_err_t err, char *text) {
    int         n;
    ngx_uint_t  level;

    level = NGX_LOG_CRIT;
    if (sslerr == SSL_ERROR_SYSCALL) {
        if (err == NGX_ECONNRESET
            || err == NGX_EPIPE
            || err == NGX_ENOTCONN
            || err == NGX_ETIMEDOUT
            || err == NGX_ECONNREFUSED
            || err == NGX_ENETDOWN
            || err == NGX_ENETUNREACH
            || err == NGX_EHOSTDOWN
            || err == NGX_EHOSTUNREACH)
        {
            switch (c->log_error) {

            case NGX_ERROR_IGNORE_ECONNRESET:
            case NGX_ERROR_INFO:
                level = NGX_LOG_INFO;
                break;

            case NGX_ERROR_ERR:
                level = NGX_LOG_ERR;
                break;

            default:
                break;
            }
        }

    } else if (sslerr == SSL_ERROR_SSL) {

        n = ERR_GET_REASON(ERR_peek_last_error());

            /* handshake failures */
        if (n == SSL_R_BAD_CHANGE_CIPHER_SPEC                        /*  103 */
#ifdef SSL_R_NO_SUITABLE_KEY_SHARE
            || n == SSL_R_NO_SUITABLE_KEY_SHARE                      /*  101 */
#endif
#ifdef SSL_R_BAD_ALERT
            || n == SSL_R_BAD_ALERT                                  /*  102 */
#endif
#ifdef SSL_R_BAD_KEY_SHARE
            || n == SSL_R_BAD_KEY_SHARE                              /*  108 */
#endif
#ifdef SSL_R_BAD_EXTENSION
            || n == SSL_R_BAD_EXTENSION                              /*  110 */
#endif
            || n == SSL_R_BAD_DIGEST_LENGTH                          /*  111 */
#ifdef SSL_R_MISSING_SIGALGS_EXTENSION
            || n == SSL_R_MISSING_SIGALGS_EXTENSION                  /*  112 */
#endif
            || n == SSL_R_BAD_PACKET_LENGTH                          /*  115 */
#ifdef SSL_R_NO_SUITABLE_SIGNATURE_ALGORITHM
            || n == SSL_R_NO_SUITABLE_SIGNATURE_ALGORITHM            /*  118 */
#endif
#ifdef SSL_R_BAD_KEY_UPDATE
            || n == SSL_R_BAD_KEY_UPDATE                             /*  122 */
#endif
            || n == SSL_R_BLOCK_CIPHER_PAD_IS_WRONG                  /*  129 */
            || n == SSL_R_CCS_RECEIVED_EARLY                         /*  133 */
#ifdef SSL_R_DECODE_ERROR
            || n == SSL_R_DECODE_ERROR                               /*  137 */
#endif
#ifdef SSL_R_DATA_BETWEEN_CCS_AND_FINISHED
            || n == SSL_R_DATA_BETWEEN_CCS_AND_FINISHED              /*  145 */
#endif
            || n == SSL_R_DATA_LENGTH_TOO_LONG                       /*  146 */
            || n == SSL_R_DIGEST_CHECK_FAILED                        /*  149 */
            || n == SSL_R_ENCRYPTED_LENGTH_TOO_LONG                  /*  150 */
            || n == SSL_R_ERROR_IN_RECEIVED_CIPHER_LIST              /*  151 */
            || n == SSL_R_EXCESSIVE_MESSAGE_SIZE                     /*  152 */
#ifdef SSL_R_GOT_A_FIN_BEFORE_A_CCS
            || n == SSL_R_GOT_A_FIN_BEFORE_A_CCS                     /*  154 */
#endif
            || n == SSL_R_HTTPS_PROXY_REQUEST                        /*  155 */
            || n == SSL_R_HTTP_REQUEST                               /*  156 */
            || n == SSL_R_LENGTH_MISMATCH                            /*  159 */
#ifdef SSL_R_LENGTH_TOO_SHORT
            || n == SSL_R_LENGTH_TOO_SHORT                           /*  160 */
#endif
#ifdef SSL_R_NO_RENEGOTIATION
            || n == SSL_R_NO_RENEGOTIATION                           /*  182 */
#endif
#ifdef SSL_R_NO_CIPHERS_PASSED
            || n == SSL_R_NO_CIPHERS_PASSED                          /*  182 */
#endif
            || n == SSL_R_NO_CIPHERS_SPECIFIED                       /*  183 */
#ifdef SSL_R_BAD_CIPHER
            || n == SSL_R_BAD_CIPHER                                 /*  186 */
#endif
            || n == SSL_R_NO_COMPRESSION_SPECIFIED                   /*  187 */
            || n == SSL_R_NO_SHARED_CIPHER                           /*  193 */
#ifdef SSL_R_PACKET_LENGTH_TOO_LONG
            || n == SSL_R_PACKET_LENGTH_TOO_LONG                     /*  198 */
#endif
            || n == SSL_R_RECORD_LENGTH_MISMATCH                     /*  213 */
#ifdef SSL_R_TOO_MANY_WARNING_ALERTS
            || n == SSL_R_TOO_MANY_WARNING_ALERTS                    /*  220 */
#endif
#ifdef SSL_R_CLIENTHELLO_TLSEXT
            || n == SSL_R_CLIENTHELLO_TLSEXT                         /*  226 */
#endif
#ifdef SSL_R_PARSE_TLSEXT
            || n == SSL_R_PARSE_TLSEXT                               /*  227 */
#endif
#ifdef SSL_R_CALLBACK_FAILED
            || n == SSL_R_CALLBACK_FAILED                            /*  234 */
#endif
#ifdef SSL_R_TLS_RSA_ENCRYPTED_VALUE_LENGTH_IS_WRONG
            || n == SSL_R_TLS_RSA_ENCRYPTED_VALUE_LENGTH_IS_WRONG    /*  234 */
#endif
#ifdef SSL_R_NO_APPLICATION_PROTOCOL
            || n == SSL_R_NO_APPLICATION_PROTOCOL                    /*  235 */
#endif
            || n == SSL_R_UNEXPECTED_MESSAGE                         /*  244 */
            || n == SSL_R_UNEXPECTED_RECORD                          /*  245 */
            || n == SSL_R_UNKNOWN_ALERT_TYPE                         /*  246 */
            || n == SSL_R_UNKNOWN_PROTOCOL                           /*  252 */
#ifdef SSL_R_NO_COMMON_SIGNATURE_ALGORITHMS
            || n == SSL_R_NO_COMMON_SIGNATURE_ALGORITHMS             /*  253 */
#endif
#ifdef SSL_R_INVALID_COMPRESSION_LIST
            || n == SSL_R_INVALID_COMPRESSION_LIST                   /*  256 */
#endif
#ifdef SSL_R_MISSING_KEY_SHARE
            || n == SSL_R_MISSING_KEY_SHARE                          /*  258 */
#endif
            || n == SSL_R_UNSUPPORTED_PROTOCOL                       /*  258 */
#ifdef SSL_R_NO_SHARED_GROUP
            || n == SSL_R_NO_SHARED_GROUP                            /*  266 */
#endif
            || n == SSL_R_WRONG_VERSION_NUMBER                       /*  267 */
#ifdef SSL_R_TOO_MUCH_SKIPPED_EARLY_DATA
            || n == SSL_R_TOO_MUCH_SKIPPED_EARLY_DATA                /*  270 */
#endif
            || n == SSL_R_BAD_LENGTH                                 /*  271 */
            || n == SSL_R_DECRYPTION_FAILED_OR_BAD_RECORD_MAC        /*  281 */
#ifdef SSL_R_APPLICATION_DATA_AFTER_CLOSE_NOTIFY
            || n == SSL_R_APPLICATION_DATA_AFTER_CLOSE_NOTIFY        /*  291 */
#endif
#ifdef SSL_R_APPLICATION_DATA_ON_SHUTDOWN
            || n == SSL_R_APPLICATION_DATA_ON_SHUTDOWN               /*  291 */
#endif
#ifdef SSL_R_BAD_LEGACY_VERSION
            || n == SSL_R_BAD_LEGACY_VERSION                         /*  292 */
#endif
#ifdef SSL_R_MIXED_HANDSHAKE_AND_NON_HANDSHAKE_DATA
            || n == SSL_R_MIXED_HANDSHAKE_AND_NON_HANDSHAKE_DATA     /*  293 */
#endif
#ifdef SSL_R_RECORD_TOO_SMALL
            || n == SSL_R_RECORD_TOO_SMALL                           /*  298 */
#endif
#ifdef SSL_R_SSL3_SESSION_ID_TOO_LONG
            || n == SSL_R_SSL3_SESSION_ID_TOO_LONG                   /*  300 */
#endif
#ifdef SSL_R_BAD_ECPOINT
            || n == SSL_R_BAD_ECPOINT                                /*  306 */
#endif
#ifdef SSL_R_RENEGOTIATE_EXT_TOO_LONG
            || n == SSL_R_RENEGOTIATE_EXT_TOO_LONG                   /*  335 */
            || n == SSL_R_RENEGOTIATION_ENCODING_ERR                 /*  336 */
            || n == SSL_R_RENEGOTIATION_MISMATCH                     /*  337 */
#endif
#ifdef SSL_R_UNSAFE_LEGACY_RENEGOTIATION_DISABLED
            || n == SSL_R_UNSAFE_LEGACY_RENEGOTIATION_DISABLED       /*  338 */
#endif
#ifdef SSL_R_SCSV_RECEIVED_WHEN_RENEGOTIATING
            || n == SSL_R_SCSV_RECEIVED_WHEN_RENEGOTIATING           /*  345 */
#endif
#ifdef SSL_R_INAPPROPRIATE_FALLBACK
            || n == SSL_R_INAPPROPRIATE_FALLBACK                     /*  373 */
#endif
#ifdef SSL_R_NO_SHARED_SIGNATURE_ALGORITHMS
            || n == SSL_R_NO_SHARED_SIGNATURE_ALGORITHMS             /*  376 */
#endif
#ifdef SSL_R_NO_SHARED_SIGATURE_ALGORITHMS
            || n == SSL_R_NO_SHARED_SIGATURE_ALGORITHMS              /*  376 */
#endif
#ifdef SSL_R_CERT_CB_ERROR
            || n == SSL_R_CERT_CB_ERROR                              /*  377 */
#endif
#ifdef SSL_R_VERSION_TOO_LOW
            || n == SSL_R_VERSION_TOO_LOW                            /*  396 */
#endif
#ifdef SSL_R_TOO_MANY_WARN_ALERTS
            || n == SSL_R_TOO_MANY_WARN_ALERTS                       /*  409 */
#endif
#ifdef SSL_R_BAD_RECORD_TYPE
            || n == SSL_R_BAD_RECORD_TYPE                            /*  443 */
#endif
            || n == 1000 /* SSL_R_SSLV3_ALERT_CLOSE_NOTIFY */
#ifdef SSL_R_SSLV3_ALERT_UNEXPECTED_MESSAGE
            || n == SSL_R_SSLV3_ALERT_UNEXPECTED_MESSAGE             /* 1010 */
            || n == SSL_R_SSLV3_ALERT_BAD_RECORD_MAC                 /* 1020 */
            || n == SSL_R_TLSV1_ALERT_DECRYPTION_FAILED              /* 1021 */
            || n == SSL_R_TLSV1_ALERT_RECORD_OVERFLOW                /* 1022 */
            || n == SSL_R_SSLV3_ALERT_DECOMPRESSION_FAILURE          /* 1030 */
            || n == SSL_R_SSLV3_ALERT_HANDSHAKE_FAILURE              /* 1040 */
            || n == SSL_R_SSLV3_ALERT_NO_CERTIFICATE                 /* 1041 */
            || n == SSL_R_SSLV3_ALERT_BAD_CERTIFICATE                /* 1042 */
            || n == SSL_R_SSLV3_ALERT_UNSUPPORTED_CERTIFICATE        /* 1043 */
            || n == SSL_R_SSLV3_ALERT_CERTIFICATE_REVOKED            /* 1044 */
            || n == SSL_R_SSLV3_ALERT_CERTIFICATE_EXPIRED            /* 1045 */
            || n == SSL_R_SSLV3_ALERT_CERTIFICATE_UNKNOWN            /* 1046 */
            || n == SSL_R_SSLV3_ALERT_ILLEGAL_PARAMETER              /* 1047 */
            || n == SSL_R_TLSV1_ALERT_UNKNOWN_CA                     /* 1048 */
            || n == SSL_R_TLSV1_ALERT_ACCESS_DENIED                  /* 1049 */
            || n == SSL_R_TLSV1_ALERT_DECODE_ERROR                   /* 1050 */
            || n == SSL_R_TLSV1_ALERT_DECRYPT_ERROR                  /* 1051 */
            || n == SSL_R_TLSV1_ALERT_EXPORT_RESTRICTION             /* 1060 */
            || n == SSL_R_TLSV1_ALERT_PROTOCOL_VERSION               /* 1070 */
            || n == SSL_R_TLSV1_ALERT_INSUFFICIENT_SECURITY          /* 1071 */
            || n == SSL_R_TLSV1_ALERT_INTERNAL_ERROR                 /* 1080 */
            || n == SSL_R_TLSV1_ALERT_USER_CANCELLED                 /* 1090 */
            || n == SSL_R_TLSV1_ALERT_NO_RENEGOTIATION               /* 1100 */
#endif
            )
        {
            switch (c->log_error) {

            case NGX_ERROR_IGNORE_ECONNRESET:
            case NGX_ERROR_INFO:
                level = NGX_LOG_INFO;
                break;

            case NGX_ERROR_ERR:
                level = NGX_LOG_ERR;
                break;

            default:
                break;
            }
        }
    }

    ngx_ssl_error(level, c->log, err, text);
}

void ngx_cdecl
ngx_ssl_error(ngx_uint_t level, ngx_log_t *log, ngx_err_t err, char *fmt, ...)
{
    int          flags;
    u_long       n;
    va_list      args;
    u_char      *p, *last;
    u_char       errstr[NGX_MAX_CONF_ERRSTR];
    const char  *data;

    last = errstr + NGX_MAX_CONF_ERRSTR;

    va_start(args, fmt);
    p = ngx_vslprintf(errstr, last - 1, fmt, args);
    va_end(args);

    if (ERR_peek_error()) {
        p = ngx_cpystrn(p, (u_char *) " (SSL:", last - p);

        for ( ;; ) {

            n = ERR_peek_error_data(&data, &flags);

            if (n == 0) {
                break;
            }

            /* ERR_error_string_n() requires at least one byte */

            if (p >= last - 1) {
                goto next;
            }

            *p++ = ' ';

            ERR_error_string_n(n, (char *) p, last - p);

            while (p < last && *p) {
                p++;
            }

            if (p < last && *data && (flags & ERR_TXT_STRING)) {
                *p++ = ':';
                p = ngx_cpystrn(p, (u_char *) data, last - p);
            }

        next:

            (void) ERR_get_error();
        }

        if (p < last) {
            *p++ = ')';
        }
    }

    ngx_log_error(level, log, err, "%*s", p - errstr, errstr);
}


ngx_int_t
ngx_ssl_certificate(ngx_conf_t *cf, ngx_ssl_t *ssl, ngx_str_t *cert,
    ngx_str_t *key, ngx_array_t *passwords)
{
    char            *err;
    X509            *x509;
    EVP_PKEY        *pkey;
    STACK_OF(X509)  *chain;

    x509 = ngx_ssl_load_certificate(cf->pool, &err, cert, &chain);
    if (x509 == NULL) {
        if (err != NULL) {
            ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                          "cannot load certificate \"%s\": %s",
                          cert->data, err);
        }

        return NGX_ERROR;
    }

    if (SSL_CTX_use_certificate(ssl->ctx, x509) == 0) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_use_certificate(\"%s\") failed", cert->data);
        X509_free(x509);
        sk_X509_pop_free(chain, X509_free);
        return NGX_ERROR;
    }

    if (X509_set_ex_data(x509, ngx_ssl_certificate_name_index, cert->data)
        == 0)
    {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0, "X509_set_ex_data() failed");
        X509_free(x509);
        sk_X509_pop_free(chain, X509_free);
        return NGX_ERROR;
    }

    if (X509_set_ex_data(x509, ngx_ssl_next_certificate_index,
                      SSL_CTX_get_ex_data(ssl->ctx, ngx_ssl_certificate_index))
        == 0)
    {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0, "X509_set_ex_data() failed");
        X509_free(x509);
        sk_X509_pop_free(chain, X509_free);
        return NGX_ERROR;
    }

    if (SSL_CTX_set_ex_data(ssl->ctx, ngx_ssl_certificate_index, x509) == 0) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_set_ex_data() failed");
        X509_free(x509);
        sk_X509_pop_free(chain, X509_free);
        return NGX_ERROR;
    }

    /*
     * Note that x509 is not freed here, but will be instead freed in
     * ngx_ssl_cleanup_ctx().  This is because we need to preserve all
     * certificates to be able to iterate all of them through exdata
     * (ngx_ssl_certificate_index, ngx_ssl_next_certificate_index),
     * while OpenSSL can free a certificate if it is replaced with another
     * certificate of the same type.
     */

#ifdef SSL_CTX_set0_chain

    if (SSL_CTX_set0_chain(ssl->ctx, chain) == 0) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_set0_chain(\"%s\") failed", cert->data);
        sk_X509_pop_free(chain, X509_free);
        return NGX_ERROR;
    }

#else
    {
    int  n;

    /* SSL_CTX_set0_chain() is only available in OpenSSL 1.0.2+ */

    n = sk_X509_num(chain);

    while (n--) {
        x509 = sk_X509_shift(chain);

        if (SSL_CTX_add_extra_chain_cert(ssl->ctx, x509) == 0) {
            ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                          "SSL_CTX_add_extra_chain_cert(\"%s\") failed",
                          cert->data);
            sk_X509_pop_free(chain, X509_free);
            return NGX_ERROR;
        }
    }

    sk_X509_free(chain);
    }
#endif

    pkey = ngx_ssl_load_certificate_key(cf->pool, &err, key, passwords);
    if (pkey == NULL) {
        if (err != NULL) {
            ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                          "cannot load certificate key \"%s\": %s",
                          key->data, err);
        }

        return NGX_ERROR;
    }

#if (T_NGX_SSL_NTLS)
    if (cert_tag == SSL_ENC_CERT) {
        if (SSL_CTX_use_enc_PrivateKey(ssl->ctx, pkey) == 0) {
            ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                          "SSL_CTX_use_enc_PrivateKey(\"%s\") failed",
                          key->data);
            EVP_PKEY_free(pkey);
            return NGX_ERROR;
        }

    } else if (cert_tag == SSL_SIGN_CERT) {
        if (SSL_CTX_use_sign_PrivateKey(ssl->ctx, pkey) == 0) {
            ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                          "SSL_CTX_use_sign_PrivateKey(\"%s\") failed",
                          key->data);
            EVP_PKEY_free(pkey);
            return NGX_ERROR;
        }

    } else
#endif
    if (SSL_CTX_use_PrivateKey(ssl->ctx, pkey) == 0) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_use_PrivateKey(\"%s\") failed", key->data);
        EVP_PKEY_free(pkey);
        return NGX_ERROR;
    }

    EVP_PKEY_free(pkey);

    return NGX_OK;
}

ngx_int_t
ngx_ssl_ciphers(ngx_conf_t *cf, ngx_ssl_t *ssl, ngx_str_t *ciphers,
    ngx_uint_t prefer_server_ciphers)
{
    if (SSL_CTX_set_cipher_list(ssl->ctx, (char *) ciphers->data) == 0) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_set_cipher_list(\"%V\") failed",
                      ciphers);
        return NGX_ERROR;
    }

    if (prefer_server_ciphers) {
        SSL_CTX_set_options(ssl->ctx, SSL_OP_CIPHER_SERVER_PREFERENCE);
    }

    return NGX_OK;
}

void
ngx_ssl_cleanup_ctx(void *data)
{
    ngx_ssl_t  *ssl = data;

    X509  *cert, *next;

    cert = SSL_CTX_get_ex_data(ssl->ctx, ngx_ssl_certificate_index);

    while (cert) {
        next = X509_get_ex_data(cert, ngx_ssl_next_certificate_index);
        X509_free(cert);
        cert = next;
    }

    SSL_CTX_free(ssl->ctx);
}

ngx_int_t
ngx_ssl_client_session_cache(ngx_conf_t *cf, ngx_ssl_t *ssl, ngx_uint_t enable)
{
    if (!enable) {
        return NGX_OK;
    }

    SSL_CTX_set_session_cache_mode(ssl->ctx,
                                   SSL_SESS_CACHE_CLIENT
                                   |SSL_SESS_CACHE_NO_INTERNAL);

    SSL_CTX_sess_set_new_cb(ssl->ctx, ngx_ssl_new_client_session);

    return NGX_OK;
}

ngx_int_t
ngx_ssl_conf_commands(ngx_conf_t *cf, ngx_ssl_t *ssl, ngx_array_t *commands)
{
    if (commands == NULL) {
        return NGX_OK;
    }

#ifdef SSL_CONF_FLAG_FILE
    {
    int            type;
    u_char        *key, *value;
    ngx_uint_t     i;
    ngx_keyval_t  *cmd;
    SSL_CONF_CTX  *cctx;

    cctx = SSL_CONF_CTX_new();
    if (cctx == NULL) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CONF_CTX_new() failed");
        return NGX_ERROR;
    }

    SSL_CONF_CTX_set_flags(cctx, SSL_CONF_FLAG_FILE);
    SSL_CONF_CTX_set_flags(cctx, SSL_CONF_FLAG_SERVER);
    SSL_CONF_CTX_set_flags(cctx, SSL_CONF_FLAG_CLIENT);
    SSL_CONF_CTX_set_flags(cctx, SSL_CONF_FLAG_CERTIFICATE);
    SSL_CONF_CTX_set_flags(cctx, SSL_CONF_FLAG_SHOW_ERRORS);

    SSL_CONF_CTX_set_ssl_ctx(cctx, ssl->ctx);

    cmd = commands->elts;
    for (i = 0; i < commands->nelts; i++) {

        key = cmd[i].key.data;
        type = SSL_CONF_cmd_value_type(cctx, (char *) key);

        if (type == SSL_CONF_TYPE_FILE || type == SSL_CONF_TYPE_DIR) {
            if (ngx_conf_full_name(cf->cycle, &cmd[i].value, 1) != NGX_OK) {
                SSL_CONF_CTX_free(cctx);
                return NGX_ERROR;
            }
        }

        value = cmd[i].value.data;

        if (SSL_CONF_cmd(cctx, (char *) key, (char *) value) <= 0) {
            ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                          "SSL_CONF_cmd(\"%s\", \"%s\") failed", key, value);
            SSL_CONF_CTX_free(cctx);
            return NGX_ERROR;
        }
    }

    if (SSL_CONF_CTX_finish(cctx) != 1) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CONF_finish() failed");
        SSL_CONF_CTX_free(cctx);
        return NGX_ERROR;
    }

    SSL_CONF_CTX_free(cctx);

    return NGX_OK;
    }
#else
    ngx_log_error(NGX_LOG_EMERG, ssl->log, 0,
                  "SSL_CONF_cmd() is not available on this platform");
    return NGX_ERROR;
#endif
}

ngx_int_t
ngx_ssl_create(ngx_ssl_t *ssl, ngx_uint_t protocols, void *data)
{
#if defined(T_NGX_HAVE_DTLS)
    if (protocols & NGX_SSL_DTLSv1 || protocols & NGX_SSL_DTLSv1_2) {


#if OPENSSL_VERSION_NUMBER < 0x10100000L

        if (protocols & NGX_SSL_DTLSv1_2) {

            /* DTLS 1.2 is only supported since 1.0.2 */

            /* DTLSv1_x_method()  functions are deprecated in 1.1.0 */

#if OPENSSL_VERSION_NUMBER < 0x10002000L

            /* ancient ... 1.0.2 */
            ngx_log_error(NGX_LOG_EMERG, ssl->log, 0,
                          "DTLSv1.2 is not supported by "
                          "the used version of OpenSSL");
            return NGX_ERROR;

#else
            /* 1.0.2 ... 1.1 */
            ssl->ctx = SSL_CTX_new(DTLSv1_2_method());
#endif
        }

        /* note: either 1.2 or 1.1 methods may be initialized, not both,
         * preferred is 1.2 if both specified in ssl_protocols
         */

        if (protocols & NGX_SSL_DTLSv1 && ssl->ctx == NULL) {
            ssl->ctx = SSL_CTX_new(DTLSv1_method());
        }
#else
        ssl->ctx = SSL_CTX_new(DTLS_method());
#endif

    } else {
        ssl->ctx = SSL_CTX_new(SSLv23_method());
    }
#else

    ssl->ctx = SSL_CTX_new(SSLv23_method());
#endif

    if (ssl->ctx == NULL) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0, "SSL_CTX_new() failed");
        return NGX_ERROR;
    }

#if defined(T_NGX_HAVE_DTLS)
    if (protocols & NGX_SSL_DTLSv1
        || protocols & NGX_SSL_DTLSv1_2) {

        SSL_CTX_set_cookie_generate_cb(ssl->ctx, ngx_dtls_generate_cookie_cb);
        SSL_CTX_set_cookie_verify_cb(ssl->ctx, ngx_dtls_verify_cookie_cb);
    }
#endif

    if (SSL_CTX_set_ex_data(ssl->ctx, ngx_ssl_server_conf_index, data) == 0) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_set_ex_data() failed");
        return NGX_ERROR;
    }

    if (SSL_CTX_set_ex_data(ssl->ctx, ngx_ssl_certificate_index, NULL) == 0) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_set_ex_data() failed");
        return NGX_ERROR;
    }

    ssl->buffer_size = NGX_SSL_BUFSIZE;

    /* client side options */

#ifdef SSL_OP_MICROSOFT_SESS_ID_BUG
    SSL_CTX_set_options(ssl->ctx, SSL_OP_MICROSOFT_SESS_ID_BUG);
#endif

#ifdef SSL_OP_NETSCAPE_CHALLENGE_BUG
    SSL_CTX_set_options(ssl->ctx, SSL_OP_NETSCAPE_CHALLENGE_BUG);
#endif

    /* server side options */

#ifdef SSL_OP_SSLREF2_REUSE_CERT_TYPE_BUG
    SSL_CTX_set_options(ssl->ctx, SSL_OP_SSLREF2_REUSE_CERT_TYPE_BUG);
#endif

#ifdef SSL_OP_MICROSOFT_BIG_SSLV3_BUFFER
    SSL_CTX_set_options(ssl->ctx, SSL_OP_MICROSOFT_BIG_SSLV3_BUFFER);
#endif

#ifdef SSL_OP_SSLEAY_080_CLIENT_DH_BUG
    SSL_CTX_set_options(ssl->ctx, SSL_OP_SSLEAY_080_CLIENT_DH_BUG);
#endif

#ifdef SSL_OP_TLS_D5_BUG
    SSL_CTX_set_options(ssl->ctx, SSL_OP_TLS_D5_BUG);
#endif

#ifdef SSL_OP_TLS_BLOCK_PADDING_BUG
    SSL_CTX_set_options(ssl->ctx, SSL_OP_TLS_BLOCK_PADDING_BUG);
#endif

#ifdef SSL_OP_DONT_INSERT_EMPTY_FRAGMENTS
    SSL_CTX_set_options(ssl->ctx, SSL_OP_DONT_INSERT_EMPTY_FRAGMENTS);
#endif

    SSL_CTX_set_options(ssl->ctx, SSL_OP_SINGLE_DH_USE);

#if OPENSSL_VERSION_NUMBER >= 0x009080dfL
    /* only in 0.9.8m+ */
    SSL_CTX_clear_options(ssl->ctx,
                          SSL_OP_NO_SSLv2|SSL_OP_NO_SSLv3|SSL_OP_NO_TLSv1);
#endif

    if (!(protocols & NGX_SSL_SSLv2)) {
        SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_SSLv2);
    }
    if (!(protocols & NGX_SSL_SSLv3)) {
        SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_SSLv3);
    }
    if (!(protocols & NGX_SSL_TLSv1)) {
        SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_TLSv1);
    }
#ifdef SSL_OP_NO_TLSv1_1
    SSL_CTX_clear_options(ssl->ctx, SSL_OP_NO_TLSv1_1);
    if (!(protocols & NGX_SSL_TLSv1_1)) {
        SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_TLSv1_1);
    }
#endif
#ifdef SSL_OP_NO_TLSv1_2
    SSL_CTX_clear_options(ssl->ctx, SSL_OP_NO_TLSv1_2);
    if (!(protocols & NGX_SSL_TLSv1_2)) {
        SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_TLSv1_2);
    }
#endif
#ifdef SSL_OP_NO_TLSv1_3
    SSL_CTX_clear_options(ssl->ctx, SSL_OP_NO_TLSv1_3);
    if (!(protocols & NGX_SSL_TLSv1_3)) {
        SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_TLSv1_3);
    }
#endif

#ifdef SSL_CTX_set_min_proto_version
    SSL_CTX_set_min_proto_version(ssl->ctx, 0);
    SSL_CTX_set_max_proto_version(ssl->ctx, TLS1_2_VERSION);
#endif

#ifdef TLS1_3_VERSION
    SSL_CTX_set_min_proto_version(ssl->ctx, 0);
    SSL_CTX_set_max_proto_version(ssl->ctx, TLS1_3_VERSION);
#endif

#ifdef SSL_OP_NO_COMPRESSION
    SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_COMPRESSION);
#endif

#ifdef SSL_OP_NO_ANTI_REPLAY
    SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_ANTI_REPLAY);
#endif

#ifdef SSL_OP_NO_CLIENT_RENEGOTIATION
    SSL_CTX_set_options(ssl->ctx, SSL_OP_NO_CLIENT_RENEGOTIATION);
#endif

#ifdef SSL_OP_IGNORE_UNEXPECTED_EOF
    SSL_CTX_set_options(ssl->ctx, SSL_OP_IGNORE_UNEXPECTED_EOF);
#endif

#ifdef SSL_MODE_RELEASE_BUFFERS
    SSL_CTX_set_mode(ssl->ctx, SSL_MODE_RELEASE_BUFFERS);
#endif

#ifdef SSL_MODE_NO_AUTO_CHAIN
    SSL_CTX_set_mode(ssl->ctx, SSL_MODE_NO_AUTO_CHAIN);
#endif

#if (NGX_SSL && NGX_SSL_ASYNC)
    if (ssl->async_enable) {
        SSL_CTX_set_mode(ssl->ctx, SSL_MODE_ASYNC);
    }
#endif

    SSL_CTX_set_read_ahead(ssl->ctx, 1);

    SSL_CTX_set_info_callback(ssl->ctx, ngx_ssl_info_callback);

    return NGX_OK;
}

ngx_int_t
ngx_ssl_crl(ngx_conf_t *cf, ngx_ssl_t *ssl, ngx_str_t *crl)
{
    X509_STORE   *store;
    X509_LOOKUP  *lookup;

    if (crl->len == 0) {
        return NGX_OK;
    }

    if (ngx_conf_full_name(cf->cycle, crl, 1) != NGX_OK) {
        return NGX_ERROR;
    }

    store = SSL_CTX_get_cert_store(ssl->ctx);

    if (store == NULL) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_get_cert_store() failed");
        return NGX_ERROR;
    }

    lookup = X509_STORE_add_lookup(store, X509_LOOKUP_file());

    if (lookup == NULL) {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "X509_STORE_add_lookup() failed");
        return NGX_ERROR;
    }

    if (X509_LOOKUP_load_file(lookup, (char *) crl->data, X509_FILETYPE_PEM)
        == 0)
    {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "X509_LOOKUP_load_file(\"%s\") failed", crl->data);
        return NGX_ERROR;
    }

    X509_STORE_set_flags(store,
                         X509_V_FLAG_CRL_CHECK|X509_V_FLAG_CRL_CHECK_ALL);

    return NGX_OK;
}

ngx_array_t *
ngx_ssl_preserve_passwords(ngx_conf_t *cf, ngx_array_t *passwords)
{
    ngx_str_t           *opwd, *pwd;
    ngx_uint_t           i;
    ngx_array_t         *pwds;
    ngx_pool_cleanup_t  *cln;
    static ngx_array_t   empty_passwords;

    if (passwords == NULL) {

        /*
         * If there are no passwords, an empty array is used
         * to make sure OpenSSL's default password callback
         * won't block on reading from stdin.
         */

        return &empty_passwords;
    }

    /*
     * Passwords are normally allocated from the temporary pool
     * and cleared after parsing configuration.  To be used at
     * runtime they have to be copied to the configuration pool.
     */

    pwds = ngx_array_create(cf->pool, passwords->nelts, sizeof(ngx_str_t));
    if (pwds == NULL) {
        return NULL;
    }

    cln = ngx_pool_cleanup_add(cf->pool, 0);
    if (cln == NULL) {
        return NULL;
    }

    cln->handler = ngx_ssl_passwords_cleanup;
    cln->data = pwds;

    opwd = passwords->elts;

    for (i = 0; i < passwords->nelts; i++) {

        pwd = ngx_array_push(pwds);
        if (pwd == NULL) {
            return NULL;
        }

        pwd->len = opwd[i].len;
        pwd->data = ngx_pnalloc(cf->pool, pwd->len);

        if (pwd->data == NULL) {
            pwds->nelts--;
            return NULL;
        }

        ngx_memcpy(pwd->data, opwd[i].data, opwd[i].len);
    }

    return pwds;
}

ngx_int_t
ngx_ssl_trusted_certificate(ngx_conf_t *cf, ngx_ssl_t *ssl, ngx_str_t *cert,
    ngx_int_t depth)
{
    SSL_CTX_set_verify(ssl->ctx, SSL_CTX_get_verify_mode(ssl->ctx),
                       ngx_ssl_verify_callback);

    SSL_CTX_set_verify_depth(ssl->ctx, depth);

    if (cert->len == 0) {
        return NGX_OK;
    }

    if (ngx_conf_full_name(cf->cycle, cert, 1) != NGX_OK) {
        return NGX_ERROR;
    }

    if (SSL_CTX_load_verify_locations(ssl->ctx, (char *) cert->data, NULL)
        == 0)
    {
        ngx_ssl_error(NGX_LOG_EMERG, ssl->log, 0,
                      "SSL_CTX_load_verify_locations(\"%s\") failed",
                      cert->data);
        return NGX_ERROR;
    }

    /*
     * SSL_CTX_load_verify_locations() may leave errors in the error queue
     * while returning success
     */

    ERR_clear_error();

    return NGX_OK;
}


static void
ngx_ssl_info_callback(const ngx_ssl_conn_t *ssl_conn, int where, int ret)
{
    BIO               *rbio, *wbio;
    ngx_connection_t  *c;

#ifndef SSL_OP_NO_RENEGOTIATION

    if ((where & SSL_CB_HANDSHAKE_START)
        && SSL_is_server((ngx_ssl_conn_t *) ssl_conn))
    {
        c = ngx_ssl_get_connection((ngx_ssl_conn_t *) ssl_conn);

        if (c->ssl->handshaked) {
            c->ssl->renegotiation = 1;
            ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL renegotiation");
        }
    }

#endif

#ifdef TLS1_3_VERSION

    if ((where & SSL_CB_ACCEPT_LOOP) == SSL_CB_ACCEPT_LOOP
        && SSL_version(ssl_conn) == TLS1_3_VERSION)
    {
        time_t        now, time, timeout, conf_timeout;
        SSL_SESSION  *sess;

        /*
         * OpenSSL with TLSv1.3 updates the session creation time on
         * session resumption and keeps the session timeout unmodified,
         * making it possible to maintain the session forever, bypassing
         * client certificate expiration and revocation.  To make sure
         * session timeouts are actually used, we now update the session
         * creation time and reduce the session timeout accordingly.
         *
         * BoringSSL with TLSv1.3 ignores configured session timeouts
         * and uses a hardcoded timeout instead, 7 days.  So we update
         * session timeout to the configured value as soon as a session
         * is created.
         */

        c = ngx_ssl_get_connection((ngx_ssl_conn_t *) ssl_conn);
        sess = SSL_get0_session(ssl_conn);

        if (!c->ssl->session_timeout_set && sess) {
            c->ssl->session_timeout_set = 1;

            now = ngx_time();
            time = SSL_SESSION_get_time(sess);
            timeout = SSL_SESSION_get_timeout(sess);
            conf_timeout = SSL_CTX_get_timeout(c->ssl->session_ctx);

            timeout = ngx_min(timeout, conf_timeout);

            if (now - time >= timeout) {
                SSL_SESSION_set1_id_context(sess, (unsigned char *) "", 0);

            } else {
                SSL_SESSION_set_time(sess, now);
                SSL_SESSION_set_timeout(sess, timeout - (now - time));
            }
        }
    }

#endif

    if ((where & SSL_CB_ACCEPT_LOOP) == SSL_CB_ACCEPT_LOOP) {
        c = ngx_ssl_get_connection((ngx_ssl_conn_t *) ssl_conn);

        if (!c->ssl->handshake_buffer_set) {
            /*
             * By default OpenSSL uses 4k buffer during a handshake,
             * which is too low for long certificate chains and might
             * result in extra round-trips.
             *
             * To adjust a buffer size we detect that buffering was added
             * to write side of the connection by comparing rbio and wbio.
             * If they are different, we assume that it's due to buffering
             * added to wbio, and set buffer size.
             */

            rbio = SSL_get_rbio(ssl_conn);
            wbio = SSL_get_wbio(ssl_conn);

            if (rbio != wbio) {
                (void) BIO_set_write_buffer_size(wbio, NGX_SSL_BUFSIZE);
                c->ssl->handshake_buffer_set = 1;
            }
        }
    }
}

static X509 *
ngx_ssl_load_certificate(ngx_pool_t *pool, char **err, ngx_str_t *cert,
    STACK_OF(X509) **chain)
{
    BIO     *bio;
    X509    *x509, *temp;
    u_long   n;

    if (ngx_strncmp(cert->data, "data:", sizeof("data:") - 1) == 0) {

        bio = BIO_new_mem_buf(cert->data + sizeof("data:") - 1,
                              cert->len - (sizeof("data:") - 1));
        if (bio == NULL) {
            *err = "BIO_new_mem_buf() failed";
            return NULL;
        }

    } else {

        if (ngx_get_full_name(pool, (ngx_str_t *) &ngx_cycle->conf_prefix, cert)
            != NGX_OK)
        {
            *err = NULL;
            return NULL;
        }

        bio = BIO_new_file((char *) cert->data, "r");
        if (bio == NULL) {
            *err = "BIO_new_file() failed";
            return NULL;
        }
    }

    /* certificate itself */

    x509 = PEM_read_bio_X509_AUX(bio, NULL, NULL, NULL);
    if (x509 == NULL) {
        *err = "PEM_read_bio_X509_AUX() failed";
        BIO_free(bio);
        return NULL;
    }

    /* rest of the chain */

    *chain = sk_X509_new_null();
    if (*chain == NULL) {
        *err = "sk_X509_new_null() failed";
        BIO_free(bio);
        X509_free(x509);
        return NULL;
    }

    for ( ;; ) {

        temp = PEM_read_bio_X509(bio, NULL, NULL, NULL);
        if (temp == NULL) {
            n = ERR_peek_last_error();

            if (ERR_GET_LIB(n) == ERR_LIB_PEM
                && ERR_GET_REASON(n) == PEM_R_NO_START_LINE)
            {
                /* end of file */
                ERR_clear_error();
                break;
            }

            /* some real error */

            *err = "PEM_read_bio_X509() failed";
            BIO_free(bio);
            X509_free(x509);
            sk_X509_pop_free(*chain, X509_free);
            return NULL;
        }

        if (sk_X509_push(*chain, temp) == 0) {
            *err = "sk_X509_push() failed";
            BIO_free(bio);
            X509_free(x509);
            sk_X509_pop_free(*chain, X509_free);
            return NULL;
        }
    }

    BIO_free(bio);

    return x509;
}


static EVP_PKEY *
ngx_ssl_load_certificate_key(ngx_pool_t *pool, char **err,
    ngx_str_t *key, ngx_array_t *passwords)
{
    BIO              *bio;
    EVP_PKEY         *pkey;
    ngx_str_t        *pwd;
    ngx_uint_t        tries;
    pem_password_cb  *cb;

    if (ngx_strncmp(key->data, "engine:", sizeof("engine:") - 1) == 0) {

#ifndef OPENSSL_NO_ENGINE

        u_char  *p, *last;
        ENGINE  *engine;

        p = key->data + sizeof("engine:") - 1;
        last = (u_char *) ngx_strchr(p, ':');

        if (last == NULL) {
            *err = "invalid syntax";
            return NULL;
        }

        *last = '\0';

        engine = ENGINE_by_id((char *) p);

        if (engine == NULL) {
            *err = "ENGINE_by_id() failed";
            return NULL;
        }

        *last++ = ':';

        pkey = ENGINE_load_private_key(engine, (char *) last, 0, 0);

        if (pkey == NULL) {
            *err = "ENGINE_load_private_key() failed";
            ENGINE_free(engine);
            return NULL;
        }

        ENGINE_free(engine);

        return pkey;

#else

        *err = "loading \"engine:...\" certificate keys is not supported";
        return NULL;

#endif
    }

    if (ngx_strncmp(key->data, "data:", sizeof("data:") - 1) == 0) {

        bio = BIO_new_mem_buf(key->data + sizeof("data:") - 1,
                              key->len - (sizeof("data:") - 1));
        if (bio == NULL) {
            *err = "BIO_new_mem_buf() failed";
            return NULL;
        }

    } else {

        if (ngx_get_full_name(pool, (ngx_str_t *) &ngx_cycle->conf_prefix, key)
            != NGX_OK)
        {
            *err = NULL;
            return NULL;
        }

        bio = BIO_new_file((char *) key->data, "r");
        if (bio == NULL) {
            *err = "BIO_new_file() failed";
            return NULL;
        }
    }

    if (passwords) {
        tries = passwords->nelts;
        pwd = passwords->elts;
        cb = ngx_ssl_password_callback;

    } else {
        tries = 1;
        pwd = NULL;
        cb = NULL;
    }

    for ( ;; ) {

        pkey = PEM_read_bio_PrivateKey(bio, NULL, cb, pwd);
        if (pkey != NULL) {
            break;
        }

        if (tries-- > 1) {
            ERR_clear_error();
            (void) BIO_reset(bio);
            pwd++;
            continue;
        }

        *err = "PEM_read_bio_PrivateKey() failed";
        BIO_free(bio);
        return NULL;
    }

    BIO_free(bio);

    return pkey;
}

static int
ngx_ssl_new_client_session(ngx_ssl_conn_t *ssl_conn, ngx_ssl_session_t *sess)
{
    ngx_connection_t  *c;

    c = ngx_ssl_get_connection(ssl_conn);

    if (c->ssl->save_session) {
        c->ssl->session = sess;

        c->ssl->save_session(c);

        c->ssl->session = NULL;
    }

    return 0;
}

static void
ngx_ssl_passwords_cleanup(void *data)
{
    ngx_array_t *passwords = data;

    ngx_str_t   *pwd;
    ngx_uint_t   i;

    pwd = passwords->elts;

    for (i = 0; i < passwords->nelts; i++) {
        ngx_explicit_memzero(pwd[i].data, pwd[i].len);
    }
}

static int
ngx_ssl_verify_callback(int ok, X509_STORE_CTX *x509_store)
{
#if (NGX_DEBUG)
    char              *subject, *issuer;
    int                err, depth;
    X509              *cert;
    X509_NAME         *sname, *iname;
    ngx_connection_t  *c;
    ngx_ssl_conn_t    *ssl_conn;

    ssl_conn = X509_STORE_CTX_get_ex_data(x509_store,
                                          SSL_get_ex_data_X509_STORE_CTX_idx());

    c = ngx_ssl_get_connection(ssl_conn);

    if (!(c->log->log_level & NGX_LOG_DEBUG_EVENT)) {
        return 1;
    }

    cert = X509_STORE_CTX_get_current_cert(x509_store);
    err = X509_STORE_CTX_get_error(x509_store);
    depth = X509_STORE_CTX_get_error_depth(x509_store);

    sname = X509_get_subject_name(cert);

    if (sname) {
        subject = X509_NAME_oneline(sname, NULL, 0);
        if (subject == NULL) {
            ngx_ssl_error(NGX_LOG_ALERT, c->log, 0,
                          "X509_NAME_oneline() failed");
        }

    } else {
        subject = NULL;
    }

    iname = X509_get_issuer_name(cert);

    if (iname) {
        issuer = X509_NAME_oneline(iname, NULL, 0);
        if (issuer == NULL) {
            ngx_ssl_error(NGX_LOG_ALERT, c->log, 0,
                          "X509_NAME_oneline() failed");
        }

    } else {
        issuer = NULL;
    }

    ngx_log_debug5(NGX_LOG_DEBUG_EVENT, c->log, 0,
                   "verify:%d, error:%d, depth:%d, "
                   "subject:\"%s\", issuer:\"%s\"",
                   ok, err, depth,
                   subject ? subject : "(none)",
                   issuer ? issuer : "(none)");

    if (subject) {
        OPENSSL_free(subject);
    }

    if (issuer) {
        OPENSSL_free(issuer);
    }
#endif

    return 1;
}

ngx_int_t
ngx_ssl_create_connection(ngx_ssl_t *ssl, ngx_connection_t *c, ngx_uint_t flags)
{
    ngx_ssl_connection_t  *sc;

    sc = ngx_pcalloc(c->pool, sizeof(ngx_ssl_connection_t));
    if (sc == NULL) {
        return NGX_ERROR;
    }

    sc->buffer = ((flags & NGX_SSL_BUFFER) != 0);
    sc->buffer_size = ssl->buffer_size;

    sc->session_ctx = ssl->ctx;

#ifdef SSL_READ_EARLY_DATA_SUCCESS
    if (SSL_CTX_get_max_early_data(ssl->ctx)) {
        sc->try_early_data = 1;
    }
#endif

    sc->connection = SSL_new(ssl->ctx);

    if (sc->connection == NULL) {
        ngx_ssl_error(NGX_LOG_ALERT, c->log, 0, "SSL_new() failed");
        return NGX_ERROR;
    }

    if (SSL_set_fd(sc->connection, c->fd) == 0) {
        ngx_ssl_error(NGX_LOG_ALERT, c->log, 0, "SSL_set_fd() failed");
        return NGX_ERROR;
    }

    if (flags & NGX_SSL_CLIENT) {
        SSL_set_connect_state(sc->connection);

    } else {
        SSL_set_accept_state(sc->connection);

#ifdef SSL_OP_NO_RENEGOTIATION
        SSL_set_options(sc->connection, SSL_OP_NO_RENEGOTIATION);
#endif
    }

    if (SSL_set_ex_data(sc->connection, ngx_ssl_connection_index, c) == 0) {
        ngx_ssl_error(NGX_LOG_ALERT, c->log, 0, "SSL_set_ex_data() failed");
        return NGX_ERROR;
    }

    c->ssl = sc;
#if (NGX_SSL && NGX_SSL_ASYNC)
    c->async_enable = ssl->async_enable;
#endif

    return NGX_OK;
}

ngx_int_t
ngx_ssl_handshake(ngx_connection_t *c)
{
    int        n, sslerr;
    ngx_err_t  err;
    ngx_int_t  rc;

#ifdef SSL_READ_EARLY_DATA_SUCCESS
    if (c->ssl->try_early_data) {
        return ngx_ssl_try_early_data(c);
    }
#endif

#if (T_NGX_HAVE_DTLS)
    struct timeval timeout;

    if (c->type == SOCK_DGRAM
        && !c->ssl->client
        && !c->ssl->bio_changed)
    {
        rc = ngx_dtls_handshake(c);
        if (rc != NGX_OK) {
            return rc;
        }
    }
#endif

    if (c->ssl->in_ocsp) {
        return ngx_ssl_ocsp_validate(c);
    }

    ngx_ssl_clear_error(c->log);

    n = SSL_do_handshake(c->ssl->connection);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_do_handshake: %d", n);

    if (n == 1) {

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
            return NGX_ERROR;
        }

#if (T_NGX_HAVE_DTLS)
        if (c->ssl->retrans &&
            c->ssl->retrans->timer_set) {
            ngx_del_timer(c->ssl->retrans);
        }
#endif

#if (NGX_DEBUG)
        ngx_ssl_handshake_log(c);
#endif
#if (T_NGX_SSL_HANDSHAKE_TIME)
        ngx_time_t *tp;
        tp = ngx_timeofday();
        c->ssl->handshake_end_msec = tp->sec * 1000 + tp->msec;
#endif

        c->recv = ngx_ssl_recv;
        c->send = ngx_ssl_write;
        c->recv_chain = ngx_ssl_recv_chain;
        c->send_chain = ngx_ssl_send_chain;

        c->read->ready = 1;
        c->write->ready = 1;

#ifndef SSL_OP_NO_RENEGOTIATION
#if OPENSSL_VERSION_NUMBER < 0x10100000L
#ifdef SSL3_FLAGS_NO_RENEGOTIATE_CIPHERS

        /* initial handshake done, disable renegotiation (CVE-2009-3555) */
        if (c->ssl->connection->s3 && SSL_is_server(c->ssl->connection)) {
            c->ssl->connection->s3->flags |= SSL3_FLAGS_NO_RENEGOTIATE_CIPHERS;
        }

#endif
#endif
#endif

#if (defined BIO_get_ktls_send && !NGX_WIN32)

        if (BIO_get_ktls_send(SSL_get_wbio(c->ssl->connection)) == 1) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                           "BIO_get_ktls_send(): 1");
            c->ssl->sendfile = 1;
        }

#endif

        rc = ngx_ssl_ocsp_validate(c);

        if (rc == NGX_ERROR) {
            return NGX_ERROR;
        }

        if (rc == NGX_AGAIN) {
            c->read->handler = ngx_ssl_handshake_handler;
            c->write->handler = ngx_ssl_handshake_handler;
            return NGX_AGAIN;
        }

        c->ssl->handshaked = 1;

        return NGX_OK;
    }

    sslerr = SSL_get_error(c->ssl->connection, n);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_get_error: %d", sslerr);

    if (sslerr == SSL_ERROR_WANT_READ) {

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif

        c->read->ready = 0;
        c->read->handler = ngx_ssl_handshake_handler;
        c->write->handler = ngx_ssl_handshake_handler;

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
            return NGX_ERROR;
        }

#if (T_NGX_HAVE_DTLS)
        if (c->type == SOCK_DGRAM) {

        /*At least on packet should be sent(SH or HR)*/
            if (!c->ssl->dtls_send) {
                ngx_ssl_error(NGX_LOG_ERR, c->log, 0,
                              "unexcepted message of dtls session");
                return NGX_ERROR;
            }

            if (DTLSv1_get_timeout(c->ssl->connection, &timeout)) {
                ngx_msec_t timer = timeout.tv_sec * 1000 + timeout.tv_usec/1000;

                if (timer) {
                    c->ssl->retrans->handler = ngx_dtls_retransmit;
                    ngx_add_timer(c->ssl->retrans, timer);
                }

            } else {
                /*No need to retransmit, delete timer*/
                if (c->ssl->retrans &&
                    c->ssl->retrans->timer_set) {
                    ngx_del_timer(c->ssl->retrans);
                }
            }
        }
#endif

        return NGX_AGAIN;
    }

    if (sslerr == SSL_ERROR_WANT_WRITE) {

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif

        c->write->ready = 0;
        c->read->handler = ngx_ssl_handshake_handler;
        c->write->handler = ngx_ssl_handshake_handler;

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        return NGX_AGAIN;
    }

#if (NGX_SSL && NGX_SSL_ASYNC)
    if (c->async_enable && sslerr == SSL_ERROR_WANT_ASYNC)
    {
        c->async->handler = ngx_ssl_handshake_async_handler;
        c->read->saved_handler = c->read->handler;
        c->read->handler = ngx_ssl_empty_handler;

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL ASYNC WANT recieved: \"%s\"", __func__);

        if (ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }

        return NGX_AGAIN;
    }
#endif

    err = (sslerr == SSL_ERROR_SYSCALL) ? ngx_errno : 0;

    c->ssl->no_wait_shutdown = 1;
    c->ssl->no_send_shutdown = 1;
    c->read->eof = 1;

    if (sslerr == SSL_ERROR_ZERO_RETURN || ERR_peek_error() == 0) {
        ngx_connection_error(c, err,
                             "peer closed connection in SSL handshake");

        return NGX_ERROR;
    }

    if (c->ssl->handshake_rejected) {
        ngx_connection_error(c, err, "handshake rejected");
        ERR_clear_error();

        return NGX_ERROR;
    }

    c->read->error = 1;

    ngx_ssl_connection_error(c, sslerr, err, "SSL_do_handshake() failed");

    return NGX_ERROR;
}

#ifdef SSL_READ_EARLY_DATA_SUCCESS
static ngx_int_t
ngx_ssl_try_early_data(ngx_connection_t *c)
{
    int        n, sslerr;
    u_char     buf;
    size_t     readbytes;
    ngx_err_t  err;
    ngx_int_t  rc;

    ngx_ssl_clear_error(c->log);

    readbytes = 0;

    n = SSL_read_early_data(c->ssl->connection, &buf, 1, &readbytes);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                   "SSL_read_early_data: %d, %uz", n, readbytes);

    if (n == SSL_READ_EARLY_DATA_FINISH) {
        c->ssl->try_early_data = 0;
        return ngx_ssl_handshake(c);
    }

    if (n == SSL_READ_EARLY_DATA_SUCCESS) {

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
            return NGX_ERROR;
        }

#if (NGX_DEBUG)
        ngx_ssl_handshake_log(c);
#endif

        c->ssl->try_early_data = 0;

        c->ssl->early_buf = buf;
        c->ssl->early_preread = 1;

        c->ssl->in_early = 1;

        c->recv = ngx_ssl_recv;
        c->send = ngx_ssl_write;
        c->recv_chain = ngx_ssl_recv_chain;
        c->send_chain = ngx_ssl_send_chain;

        c->read->ready = 1;
        c->write->ready = 1;

#if (defined BIO_get_ktls_send && !NGX_WIN32)

        if (BIO_get_ktls_send(SSL_get_wbio(c->ssl->connection)) == 1) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                           "BIO_get_ktls_send(): 1");
            c->ssl->sendfile = 1;
        }

#endif

        rc = ngx_ssl_ocsp_validate(c);

        if (rc == NGX_ERROR) {
            return NGX_ERROR;
        }

        if (rc == NGX_AGAIN) {
            c->read->handler = ngx_ssl_handshake_handler;
            c->write->handler = ngx_ssl_handshake_handler;
            return NGX_AGAIN;
        }

        c->ssl->handshaked = 1;

        return NGX_OK;
    }

    /* SSL_READ_EARLY_DATA_ERROR */

    sslerr = SSL_get_error(c->ssl->connection, n);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_get_error: %d", sslerr);

    if (sslerr == SSL_ERROR_WANT_READ) {
        c->read->ready = 0;
        c->read->handler = ngx_ssl_handshake_handler;
        c->write->handler = ngx_ssl_handshake_handler;

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        return NGX_AGAIN;
    }

    if (sslerr == SSL_ERROR_WANT_WRITE) {
        c->write->ready = 0;
        c->read->handler = ngx_ssl_handshake_handler;
        c->write->handler = ngx_ssl_handshake_handler;

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        return NGX_AGAIN;
    }

#if (NGX_SSL && NGX_SSL_ASYNC)
    if (c->async_enable && sslerr == SSL_ERROR_WANT_ASYNC)
    {
        c->async->handler = ngx_ssl_handshake_async_handler;
        c->read->saved_handler = c->read->handler;
        c->read->handler = ngx_ssl_empty_handler;

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL ASYNC WANT recieved: \"%s\"", __func__);

        if (ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }

        return NGX_AGAIN;
    }
#endif

    err = (sslerr == SSL_ERROR_SYSCALL) ? ngx_errno : 0;

    c->ssl->no_wait_shutdown = 1;
    c->ssl->no_send_shutdown = 1;
    c->read->eof = 1;

    if (sslerr == SSL_ERROR_ZERO_RETURN || ERR_peek_error() == 0) {
        ngx_connection_error(c, err,
                             "peer closed connection in SSL handshake");

        return NGX_ERROR;
    }

    c->read->error = 1;

    ngx_ssl_connection_error(c, sslerr, err, "SSL_read_early_data() failed");

    return NGX_ERROR;
}

static ssize_t
ngx_ssl_recv_early(ngx_connection_t *c, u_char *buf, size_t size)
{
    int        n, bytes;
    size_t     readbytes;

    if (c->ssl->last == NGX_ERROR) {
        c->read->ready = 0;
        c->read->error = 1;
        return NGX_ERROR;
    }

    if (c->ssl->last == NGX_DONE) {
        c->read->ready = 0;
        c->read->eof = 1;
        return 0;
    }

    bytes = 0;

    ngx_ssl_clear_error(c->log);

    if (c->ssl->early_preread) {

        if (size == 0) {
            c->read->ready = 0;
            c->read->eof = 1;
            return 0;
        }

        *buf = c->ssl->early_buf;

        c->ssl->early_preread = 0;

        bytes = 1;
        size -= 1;
        buf += 1;
    }

    if (c->ssl->write_blocked) {
        return NGX_AGAIN;
    }

    /*
     * SSL_read_early_data() may return data in parts, so try to read
     * until SSL_read_early_data() would return no data
     */

    for ( ;; ) {

        readbytes = 0;

        n = SSL_read_early_data(c->ssl->connection, buf, size, &readbytes);

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL_read_early_data: %d, %uz", n, readbytes);

        if (n == SSL_READ_EARLY_DATA_SUCCESS) {

            c->ssl->last = ngx_ssl_handle_recv(c, 1);

            bytes += readbytes;
            size -= readbytes;

            if (size == 0) {
                c->read->ready = 1;
                return bytes;
            }

            buf += readbytes;

            continue;
        }

        if (n == SSL_READ_EARLY_DATA_FINISH) {

            c->ssl->last = ngx_ssl_handle_recv(c, 1);
            c->ssl->in_early = 0;

            if (bytes) {
                c->read->ready = 1;
                return bytes;
            }

            return ngx_ssl_recv(c, buf, size);
        }

        /* SSL_READ_EARLY_DATA_ERROR */

        c->ssl->last = ngx_ssl_handle_recv(c, 0);

        if (bytes) {
            if (c->ssl->last != NGX_AGAIN) {
                c->read->ready = 1;
            }

            return bytes;
        }

        switch (c->ssl->last) {

        case NGX_DONE:
            c->read->ready = 0;
            c->read->eof = 1;
            return 0;

        case NGX_ERROR:
            c->read->ready = 0;
            c->read->error = 1;

            /* fall through */

        case NGX_AGAIN:
            return c->ssl->last;
        }
    }
}

static ssize_t
ngx_ssl_write_early(ngx_connection_t *c, u_char *data, size_t size)
{
    int        n, sslerr;
    size_t     written;
    ngx_err_t  err;

    ngx_ssl_clear_error(c->log);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL to write: %uz", size);

    written = 0;

    n = SSL_write_early_data(c->ssl->connection, data, size, &written);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                   "SSL_write_early_data: %d, %uz", n, written);

    if (n > 0) {

        if (c->ssl->saved_read_handler) {

            c->read->handler = c->ssl->saved_read_handler;
            c->ssl->saved_read_handler = NULL;
            c->read->ready = 1;

            if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
                return NGX_ERROR;
            }

            ngx_post_event(c->read, &ngx_posted_events);
        }

        if (c->ssl->write_blocked) {
            c->ssl->write_blocked = 0;
            ngx_post_event(c->read, &ngx_posted_events);
        }

        c->sent += written;

        return written;
    }

    sslerr = SSL_get_error(c->ssl->connection, n);

    err = (sslerr == SSL_ERROR_SYSCALL) ? ngx_errno : 0;

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_get_error: %d", sslerr);

    if (sslerr == SSL_ERROR_WANT_WRITE) {

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL_write_early_data: want write");

        if (c->ssl->saved_read_handler) {

            c->read->handler = c->ssl->saved_read_handler;
            c->ssl->saved_read_handler = NULL;
            c->read->ready = 1;

            if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
                return NGX_ERROR;
            }

            ngx_post_event(c->read, &ngx_posted_events);
        }

        /*
         * OpenSSL 1.1.1a fails to handle SSL_read_early_data()
         * if an SSL_write_early_data() call blocked on writing,
         * see https://github.com/openssl/openssl/issues/7757
         */

        c->ssl->write_blocked = 1;

        c->write->ready = 0;
        return NGX_AGAIN;
    }

    if (sslerr == SSL_ERROR_WANT_READ) {

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL_write_early_data: want read");

        c->read->ready = 0;

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        /*
         * we do not set the timer because there is already
         * the write event timer
         */

        if (c->ssl->saved_read_handler == NULL) {
            c->ssl->saved_read_handler = c->read->handler;
            c->read->handler = ngx_ssl_read_handler;
        }

        return NGX_AGAIN;
    }

    c->ssl->no_wait_shutdown = 1;
    c->ssl->no_send_shutdown = 1;
    c->write->error = 1;

    ngx_ssl_connection_error(c, sslerr, err, "SSL_write_early_data() failed");

    return NGX_ERROR;
}
#endif

static int
ngx_ssl_password_callback(char *buf, int size, int rwflag, void *userdata)
{
    ngx_str_t *pwd = userdata;

    if (rwflag) {
        ngx_log_error(NGX_LOG_ALERT, ngx_cycle->log, 0,
                      "ngx_ssl_password_callback() is called for encryption");
        return 0;
    }

    if (pwd == NULL) {
        return 0;
    }

    if (pwd->len > (size_t) size) {
        ngx_log_error(NGX_LOG_ERR, ngx_cycle->log, 0,
                      "password is truncated to %d bytes", size);
    } else {
        size = pwd->len;
    }

    ngx_memcpy(buf, pwd->data, size);

    return size;
}

ngx_int_t
ngx_ssl_connection_certificate(ngx_connection_t *c, ngx_pool_t *pool,
    ngx_str_t *cert, ngx_str_t *key, ngx_array_t *passwords)
{
    char            *err;
    X509            *x509;
    EVP_PKEY        *pkey;
    STACK_OF(X509)  *chain;

    x509 = ngx_ssl_load_certificate(pool, &err, cert, &chain);
    if (x509 == NULL) {
        if (err != NULL) {
            ngx_ssl_error(NGX_LOG_ERR, c->log, 0,
                          "cannot load certificate \"%s\": %s",
                          cert->data, err);
        }

        return NGX_ERROR;
    }

    if (SSL_use_certificate(c->ssl->connection, x509) == 0) {
        ngx_ssl_error(NGX_LOG_ERR, c->log, 0,
                      "SSL_use_certificate(\"%s\") failed", cert->data);
        X509_free(x509);
        sk_X509_pop_free(chain, X509_free);
        return NGX_ERROR;
    }

    X509_free(x509);

#ifdef SSL_set0_chain

    /*
     * SSL_set0_chain() is only available in OpenSSL 1.0.2+,
     * but this function is only called via certificate callback,
     * which is only available in OpenSSL 1.0.2+ as well
     */

    if (SSL_set0_chain(c->ssl->connection, chain) == 0) {
        ngx_ssl_error(NGX_LOG_ERR, c->log, 0,
                      "SSL_set0_chain(\"%s\") failed", cert->data);
        sk_X509_pop_free(chain, X509_free);
        return NGX_ERROR;
    }

#endif

    pkey = ngx_ssl_load_certificate_key(pool, &err, key, passwords);
    if (pkey == NULL) {
        if (err != NULL) {
            ngx_ssl_error(NGX_LOG_ERR, c->log, 0,
                          "cannot load certificate key \"%s\": %s",
                          key->data, err);
        }

        return NGX_ERROR;
    }

    if (SSL_use_PrivateKey(c->ssl->connection, pkey) == 0) {
        ngx_ssl_error(NGX_LOG_ERR, c->log, 0,
                      "SSL_use_PrivateKey(\"%s\") failed", key->data);
        EVP_PKEY_free(pkey);
        return NGX_ERROR;
    }

    EVP_PKEY_free(pkey);

    return NGX_OK;
}

static void
ngx_ssl_handshake_handler(ngx_event_t *ev)
{
    ngx_connection_t  *c;

    c = ev->data;

    ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                   "SSL handshake handler: %d", ev->write);

    if (ev->timedout) {
        c->ssl->handler(c);
        return;
    }

    if (ngx_ssl_handshake(c) == NGX_AGAIN) {
        return;
    }

#if (NGX_SSL && NGX_SSL_ASYNC)
    /*
     * empty the handler of async event to avoid
     * going back to previous ssl handshake state
     */
    c->async->handler = ngx_ssl_empty_handler;
#endif
    c->ssl->handler(c);
}

static void
ngx_ssl_handshake_log(ngx_connection_t *c)
{
    char         buf[129], *s, *d;
#if OPENSSL_VERSION_NUMBER >= 0x10000000L
    const
#endif
    SSL_CIPHER  *cipher;

    if (!(c->log->log_level & NGX_LOG_DEBUG_EVENT)) {
        return;
    }

    cipher = SSL_get_current_cipher(c->ssl->connection);

    if (cipher) {
        SSL_CIPHER_description(cipher, &buf[1], 128);

        for (s = &buf[1], d = buf; *s; s++) {
            if (*s == ' ' && *d == ' ') {
                continue;
            }

            if (*s == LF || *s == CR) {
                continue;
            }

            *++d = *s;
        }

        if (*d != ' ') {
            d++;
        }

        *d = '\0';

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL: %s, cipher: \"%s\"",
                       SSL_get_version(c->ssl->connection), &buf[1]);

        if (SSL_session_reused(c->ssl->connection)) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                           "SSL reused session");
        }

    } else {
        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL no shared ciphers");
    }
}

ssize_t
ngx_ssl_recv(ngx_connection_t *c, u_char *buf, size_t size)
{
    int  n, bytes;

#ifdef SSL_READ_EARLY_DATA_SUCCESS
    if (c->ssl->in_early) {
        return ngx_ssl_recv_early(c, buf, size);
    }
#endif

    if (c->ssl->last == NGX_ERROR) {
        c->read->ready = 0;
        c->read->error = 1;
        return NGX_ERROR;
    }

    if (c->ssl->last == NGX_DONE) {
        c->read->ready = 0;
        c->read->eof = 1;
        return 0;
    }

    bytes = 0;

    ngx_ssl_clear_error(c->log);

    /*
     * SSL_read() may return data in parts, so try to read
     * until SSL_read() would return no data
     */

    for ( ;; ) {

        n = SSL_read(c->ssl->connection, buf, size);

        ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_read: %d", n);

        if (n > 0) {
            bytes += n;
        }

        c->ssl->last = ngx_ssl_handle_recv(c, n);

        if (c->ssl->last == NGX_OK) {

            size -= n;

            if (size == 0) {
                c->read->ready = 1;

                if (c->read->available >= 0) {
                    c->read->available -= bytes;

                    /*
                     * there can be data buffered at SSL layer,
                     * so we post an event to continue reading on the next
                     * iteration of the event loop
                     */

                    if (c->read->available < 0) {
                        c->read->available = 0;
                        c->read->ready = 0;

                        if (c->read->posted) {
                            ngx_delete_posted_event(c->read);
                        }

                        ngx_post_event(c->read, &ngx_posted_next_events);
                    }

                    ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                                   "SSL_read: avail:%d", c->read->available);

                } else {

#if (NGX_HAVE_FIONREAD)

                    if (ngx_socket_nread(c->fd, &c->read->available) == -1) {
                        c->read->ready = 0;
                        c->read->error = 1;
                        ngx_connection_error(c, ngx_socket_errno,
                                             ngx_socket_nread_n " failed");
                        return NGX_ERROR;
                    }

                    ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                                   "SSL_read: avail:%d", c->read->available);

#endif
                }

                return bytes;
            }

            buf += n;

            continue;
        }

        if (bytes) {
            if (c->ssl->last != NGX_AGAIN) {
                c->read->ready = 1;
            }

            return bytes;
        }

        switch (c->ssl->last) {

        case NGX_DONE:
            c->read->ready = 0;
            c->read->eof = 1;
            return 0;

        case NGX_ERROR:
            c->read->ready = 0;
            c->read->error = 1;

            /* fall through */

        case NGX_AGAIN:
            return c->ssl->last;
        }
    }
}

ssize_t
ngx_ssl_recv_chain(ngx_connection_t *c, ngx_chain_t *cl, off_t limit)
{
    u_char     *last;
    ssize_t     n, bytes, size;
    ngx_buf_t  *b;

    bytes = 0;

    b = cl->buf;
    last = b->last;

    for ( ;; ) {
        size = b->end - last;

        if (limit) {
            if (bytes >= limit) {
                return bytes;
            }

            if (bytes + size > limit) {
                size = (ssize_t) (limit - bytes);
            }
        }

        n = ngx_ssl_recv(c, last, size);

        if (n > 0) {
            last += n;
            bytes += n;

            if (!c->read->ready) {
                return bytes;
            }

            if (last == b->end) {
                cl = cl->next;

                if (cl == NULL) {
                    return bytes;
                }

                b = cl->buf;
                last = b->last;
            }

            continue;
        }

        if (bytes) {

            if (n == 0 || n == NGX_ERROR) {
                c->read->ready = 1;
            }

            return bytes;
        }

        return n;
    }
}

ssize_t
ngx_ssl_write(ngx_connection_t *c, u_char *data, size_t size)
{
    int        n, sslerr;
    ngx_err_t  err;

#ifdef SSL_READ_EARLY_DATA_SUCCESS
    if (c->ssl->in_early) {
        return ngx_ssl_write_early(c, data, size);
    }
#endif

    ngx_ssl_clear_error(c->log);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL to write: %uz", size);

    n = SSL_write(c->ssl->connection, data, size);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_write: %d", n);

    if (n > 0) {

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif

        if (c->ssl->saved_read_handler) {

            c->read->handler = c->ssl->saved_read_handler;
            c->ssl->saved_read_handler = NULL;
            c->read->ready = 1;

            if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
                return NGX_ERROR;
            }

            ngx_post_event(c->read, &ngx_posted_events);
        }

        c->sent += n;

        return n;
    }

    sslerr = SSL_get_error(c->ssl->connection, n);

    if (sslerr == SSL_ERROR_ZERO_RETURN) {

        /*
         * OpenSSL 1.1.1 fails to return SSL_ERROR_SYSCALL if an error
         * happens during SSL_write() after close_notify alert from the
         * peer, and returns SSL_ERROR_ZERO_RETURN instead,
         * https://git.openssl.org/?p=openssl.git;a=commitdiff;h=8051ab2
         */

        sslerr = SSL_ERROR_SYSCALL;
    }

    err = (sslerr == SSL_ERROR_SYSCALL) ? ngx_errno : 0;

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_get_error: %d", sslerr);

    if (sslerr == SSL_ERROR_WANT_WRITE) {
#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif

        if (c->ssl->saved_read_handler) {

            c->read->handler = c->ssl->saved_read_handler;
            c->ssl->saved_read_handler = NULL;
            c->read->ready = 1;

            if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
                return NGX_ERROR;
            }

            ngx_post_event(c->read, &ngx_posted_events);
        }

        c->write->ready = 0;
        return NGX_AGAIN;
    }

    if (sslerr == SSL_ERROR_WANT_READ) {

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL_write: want read");

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif
        c->read->ready = 0;

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        /*
         * we do not set the timer because there is already
         * the write event timer
         */

        if (c->ssl->saved_read_handler == NULL) {
            c->ssl->saved_read_handler = c->read->handler;
            c->read->handler = ngx_ssl_read_handler;
        }

        return NGX_AGAIN;
    }

#if (NGX_SSL && NGX_SSL_ASYNC)
    if (c->async_enable && sslerr == SSL_ERROR_WANT_ASYNC) {
        c->async->handler = ngx_ssl_write_async_handler;
        c->read->saved_handler = c->read->handler;
        c->read->handler = ngx_ssl_empty_handler;

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL ASYNC WANT recieved: \"%s\"", __func__);

        if (ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }

        return NGX_AGAIN;
    }
#endif
    c->ssl->no_wait_shutdown = 1;
    c->ssl->no_send_shutdown = 1;
    c->write->error = 1;

    ngx_ssl_connection_error(c, sslerr, err, "SSL_write() failed");

    return NGX_ERROR;
}

/*
 * OpenSSL has no SSL_writev() so we copy several bufs into our 16K buffer
 * before the SSL_write() call to decrease a SSL overhead.
 *
 * Besides for protocols such as HTTP it is possible to always buffer
 * the output to decrease a SSL overhead some more.
 */
ngx_chain_t *
ngx_ssl_send_chain(ngx_connection_t *c, ngx_chain_t *in, off_t limit)
{
    int           n;
    ngx_uint_t    flush;
    ssize_t       send, size, file_size;
    ngx_buf_t    *buf;
    ngx_chain_t  *cl;

    if (!c->ssl->buffer) {

        while (in) {
            if (ngx_buf_special(in->buf)) {
                in = in->next;
                continue;
            }

            n = ngx_ssl_write(c, in->buf->pos, in->buf->last - in->buf->pos);

            if (n == NGX_ERROR) {
                return NGX_CHAIN_ERROR;
            }

            if (n == NGX_AGAIN) {
                return in;
            }

            in->buf->pos += n;

            if (in->buf->pos == in->buf->last) {
                in = in->next;
            }
        }

        return in;
    }


    /* the maximum limit size is the maximum int32_t value - the page size */

    if (limit == 0 || limit > (off_t) (NGX_MAX_INT32_VALUE - ngx_pagesize)) {
        limit = NGX_MAX_INT32_VALUE - ngx_pagesize;
    }

    buf = c->ssl->buf;

    if (buf == NULL) {
        buf = ngx_create_temp_buf(c->pool, c->ssl->buffer_size);
        if (buf == NULL) {
            return NGX_CHAIN_ERROR;
        }

        c->ssl->buf = buf;
    }

    if (buf->start == NULL) {
        buf->start = ngx_palloc(c->pool, c->ssl->buffer_size);
        if (buf->start == NULL) {
            return NGX_CHAIN_ERROR;
        }

        buf->pos = buf->start;
        buf->last = buf->start;
        buf->end = buf->start + c->ssl->buffer_size;
    }

    send = buf->last - buf->pos;
    flush = (in == NULL) ? 1 : buf->flush;

    for ( ;; ) {

        while (in && buf->last < buf->end && send < limit) {
            if (in->buf->last_buf || in->buf->flush) {
                flush = 1;
            }

            if (ngx_buf_special(in->buf)) {
                in = in->next;
                continue;
            }

            if (in->buf->in_file && c->ssl->sendfile) {
                flush = 1;
                break;
            }

            size = in->buf->last - in->buf->pos;

            if (size > buf->end - buf->last) {
                size = buf->end - buf->last;
            }

            if (send + size > limit) {
                size = (ssize_t) (limit - send);
            }

            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                           "SSL buf copy: %z", size);

            ngx_memcpy(buf->last, in->buf->pos, size);

            buf->last += size;
            in->buf->pos += size;
            send += size;

            if (in->buf->pos == in->buf->last) {
                in = in->next;
            }
        }

        if (!flush && send < limit && buf->last < buf->end) {
            break;
        }

        size = buf->last - buf->pos;

        if (size == 0) {

            if (in && in->buf->in_file && send < limit) {

                /* coalesce the neighbouring file bufs */

                cl = in;
                file_size = (size_t) ngx_chain_coalesce_file(&cl, limit - send);

                n = ngx_ssl_sendfile(c, in->buf, file_size);

                if (n == NGX_ERROR) {
                    return NGX_CHAIN_ERROR;
                }

                if (n == NGX_AGAIN) {
                    break;
                }

                in = ngx_chain_update_sent(in, n);

                send += n;
                flush = 0;

                continue;
            }

            buf->flush = 0;
            c->buffered &= ~NGX_SSL_BUFFERED;

            return in;
        }

        n = ngx_ssl_write(c, buf->pos, size);

        if (n == NGX_ERROR) {
            return NGX_CHAIN_ERROR;
        }

        if (n == NGX_AGAIN) {
            break;
        }

        buf->pos += n;

        if (n < size) {
            break;
        }

        flush = 0;

        buf->pos = buf->start;
        buf->last = buf->start;

        if (in == NULL || send >= limit) {
            break;
        }
    }

    buf->flush = flush;

    if (buf->pos < buf->last) {
        c->buffered |= NGX_SSL_BUFFERED;

    } else {
        c->buffered &= ~NGX_SSL_BUFFERED;
    }

    return in;
}

ngx_int_t
ngx_ssl_check_host(ngx_connection_t *c, ngx_str_t *name)
{
    X509   *cert;

    cert = SSL_get_peer_certificate(c->ssl->connection);
    if (cert == NULL) {
        return NGX_ERROR;
    }

#ifdef X509_CHECK_FLAG_ALWAYS_CHECK_SUBJECT

    /* X509_check_host() is only available in OpenSSL 1.0.2+ */

    if (name->len == 0) {
        goto failed;
    }

    if (X509_check_host(cert, (char *) name->data, name->len, 0, NULL) != 1) {
        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "X509_check_host(): no match");
        goto failed;
    }

    ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                   "X509_check_host(): match");

    goto found;

#else
    {
    int                      n, i;
    X509_NAME               *sname;
    ASN1_STRING             *str;
    X509_NAME_ENTRY         *entry;
    GENERAL_NAME            *altname;
    STACK_OF(GENERAL_NAME)  *altnames;

    /*
     * As per RFC6125 and RFC2818, we check subjectAltName extension,
     * and if it's not present - commonName in Subject is checked.
     */

    altnames = X509_get_ext_d2i(cert, NID_subject_alt_name, NULL, NULL);

    if (altnames) {
        n = sk_GENERAL_NAME_num(altnames);

        for (i = 0; i < n; i++) {
            altname = sk_GENERAL_NAME_value(altnames, i);

            if (altname->type != GEN_DNS) {
                continue;
            }

            str = altname->d.dNSName;

            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                           "SSL subjectAltName: \"%*s\"",
                           ASN1_STRING_length(str), ASN1_STRING_data(str));

            if (ngx_ssl_check_name(name, str) == NGX_OK) {
                ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                               "SSL subjectAltName: match");
                GENERAL_NAMES_free(altnames);
                goto found;
            }
        }

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL subjectAltName: no match");

        GENERAL_NAMES_free(altnames);
        goto failed;
    }

    /*
     * If there is no subjectAltName extension, check commonName
     * in Subject.  While RFC2818 requires to only check "most specific"
     * CN, both Apache and OpenSSL check all CNs, and so do we.
     */

    sname = X509_get_subject_name(cert);

    if (sname == NULL) {
        goto failed;
    }

    i = -1;
    for ( ;; ) {
        i = X509_NAME_get_index_by_NID(sname, NID_commonName, i);

        if (i < 0) {
            break;
        }

        entry = X509_NAME_get_entry(sname, i);
        str = X509_NAME_ENTRY_get_data(entry);

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL commonName: \"%*s\"",
                       ASN1_STRING_length(str), ASN1_STRING_data(str));

        if (ngx_ssl_check_name(name, str) == NGX_OK) {
            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                           "SSL commonName: match");
            goto found;
        }
    }

    ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                   "SSL commonName: no match");
    }
#endif

failed:

    X509_free(cert);
    return NGX_ERROR;

found:

    X509_free(cert);
    return NGX_OK;
}

static ngx_int_t
ngx_ssl_handle_recv(ngx_connection_t *c, int n)
{
    int        sslerr;
    ngx_err_t  err;

#ifndef SSL_OP_NO_RENEGOTIATION

    if (c->ssl->renegotiation) {
        /*
         * disable renegotiation (CVE-2009-3555):
         * OpenSSL (at least up to 0.9.8l) does not handle disabled
         * renegotiation gracefully, so drop connection here
         */

        ngx_log_error(NGX_LOG_NOTICE, c->log, 0, "SSL renegotiation disabled");

        while (ERR_peek_error()) {
            ngx_ssl_error(NGX_LOG_DEBUG, c->log, 0,
                          "ignoring stale global SSL error");
        }

        ERR_clear_error();

        c->ssl->no_wait_shutdown = 1;
        c->ssl->no_send_shutdown = 1;

        return NGX_ERROR;
    }

#endif

    if (n > 0) {

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif

        if (c->ssl->saved_write_handler) {

            c->write->handler = c->ssl->saved_write_handler;
            c->ssl->saved_write_handler = NULL;
            c->write->ready = 1;

            if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
                return NGX_ERROR;
            }

            ngx_post_event(c->write, &ngx_posted_events);
        }

        return NGX_OK;
    }

    sslerr = SSL_get_error(c->ssl->connection, n);

    err = (sslerr == SSL_ERROR_SYSCALL) ? ngx_errno : 0;

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_get_error: %d", sslerr);

    if (sslerr == SSL_ERROR_WANT_READ) {
#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif

        if (c->ssl->saved_write_handler) {

            c->write->handler = c->ssl->saved_write_handler;
            c->ssl->saved_write_handler = NULL;
            c->write->ready = 1;

            if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
                return NGX_ERROR;
            }

            ngx_post_event(c->write, &ngx_posted_events);
        }

        c->read->ready = 0;
        return NGX_AGAIN;
    }

    if (sslerr == SSL_ERROR_WANT_WRITE) {

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL_read: want write");

#if (NGX_SSL && NGX_SSL_ASYNC)
        if (c->async_enable && ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }
#endif

        c->write->ready = 0;

        if (ngx_handle_write_event(c->write, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        /*
         * we do not set the timer because there is already the read event timer
         */

        if (c->ssl->saved_write_handler == NULL) {
            c->ssl->saved_write_handler = c->write->handler;
            c->write->handler = ngx_ssl_write_handler;
        }

        return NGX_AGAIN;
    }

#if (NGX_SSL && NGX_SSL_ASYNC)
    if (c->async_enable && sslerr == SSL_ERROR_WANT_ASYNC) {
        c->async->handler = ngx_ssl_read_async_handler;
        c->read->saved_handler = c->read->handler;
        c->read->handler = ngx_ssl_empty_handler;

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL ASYNC WANT recieved: \"%s\"", __func__);

        if (ngx_ssl_async_process_fds(c) == NGX_ERROR) {
            return NGX_ERROR;
        }

        return NGX_AGAIN;
    }
#endif
    c->ssl->no_wait_shutdown = 1;
    c->ssl->no_send_shutdown = 1;

    if (sslerr == SSL_ERROR_ZERO_RETURN || ERR_peek_error() == 0) {
        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "peer shutdown SSL cleanly");
        return NGX_DONE;
    }

    ngx_ssl_connection_error(c, sslerr, err, "SSL_read() failed");

    return NGX_ERROR;
}


static void
ngx_ssl_write_handler(ngx_event_t *wev)
{
    ngx_connection_t  *c;

    c = wev->data;

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL write handler");

    c->read->handler(c->read);
}

static void
ngx_ssl_read_handler(ngx_event_t *rev)
{
    ngx_connection_t  *c;

    c = rev->data;

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL read handler");

    c->write->handler(c->write);
}

static ssize_t
ngx_ssl_sendfile(ngx_connection_t *c, ngx_buf_t *file, size_t size)
{
#if (defined BIO_get_ktls_send && !NGX_WIN32)

    int        sslerr, flags;
    ssize_t    n;
    ngx_err_t  err;

    ngx_ssl_clear_error(c->log);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                   "SSL to sendfile: @%O %uz",
                   file->file_pos, size);

    ngx_set_errno(0);

#if (NGX_HAVE_SENDFILE_NODISKIO)

    flags = (c->busy_count <= 2) ? SF_NODISKIO : 0;

    if (file->file->directio) {
        flags |= SF_NOCACHE;
    }

#else
    flags = 0;
#endif

    n = SSL_sendfile(c->ssl->connection, file->file->fd, file->file_pos,
                     size, flags);

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_sendfile: %z", n);

    if (n > 0) {

        if (c->ssl->saved_read_handler) {

            c->read->handler = c->ssl->saved_read_handler;
            c->ssl->saved_read_handler = NULL;
            c->read->ready = 1;

            if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
                return NGX_ERROR;
            }

            ngx_post_event(c->read, &ngx_posted_events);
        }

#if (NGX_HAVE_SENDFILE_NODISKIO)
        c->busy_count = 0;
#endif

        c->sent += n;

        return n;
    }

    if (n == 0) {

        /*
         * if sendfile returns zero, then someone has truncated the file,
         * so the offset became beyond the end of the file
         */

        ngx_log_error(NGX_LOG_ALERT, c->log, 0,
                      "SSL_sendfile() reported that \"%s\" was truncated at %O",
                      file->file->name.data, file->file_pos);

        return NGX_ERROR;
    }

    sslerr = SSL_get_error(c->ssl->connection, n);

    if (sslerr == SSL_ERROR_ZERO_RETURN) {

        /*
         * OpenSSL fails to return SSL_ERROR_SYSCALL if an error
         * happens during writing after close_notify alert from the
         * peer, and returns SSL_ERROR_ZERO_RETURN instead
         */

        sslerr = SSL_ERROR_SYSCALL;
    }

    if (sslerr == SSL_ERROR_SSL
        && ERR_GET_REASON(ERR_peek_error()) == SSL_R_UNINITIALIZED
        && ngx_errno != 0)
    {
        /*
         * OpenSSL fails to return SSL_ERROR_SYSCALL if an error
         * happens in sendfile(), and returns SSL_ERROR_SSL with
         * SSL_R_UNINITIALIZED reason instead
         */

        sslerr = SSL_ERROR_SYSCALL;
    }

    err = (sslerr == SSL_ERROR_SYSCALL) ? ngx_errno : 0;

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "SSL_get_error: %d", sslerr);

    if (sslerr == SSL_ERROR_WANT_WRITE) {

        if (c->ssl->saved_read_handler) {

            c->read->handler = c->ssl->saved_read_handler;
            c->ssl->saved_read_handler = NULL;
            c->read->ready = 1;

            if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
                return NGX_ERROR;
            }

            ngx_post_event(c->read, &ngx_posted_events);
        }

#if (NGX_HAVE_SENDFILE_NODISKIO)

        if (ngx_errno == EBUSY) {
            c->busy_count++;

            ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                           "SSL_sendfile() busy, count:%d", c->busy_count);

            if (c->write->posted) {
                ngx_delete_posted_event(c->write);
            }

            ngx_post_event(c->write, &ngx_posted_next_events);
        }

#endif

        c->write->ready = 0;
        return NGX_AGAIN;
    }

    if (sslerr == SSL_ERROR_WANT_READ) {

        ngx_log_error(NGX_LOG_STDERR, c->log, 0,
                       "SSL_sendfile: want read");

        c->read->ready = 0;

        if (ngx_handle_read_event(c->read, 0) != NGX_OK) {
            return NGX_ERROR;
        }

        /*
         * we do not set the timer because there is already
         * the write event timer
         */

        if (c->ssl->saved_read_handler == NULL) {
            c->ssl->saved_read_handler = c->read->handler;
            c->read->handler = ngx_ssl_read_handler;
        }

        return NGX_AGAIN;
    }

    c->ssl->no_wait_shutdown = 1;
    c->ssl->no_send_shutdown = 1;
    c->write->error = 1;

    ngx_ssl_connection_error(c, sslerr, err, "SSL_sendfile() failed");

#else
    ngx_log_error(NGX_LOG_ALERT, c->log, 0,
                  "SSL_sendfile() not available");
#endif

    return NGX_ERROR;
}

ngx_ssl_session_t *
ngx_ssl_get_session(ngx_connection_t *c)
{
#ifdef TLS1_3_VERSION
    if (c->ssl->session) {
        SSL_SESSION_up_ref(c->ssl->session);
        return c->ssl->session;
    }
#endif

    return SSL_get1_session(c->ssl->connection);
}

ngx_ssl_session_t *
ngx_ssl_get0_session(ngx_connection_t *c)
{
    if (c->ssl->session) {
        return c->ssl->session;
    }

    return SSL_get0_session(c->ssl->connection);
}

ngx_int_t
ngx_ssl_set_session(ngx_connection_t *c, ngx_ssl_session_t *session)
{
    if (session) {
        if (SSL_set_session(c->ssl->connection, session) == 0) {
            ngx_ssl_error(NGX_LOG_ALERT, c->log, 0, "SSL_set_session() failed");
            return NGX_ERROR;
        }
    }

    return NGX_OK;
}


#if defined(T_NGX_HAVE_DTLS)
void ngx_dtls_retransmit(ngx_event_t *ev)
{
    ngx_connection_t *c = ev->data;

    ngx_ssl_error(NGX_LOG_DEBUG, c->log, 0,
                  "ngx_dtls_retransmit %i", (ngx_int_t)c->ssl->retrans_times);

    if (c->ssl->retrans_times > NGX_DTLS_MAX_RETRANSMIT_TIME) {
        c->ssl->handler(c);
        return;
    }

    c->ssl->retrans_times++;
    if (ngx_ssl_handshake(c) == NGX_AGAIN) {
        return;
    }

    c->ssl->handler(c);
}

/*
 * RFC 6347, 4.2.1:
 *
 * When responding to a HelloVerifyRequest, the client MUST use the same
 * parameter values (version, random, session_id, cipher_suites,
 * compression_method) as it did in the original ClientHello.  The
 * server SHOULD use those values to generate its cookie and verify that
 * they are correct upon cookie receipt.
 */

static int
ngx_dtls_client_hmac(SSL *ssl, u_char res[EVP_MAX_MD_SIZE], unsigned int *rlen)
{
    u_char            *p;
    size_t             len;
    ngx_connection_t  *c;

    u_char             buffer[64];

    c = ngx_ssl_get_connection(ssl);

    p = buffer;

    p = ngx_cpymem(p, c->addr_text.data, c->addr_text.len);
    p = ngx_sprintf(p, "%d", ngx_inet_get_port(c->sockaddr));

    len = p - buffer;

    HMAC(EVP_sha1(), (const void*) c->ssl->dtls_cookie_secret,
         sizeof(c->ssl->dtls_cookie_secret), (const u_char*) buffer, len, res, rlen);

    return NGX_OK;
}


static int
ngx_dtls_generate_cookie_cb(SSL *ssl, unsigned char *cookie,
    unsigned int *cookie_len)
{
    unsigned int  rlen;
    u_char        res[EVP_MAX_MD_SIZE];

    if (ngx_dtls_client_hmac(ssl, res, &rlen) != NGX_OK) {
        return 0;
    }

    ngx_memcpy(cookie, res, rlen);
    *cookie_len = rlen;

    return 1;
}


static int
ngx_dtls_verify_cookie_cb(SSL *ssl,
#if OPENSSL_VERSION_NUMBER >= 0x10100000L
    const
#endif
    unsigned char *cookie, unsigned int cookie_len)
{
    unsigned int  rlen;
    u_char        res[EVP_MAX_MD_SIZE];

    if (ngx_dtls_client_hmac(ssl, res, &rlen) != NGX_OK) {
        return 0;
    }

    if (cookie_len == rlen && ngx_memcmp(res, cookie, rlen) == 0) {
        return 1;
    }

    return 0;
}

static int
ngx_dtls_new(BIO *bi)
{

#if OPENSSL_VERSION_NUMBER >= 0x10100000L
    BIO_set_init(bi, 0);
    BIO_set_data(bi, NULL);
    BIO_clear_flags(bi, ~0);
#else
    bi->init = 0;
    bi->num = 0;
    bi->ptr = NULL;
    bi->flags = 0;
#endif
    return (1);
}

static int
ngx_dtls_free(BIO *a)
{
    int shutdown, init, num;

    if (a == NULL)
        return (0);

#if OPENSSL_VERSION_NUMBER >= 0x10100000L
    ngx_connection_t *c;
    shutdown = BIO_get_shutdown(a);
    init = BIO_get_init(a);
    c = BIO_get_data(a);
    num = c->fd;
#else
    shutdown = a->shutdown;
    init = a->init;
    num = a->num;
#endif
    //
    if (shutdown) {
        //BIO_get_init
        if (init) {
            ngx_close_socket(num);
        }
#if OPENSSL_VERSION_NUMBER >= 0x10100000L
        BIO_set_data(a, NULL);
        BIO_set_init(a, 0);
#else
        a->init = 0;
        a->flags = 0;
#endif
    }

    return (1);
}

static int
ngx_dtls_read(BIO *b, char *out, int outl)
{
    ssize_t ret = 0, n = 0;

#if OPENSSL_VERSION_NUMBER >= 0x10100000L
    ngx_connection_t *c = BIO_get_data(b);
#else
    ngx_connection_t *c = (ngx_connection_t *)b->ptr;
#endif

    if (out != NULL) {

        if (c->buffer && c->buffer->pos < c->buffer->last) {
            n = c->buffer->last - c->buffer->pos;

            if (n > (ssize_t)outl) {
                n = outl;
            }
            ngx_memcpy(out, c->buffer->pos, n);
            c->buffer->pos += n;
        }

        /*ngx_udp_shared_recv*/
        ret = c->recv(c, (u_char*)out, (size_t)outl);

        BIO_clear_retry_flags(b);

        if (ret == NGX_AGAIN) {
            BIO_set_retry_read(b);

        } else {
            n += ret;
        }

        if (n == 0) {
            n = -1;
        }
    }

    return n;
}

static int
ngx_dtls_write(BIO *b, const char *in, int inl)
{
    ssize_t ret;

#if OPENSSL_VERSION_NUMBER >= 0x10100000L
    ngx_connection_t *c = BIO_get_data(b);
#else
    ngx_connection_t *c = (ngx_connection_t *)b->ptr;
#endif

    c->ssl->dtls_send = 1;
    ret = ngx_udp_send(c, (u_char *)in, (size_t)inl);

    BIO_clear_retry_flags(b);

    if (ret <= 0) {
        if (ret == NGX_AGAIN) {
            BIO_set_retry_write(b);
        }
    }

    return (ret);
}

static int
ngx_dtls_puts(BIO *bp, const char *str)
{
    return ngx_dtls_write(bp, str, strlen(str));
}

static long
ngx_dtls_ctrl(BIO *b, int cmd, long num, void *ptr)
{
    long ret = 1;
    ngx_connection_t *c;

    /*BIO_get_data*/

    switch (cmd) {
    case BIO_C_SET_CONNECT:
#if OPENSSL_VERSION_NUMBER >= 0x10100000L
        BIO_set_data(b, ptr);
        BIO_set_init(b, 1);
        BIO_set_shutdown(b, num);
        (void)c;
#else
        b->ptr = ptr;
        c = (ngx_connection_t *)ptr;
        b->num = c->fd;
        b->init = 1;
        b->shutdown = (int)num;
#endif
        break;
    case BIO_CTRL_GET_CLOSE:
#if OPENSSL_VERSION_NUMBER >= 0x10100000L
        ret = BIO_get_shutdown(b);
#else
        ret = b->shutdown;
#endif
        break;
    case BIO_CTRL_SET_CLOSE:
#if OPENSSL_VERSION_NUMBER >= 0x10100000L
        BIO_set_shutdown(b, num);
#else
        b->shutdown = (int)num;
#endif
        break;
    case BIO_CTRL_DUP:
    case BIO_CTRL_FLUSH:
        ret = 1;
        break;
    default:
        ret = 0;
        break;
    }
    return (ret);
}

#if OPENSSL_VERSION_NUMBER >= 0x10100000L
static BIO_METHOD *methods_ngx_dtls = NULL;
#else
static BIO_METHOD methods_ngx_dtls = {
    BIO_TYPE_SOCKET,
    "ngx dtls",
    ngx_dtls_write,
    ngx_dtls_read,
    ngx_dtls_puts,
    NULL,
    ngx_dtls_ctrl,
    ngx_dtls_new,
    ngx_dtls_free,
    NULL,
};

#endif
static ngx_int_t
ngx_dtls_handshake(ngx_connection_t *c)
{
    BIO *rbio;

#if OPENSSL_VERSION_NUMBER >= 0x10100000L
    if (methods_ngx_dtls == NULL) {
        methods_ngx_dtls = BIO_meth_new(BIO_TYPE_SOCKET, "ngx dtls");
        if (!methods_ngx_dtls
            || !BIO_meth_set_create(methods_ngx_dtls, ngx_dtls_new)
            || !BIO_meth_set_destroy(methods_ngx_dtls, ngx_dtls_free)
            || !BIO_meth_set_write(methods_ngx_dtls, ngx_dtls_write)
            || !BIO_meth_set_read(methods_ngx_dtls, ngx_dtls_read)
            || !BIO_meth_set_puts(methods_ngx_dtls, ngx_dtls_puts)
            || !BIO_meth_set_ctrl(methods_ngx_dtls, ngx_dtls_ctrl)) {

            return NGX_ERROR;
        }
    }
    rbio = BIO_new(methods_ngx_dtls);
#else
    /*Need to replace rbio*/
    rbio = BIO_new(&methods_ngx_dtls);
#endif

    if (rbio == NULL) {
        return NGX_ERROR;
    }

    BIO_ctrl(rbio, BIO_C_SET_CONNECT, BIO_NOCLOSE, (char *)c);

    SSL_set_bio(c->ssl->connection, rbio, rbio);
    /*TODO: directive*/
    SSL_set_options(c->ssl->connection, SSL_OP_COOKIE_EXCHANGE);

    c->ssl->retrans = ngx_pcalloc(c->pool, sizeof(ngx_event_t));
    if (!c->ssl->retrans) {
        return NGX_ERROR;
    }

    c->ssl->retrans->log = c->log;
    c->ssl->retrans->data = c;

    c->ssl->bio_changed = 1;

    if (!RAND_bytes(c->ssl->dtls_cookie_secret, sizeof(c->ssl->dtls_cookie_secret))) {
        return NGX_ERROR;
    }
    return NGX_OK;
}

#endif





