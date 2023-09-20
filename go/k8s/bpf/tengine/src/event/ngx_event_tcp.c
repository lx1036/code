


#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_event.h>


static ngx_int_t ngx_disable_accept_events(ngx_cycle_t *cycle, ngx_uint_t all);
#if (NGX_HAVE_EPOLLEXCLUSIVE)
static void ngx_reorder_accept_events(ngx_listening_t *ls);
#endif
static void ngx_close_accepted_connection(ngx_connection_t *c);

ngx_int_t ngx_trylock_accept_mutex(ngx_cycle_t *cycle) {
    if (ngx_shmtx_trylock(&ngx_accept_mutex)) {
        ngx_log_error(NGX_LOG_STDERR, cycle->log, 0, "accept mutex locked");
        if (ngx_accept_mutex_held && ngx_accept_events == 0) {
            return NGX_OK;
        }

        if (ngx_enable_accept_events(cycle) == NGX_ERROR) {
            ngx_shmtx_unlock(&ngx_accept_mutex);
            return NGX_ERROR;
        }

        ngx_accept_events = 0;
        ngx_accept_mutex_held = 1;

        return NGX_OK;
    }

    ngx_log_error(NGX_LOG_STDERR, cycle->log, 0, "accept mutex lock failed: %ui", ngx_accept_mutex_held);
    if (ngx_accept_mutex_held) {
        if (ngx_disable_accept_events(cycle, 0) == NGX_ERROR) {
            return NGX_ERROR;
        }

        ngx_accept_mutex_held = 0;
    }

    return NGX_OK;
}

// 当前 connection 没法去 read event
ngx_int_t ngx_enable_accept_events(ngx_cycle_t *cycle){
    ngx_uint_t         i;
    ngx_listening_t   *ls;
    ngx_connection_t  *c;

    ls = cycle->listening.elts;
    for (i = 0; i < cycle->listening.nelts; i++) {
        c = ls[i].connection;
        if (c == NULL || c->read->active) {
            continue;
        }

        if (ngx_add_event(c->read, NGX_READ_EVENT, 0) == NGX_ERROR) {
            return NGX_ERROR;
        }
    }

    return NGX_OK;
}

static ngx_int_t ngx_disable_accept_events(ngx_cycle_t *cycle, ngx_uint_t all) {
    ngx_uint_t         i;
    ngx_listening_t   *ls;
    ngx_connection_t  *c;

    ls = cycle->listening.elts;
    for (i = 0; i < cycle->listening.nelts; i++) {
        c = ls[i].connection;
        if (c == NULL || !c->read->active) {
            continue;
        }

#if (NGX_HAVE_REUSEPORT)
        /*
         * do not disable accept on worker's own sockets
         * when disabling accept events due to accept mutex
         */

        if (ls[i].reuseport && !all) {
            continue;
        }
#endif

        if (ngx_del_event(c->read, NGX_READ_EVENT, NGX_DISABLE_EVENT)
            == NGX_ERROR) {
            return NGX_ERROR;
        }
    }

    return NGX_OK;
}

void ngx_event_accept(ngx_event_t *ev) {
    socklen_t          socklen;
    ngx_err_t          err;
    ngx_log_t         *log;
    ngx_uint_t         level;
    ngx_socket_t       s;
    ngx_event_t       *rev, *wev;
    ngx_sockaddr_t     sa;
    ngx_listening_t   *ls;
    ngx_connection_t  *c, *lc;
    ngx_event_conf_t  *ecf;

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
    ngx_log_error(NGX_LOG_STDERR, ev->log, 0, "accept on %V, ready: %d", &ls->addr_text, ev->available);

    do {
        // accept a connection on a socket, return a file descriptor for the accepted socket
        // https://man7.org/linux/man-pages/man2/accept.2.html
        socklen = sizeof(ngx_sockaddr_t);
        s = accept(lc->fd, &sa.sockaddr, &socklen); // &sa.sockaddr 是客户端地址
        if (s == (ngx_socket_t) -1) {
            err = ngx_socket_errno;
            if (err == NGX_EAGAIN) {
                ngx_log_error(NGX_LOG_STDERR, ev->log, err, "accept() not ready");
                return;
            }

            level = NGX_LOG_ALERT;
            if (err == NGX_ECONNABORTED) {
                level = NGX_LOG_ERR;

            } else if (err == NGX_EMFILE || err == NGX_ENFILE) {
                level = NGX_LOG_CRIT;
            }

            ngx_log_error(level, ev->log, err, "accept() failed");
            if (err == NGX_ECONNABORTED) {
                if (ngx_event_flags & NGX_USE_KQUEUE_EVENT) {
                    ev->available--;
                }
                if (ev->available) {
                    continue;
                }
            }

            if (err == NGX_EMFILE || err == NGX_ENFILE) {
                if (ngx_disable_accept_events((ngx_cycle_t *) ngx_cycle, 1) != NGX_OK) {
                    return;
                }

                if (ngx_use_accept_mutex) {
                    if (ngx_accept_mutex_held) {
                        ngx_shmtx_unlock(&ngx_accept_mutex);
                        ngx_accept_mutex_held = 0;
                    }

                    ngx_accept_disabled = 1;
                } else {
                    ngx_add_timer(ev, ecf->accept_mutex_delay);
                }
            }

            return;
        }

#if (NGX_STAT_STUB)
        (void) ngx_atomic_fetch_add(ngx_stat_accepted, 1); // accept() 一个 client_fd
#endif
        // 重要: 使用 ngx_log_error(NGX_LOG_STDERR, xxx) 来打印日志!!!
        u_char text[NGX_SOCKADDR_STRLEN];
        ngx_inet_get_addr(&sa.sockaddr, text);
        ngx_log_error(NGX_LOG_STDERR, ev->log, 0, "accept() success from client: %s:%d to server: %s ", text, ngx_inet_get_port(&sa.sockaddr), ls->addr_text.data);
        // printf("accept() success from client: %s:%d to server: %s \n", text, ngx_inet_get_port(&sa.sockaddr), ls->addr_text.data);
        ngx_accept_disabled = ngx_cycle->connection_n / 8 - ngx_cycle->free_connection_n;
        c = ngx_get_connection(s, ev->log);
        if (c == NULL) {
            if (ngx_close_socket(s) == -1) {
                ngx_log_error(NGX_LOG_ALERT, ev->log, ngx_socket_errno, ngx_close_socket_n " failed");
            }

            return;
        }

        c->type = SOCK_STREAM;

#if (NGX_STAT_STUB)
        (void) ngx_atomic_fetch_add(ngx_stat_active, 1);
#endif

        // 继续初始化 connection 结构体
        c->pool = ngx_create_pool(ls->pool_size, ev->log);
        if (c->pool == NULL) {
            ngx_close_accepted_connection(c);
            return;
        }
        if (socklen > (socklen_t) sizeof(ngx_sockaddr_t)) {
            socklen = sizeof(ngx_sockaddr_t);
        }
        c->sockaddr = ngx_palloc(c->pool, socklen);
        if (c->sockaddr == NULL) {
            ngx_close_accepted_connection(c);
            return;
        }
        // 这里 socklen 是 type(ngx_sockaddr_t) 的前部分字节，就是 ngx_sockaddr_t.sockaddr, 这里取值注意下
        ngx_memcpy(c->sockaddr, &sa, socklen);

        log = ngx_palloc(c->pool, sizeof(ngx_log_t));
        if (log == NULL) {
            ngx_close_accepted_connection(c);
            return;
        }

        /* set a blocking mode for iocp and non-blocking mode for others */

        if (ngx_inherited_nonblocking) {
            if (ngx_event_flags & NGX_USE_IOCP_EVENT) {
                if (ngx_blocking(s) == -1) {
                    ngx_log_error(NGX_LOG_ALERT, ev->log, ngx_socket_errno,
                                  ngx_blocking_n " failed");
                    ngx_close_accepted_connection(c);
                    return;
                }
            }

        } else {
            if (!(ngx_event_flags & NGX_USE_IOCP_EVENT)) {
                if (ngx_nonblocking(s) == -1) {
                    ngx_log_error(NGX_LOG_ALERT, ev->log, ngx_socket_errno,
                                  ngx_nonblocking_n " failed");
                    ngx_close_accepted_connection(c);
                    return;
                }
            }
        }

        *log = ls->log;
        c->recv = ngx_recv;
        c->send = ngx_send;
        c->recv_chain = ngx_recv_chain;
        c->send_chain = ngx_send_chain;
        c->log = log;
        c->pool->log = log;
        c->socklen = socklen;
        c->listening = ls;
        c->local_sockaddr = ls->sockaddr;
        c->local_socklen = ls->socklen;
        rev = c->read; // read event 没有设置 rev->ready=1
        wev = c->write;
        wev->ready = 1;
        if (ngx_event_flags & NGX_USE_IOCP_EVENT) {
            rev->ready = 1;
        }
        if (ev->deferred_accept) {
            rev->ready = 1;
#if (NGX_HAVE_KQUEUE || NGX_HAVE_EPOLLRDHUP)
            rev->available = 1;
#endif
        }

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
                ngx_close_accepted_connection(c);
                return;
            }

            c->addr_text.len = ngx_sock_ntop(c->sockaddr, c->socklen, c->addr_text.data, ls->addr_text_max_len, 0);
            if (c->addr_text.len == 0) {
                ngx_close_accepted_connection(c);
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
            addr.len = ngx_sock_ntop(c->sockaddr, c->socklen, text, NGX_SOCKADDR_STRLEN, 1);
            ngx_log_debug3(NGX_LOG_DEBUG_EVENT, log, 0, "*%uA accept: %V fd:%d", c->number, &addr, s);
        }

        }
#endif
        // linux 里 ngx_epoll_add_connection() -> epoll_ctl(ep, EPOLL_CTL_ADD, c->fd, &ee)
        if (ngx_add_conn && (ngx_event_flags & NGX_USE_EPOLL_EVENT) == 0) {
            if (ngx_add_conn(c) == NGX_ERROR) {
                ngx_close_accepted_connection(c);
                return;
            }
        }

        log->data = NULL;
        log->handler = NULL;

#if (T_NGX_ACCEPT_FILTER)
        /* accept filter */
        ngx_int_t          rc;
        rc = ngx_event_top_accept_filter(c);
        if (rc == NGX_ERROR) {
            ngx_close_accepted_connection(c);
            return;
        }
        if (rc == NGX_DECLINED) {
            if (ngx_event_flags & NGX_USE_KQUEUE_EVENT) {
                ev->available--;
            }
            continue;
        }
#endif

        // 这里是重点，读/写数据, -> ngx_stream_init_connection()
        ls->handler(c);
        if (ngx_event_flags & NGX_USE_KQUEUE_EVENT) {
            ev->available--;
        }

    } while (ev->available);
}

static void ngx_close_accepted_connection(ngx_connection_t *c) {
    ngx_socket_t  fd;

    ngx_free_connection(c);

    fd = c->fd;
    c->fd = (ngx_socket_t) -1;

    if (ngx_close_socket(fd) == -1) {
        ngx_log_error(NGX_LOG_ALERT, c->log, ngx_socket_errno,
                      ngx_close_socket_n " failed");
    }

    if (c->pool) {
        ngx_destroy_pool(c->pool);
    }

#if (NGX_STAT_STUB)
    (void) ngx_atomic_fetch_add(ngx_stat_active, -1);
#endif
}

u_char * ngx_accept_log_error(ngx_log_t *log, u_char *buf, size_t len) {
    return ngx_snprintf(buf, len, " while accepting new connection on %V", log->data);
}


#if (NGX_DEBUG)

void ngx_debug_accepted_connection(ngx_event_conf_t *ecf, ngx_connection_t *c) {
    struct sockaddr_in   *sin;
    ngx_cidr_t           *cidr;
    ngx_uint_t            i;
    cidr = ecf->debug_connection.elts;
    for (i = 0; i < ecf->debug_connection.nelts; i++) {
        if (cidr[i].family != (ngx_uint_t) c->sockaddr->sa_family) {
            goto next;
        }

        switch (cidr[i].family) {
#if (NGX_HAVE_UNIX_DOMAIN)
        case AF_UNIX:
            break;
#endif

        default: /* AF_INET */
            sin = (struct sockaddr_in *) c->sockaddr;
            if ((sin->sin_addr.s_addr & cidr[i].u.in.mask)
                != cidr[i].u.in.addr)
            {
                goto next;
            }
            break;
        }

        c->log->log_level = NGX_LOG_DEBUG_CONNECTION|NGX_LOG_DEBUG_ALL;
        break;

    next:
        continue;
    }
}

#endif