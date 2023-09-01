
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




ngx_int_t
ngx_ssl_shutdown(ngx_connection_t *c)
{
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

        ngx_log_debug1(NGX_LOG_DEBUG_EVENT, c->log, 0, "SSL_shutdown: %d", n);

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

        ngx_log_debug1(NGX_LOG_DEBUG_EVENT, c->log, 0,
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

            ngx_log_debug1(NGX_LOG_DEBUG_EVENT, c->log, 0,
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

    ngx_log_debug0(NGX_LOG_DEBUG_EVENT, ev->log, 0, "SSL shutdown handler");

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



