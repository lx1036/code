


#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_event.h>



struct ngx_udp_connection_s {
    ngx_rbtree_node_t   node;
    ngx_connection_t   *connection;
    ngx_buf_t          *buffer;
};


static void ngx_close_accepted_udp_connection(ngx_connection_t *c);
static ssize_t ngx_udp_shared_recv(ngx_connection_t *c, u_char *buf,
    size_t size);
static ngx_int_t ngx_insert_udp_connection(ngx_connection_t *c);
static ngx_connection_t *ngx_lookup_udp_connection(ngx_listening_t *ls,
    struct sockaddr *sockaddr, socklen_t socklen,
    struct sockaddr *local_sockaddr, socklen_t local_socklen);



void ngx_event_recvmsg(ngx_event_t *ev) {
    ssize_t            n;
    ngx_buf_t          buf;
    ngx_log_t         *log;
    ngx_err_t          err;
    socklen_t          socklen, local_socklen;
    ngx_event_t       *rev, *wev;
    struct iovec       iov[1];
    struct msghdr      msg;
    ngx_sockaddr_t     sa, lsa;
    struct sockaddr   *sockaddr, *local_sockaddr;
    ngx_listening_t   *ls;
    ngx_event_conf_t  *ecf;
    ngx_connection_t  *c, *lc;
    static u_char      buffer[65535];

#if (NGX_HAVE_ADDRINFO_CMSG)
    u_char             msg_control[CMSG_SPACE(sizeof(ngx_addrinfo_t))];
#endif

    if (ev->timedout) {
        if (ngx_enable_accept_events((ngx_cycle_t *) ngx_cycle) != NGX_OK) {
            return;
        }

        ev->timedout = 0;
    }

    ecf = ngx_event_get_conf(ngx_cycle->conf_ctx, ngx_event_core_module);

    if (!(ngx_event_flags & NGX_USE_KQUEUE_EVENT)) {
        ev->available = ecf->multi_accept;
    }

    lc = ev->data;
    ls = lc->listening;
    ev->ready = 0;

    ngx_log_debug2(NGX_LOG_DEBUG_EVENT, ev->log, 0,
                   "recvmsg on %V, ready: %d", &ls->addr_text, ev->available);

    do {
        ngx_memzero(&msg, sizeof(struct msghdr));

        iov[0].iov_base = (void *) buffer;
        iov[0].iov_len = sizeof(buffer);

        msg.msg_name = &sa;
        msg.msg_namelen = sizeof(ngx_sockaddr_t);
        msg.msg_iov = iov;
        msg.msg_iovlen = 1;

#if (NGX_HAVE_ADDRINFO_CMSG)
        if (ls->wildcard) {
            msg.msg_control = &msg_control;
            msg.msg_controllen = sizeof(msg_control);

            ngx_memzero(&msg_control, sizeof(msg_control));
        }
#endif

        n = recvmsg(lc->fd, &msg, 0);
        if (n == -1) {
            err = ngx_socket_errno;

            if (err == NGX_EAGAIN) {
                ngx_log_debug0(NGX_LOG_DEBUG_EVENT, ev->log, err,
                               "recvmsg() not ready");
                return;
            }

            ngx_log_error(NGX_LOG_ALERT, ev->log, err, "recvmsg() failed");

            return;
        }

#if (NGX_HAVE_ADDRINFO_CMSG)
        if (msg.msg_flags & (MSG_TRUNC|MSG_CTRUNC)) {
            ngx_log_error(NGX_LOG_ALERT, ev->log, 0,
                          "recvmsg() truncated data");
            continue;
        }
#endif

        sockaddr = msg.msg_name;
        socklen = msg.msg_namelen;

        if (socklen > (socklen_t) sizeof(ngx_sockaddr_t)) {
            socklen = sizeof(ngx_sockaddr_t);
        }

        if (socklen == 0) {

            /*
             * on Linux recvmsg() returns zero msg_namelen
             * when receiving packets from unbound AF_UNIX sockets
             */

            socklen = sizeof(struct sockaddr);
            ngx_memzero(&sa, sizeof(struct sockaddr));
            sa.sockaddr.sa_family = ls->sockaddr->sa_family;
        }

        local_sockaddr = ls->sockaddr;
        local_socklen = ls->socklen;

#if (NGX_HAVE_ADDRINFO_CMSG)

        if (ls->wildcard) {
            struct cmsghdr  *cmsg;

            ngx_memcpy(&lsa, local_sockaddr, local_socklen);
            local_sockaddr = &lsa.sockaddr;

            for (cmsg = CMSG_FIRSTHDR(&msg);
                 cmsg != NULL;
                 cmsg = CMSG_NXTHDR(&msg, cmsg))
            {
                if (ngx_get_srcaddr_cmsg(cmsg, local_sockaddr) == NGX_OK) {
                    break;
                }
            }
        }

#endif

        c = ngx_lookup_udp_connection(ls, sockaddr, socklen, local_sockaddr,
                                      local_socklen);

        if (c) {

#if (NGX_DEBUG)
            if (c->log->log_level & NGX_LOG_DEBUG_EVENT) {
                ngx_log_handler_pt  handler;

                handler = c->log->handler;
                c->log->handler = NULL;

                ngx_log_debug2(NGX_LOG_DEBUG_EVENT, c->log, 0,
                               "recvmsg: fd:%d n:%z", c->fd, n);

                c->log->handler = handler;
            }
#endif

            ngx_memzero(&buf, sizeof(ngx_buf_t));

            buf.pos = buffer;
            buf.last = buffer + n;

            rev = c->read;

            c->udp->buffer = &buf;

            rev->ready = 1;
            rev->active = 0;

            rev->handler(rev);

            if (c->udp) {
                c->udp->buffer = NULL;
            }

            rev->ready = 0;
            rev->active = 1;

            goto next;
        }

#if (NGX_STAT_STUB)
        (void) ngx_atomic_fetch_add(ngx_stat_accepted, 1);
#endif

        ngx_accept_disabled = ngx_cycle->connection_n / 8
                              - ngx_cycle->free_connection_n;

        c = ngx_get_connection(lc->fd, ev->log);
        if (c == NULL) {
            return;
        }

        c->shared = 1;
        c->type = SOCK_DGRAM;
        c->socklen = socklen;

#if (NGX_STAT_STUB)
        (void) ngx_atomic_fetch_add(ngx_stat_active, 1);
#endif

        c->pool = ngx_create_pool(ls->pool_size, ev->log);
        if (c->pool == NULL) {
            ngx_close_accepted_udp_connection(c);
            return;
        }

        c->sockaddr = ngx_palloc(c->pool, socklen);
        if (c->sockaddr == NULL) {
            ngx_close_accepted_udp_connection(c);
            return;
        }

        ngx_memcpy(c->sockaddr, sockaddr, socklen);

        log = ngx_palloc(c->pool, sizeof(ngx_log_t));
        if (log == NULL) {
            ngx_close_accepted_udp_connection(c);
            return;
        }

        *log = ls->log;

        c->recv = ngx_udp_shared_recv;
        c->send = ngx_udp_send;
        c->send_chain = ngx_udp_send_chain;

        c->need_flush_buf = 1;

        c->log = log;
        c->pool->log = log;
        c->listening = ls;

        if (local_sockaddr == &lsa.sockaddr) {
            local_sockaddr = ngx_palloc(c->pool, local_socklen);
            if (local_sockaddr == NULL) {
                ngx_close_accepted_udp_connection(c);
                return;
            }

            ngx_memcpy(local_sockaddr, &lsa, local_socklen);
        }

        c->local_sockaddr = local_sockaddr;
        c->local_socklen = local_socklen;

        c->buffer = ngx_create_temp_buf(c->pool, n);
        if (c->buffer == NULL) {
            ngx_close_accepted_udp_connection(c);
            return;
        }

        c->buffer->last = ngx_cpymem(c->buffer->last, buffer, n);

        rev = c->read;
        wev = c->write;

        rev->active = 1;
        wev->ready = 1;

        rev->log = log;
        wev->log = log;

        /*
         * TODO: MT: - ngx_atomic_fetch_add()
         *             or protection by critical section or light mutex
         *
         * TODO: MP: - allocated in a shared memory
         *           - ngx_atomic_fetch_add()
         *             or protection by critical section or light mutex
         */

        c->number = ngx_atomic_fetch_add(ngx_connection_counter, 1);

        c->start_time = ngx_current_msec;

#if (NGX_STAT_STUB)
        (void) ngx_atomic_fetch_add(ngx_stat_handled, 1);
#endif

        if (ls->addr_ntop) {
            c->addr_text.data = ngx_pnalloc(c->pool, ls->addr_text_max_len);
            if (c->addr_text.data == NULL) {
                ngx_close_accepted_udp_connection(c);
                return;
            }

            c->addr_text.len = ngx_sock_ntop(c->sockaddr, c->socklen,
                                             c->addr_text.data,
                                             ls->addr_text_max_len, 0);
            if (c->addr_text.len == 0) {
                ngx_close_accepted_udp_connection(c);
                return;
            }
        }

#if (NGX_DEBUG)
        {
        ngx_str_t  addr;
        u_char     text[NGX_SOCKADDR_STRLEN];

        ngx_debug_accepted_connection(ecf, c);

        if (log->log_level & NGX_LOG_DEBUG_EVENT) {
            addr.data = text;
            addr.len = ngx_sock_ntop(c->sockaddr, c->socklen, text,
                                     NGX_SOCKADDR_STRLEN, 1);

            ngx_log_debug4(NGX_LOG_DEBUG_EVENT, log, 0,
                           "*%uA recvmsg: %V fd:%d n:%z",
                           c->number, &addr, c->fd, n);
        }

        }
#endif

        if (ngx_insert_udp_connection(c) != NGX_OK) {
            ngx_close_accepted_udp_connection(c);
            return;
        }

        log->data = NULL;
        log->handler = NULL;

        ls->handler(c);

    next:

        if (ngx_event_flags & NGX_USE_KQUEUE_EVENT) {
            ev->available -= n;
        }

    } while (ev->available);
}

static void
ngx_close_accepted_udp_connection(ngx_connection_t *c)
{
    ngx_free_connection(c);
    c->fd = (ngx_socket_t) -1;

    if (c->pool) {
        ngx_destroy_pool(c->pool);
    }

#if (NGX_STAT_STUB)
    (void) ngx_atomic_fetch_add(ngx_stat_active, -1);
#endif
}

