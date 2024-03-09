
#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_event.h>


typedef struct {
    ngx_uint_t  events;
    ngx_uint_t  aio_requests;
} ngx_epoll_conf_t;


static ngx_int_t ngx_epoll_init(ngx_cycle_t *cycle, ngx_msec_t timer);
#if (NGX_HAVE_EVENTFD)
static ngx_int_t ngx_epoll_notify_init(ngx_log_t *log);
static void ngx_epoll_notify_handler(ngx_event_t *ev);
#endif
#if (NGX_HAVE_EPOLLRDHUP)
static void ngx_epoll_test_rdhup(ngx_cycle_t *cycle);
#endif
static void ngx_epoll_done(ngx_cycle_t *cycle);
static ngx_int_t ngx_epoll_add_event(ngx_event_t *ev, ngx_int_t event,
    ngx_uint_t flags);
static ngx_int_t ngx_epoll_del_event(ngx_event_t *ev, ngx_int_t event,
    ngx_uint_t flags);
static ngx_int_t ngx_epoll_add_connection(ngx_connection_t *c);
static ngx_int_t ngx_epoll_del_connection(ngx_connection_t *c,
    ngx_uint_t flags);
#if (NGX_HAVE_EVENTFD)
static ngx_int_t ngx_epoll_notify(ngx_event_handler_pt handler);
#endif
static ngx_int_t ngx_epoll_process_events(ngx_cycle_t *cycle, ngx_msec_t timer,
    ngx_uint_t flags);
#if (NGX_SSL && NGX_SSL_ASYNC)
static ngx_int_t ngx_epoll_add_async_connection(ngx_connection_t *c);
static ngx_int_t ngx_epoll_del_async_connection(ngx_connection_t *c,
    ngx_uint_t flags);
#endif

#if (NGX_HAVE_FILE_AIO)
static void ngx_epoll_eventfd_handler(ngx_event_t *ev);
#endif

static void *ngx_epoll_create_conf(ngx_cycle_t *cycle);
static char *ngx_epoll_init_conf(ngx_cycle_t *cycle, void *conf);

static int                  ep = -1;
static struct epoll_event  *event_list;
static ngx_uint_t           nevents;

// https://www.cnblogs.com/xuewangkai/p/11158576.html
static ngx_str_t      epoll_name = ngx_string("epoll");
static ngx_command_t  ngx_epoll_commands[] = {
    { ngx_string("epoll_events"),
      NGX_EVENT_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_num_slot,
      0,
      offsetof(ngx_epoll_conf_t, events),
      NULL },

    { ngx_string("worker_aio_requests"),
      NGX_EVENT_CONF|NGX_CONF_TAKE1,
      ngx_conf_set_num_slot,
      0,
      offsetof(ngx_epoll_conf_t, aio_requests),
      NULL },

      ngx_null_command
};
static ngx_event_module_t  ngx_epoll_module_ctx = {
    &epoll_name,
    ngx_epoll_create_conf,               /* create configuration */
    ngx_epoll_init_conf,                 /* init configuration */

    {
        ngx_epoll_add_event,             /* add an event */
        ngx_epoll_del_event,             /* delete an event */
        ngx_epoll_add_event,             /* enable an event */
        ngx_epoll_del_event,             /* disable an event */
        ngx_epoll_add_connection,        /* add an connection */
        ngx_epoll_del_connection,        /* delete an connection */
#if (NGX_HAVE_EVENTFD)
        ngx_epoll_notify,                /* trigger a notify */
#else
        NULL,                            /* trigger a notify */
#endif
        ngx_epoll_process_events,        /* process the events */
        ngx_epoll_init,                  /* init the events */
        ngx_epoll_done,                  /* done the events */
    }
};
ngx_module_t  ngx_epoll_module = {
    NGX_MODULE_V1,
    &ngx_epoll_module_ctx,               /* module context */
    ngx_epoll_commands,                  /* module directives */
    NGX_EVENT_MODULE,                    /* module type */
    NULL,                                /* init master */
    NULL,                                /* init module */
    NULL,                                /* init process */
    NULL,                                /* init thread */
    NULL,                                /* exit thread */
    NULL,                                /* exit process */
    NULL,                                /* exit master */
    NGX_MODULE_V1_PADDING
};

static void * ngx_epoll_create_conf(ngx_cycle_t *cycle) {
    ngx_epoll_conf_t  *epcf;
    epcf = ngx_palloc(cycle->pool, sizeof(ngx_epoll_conf_t));
    if (epcf == NULL) {
        return NULL;
    }

    epcf->events = NGX_CONF_UNSET;
    epcf->aio_requests = NGX_CONF_UNSET;

    return epcf;
}

static char * ngx_epoll_init_conf(ngx_cycle_t *cycle, void *conf) {
    ngx_epoll_conf_t *epcf = conf;
    ngx_conf_init_uint_value(epcf->events, 512);
    ngx_conf_init_uint_value(epcf->aio_requests, 32);

    return NGX_CONF_OK;
}

// 参考 epoll_test.c, 常用逻辑
static ngx_int_t ngx_epoll_add_event(ngx_event_t *ev, ngx_int_t event, ngx_uint_t flags)
{
    int                  op;
    uint32_t             events, prev;
    ngx_event_t         *e;
    ngx_connection_t    *c;
    struct epoll_event   ee;

    c = ev->data; // c->fd 就是 每次 epoll_ctl(xxx,fd,xxx)
    events = (uint32_t) event;
    if (event == NGX_READ_EVENT) {
        e = c->write;
        prev = EPOLLOUT; // 设置用于注测的写操作事件
#if (NGX_READ_EVENT != EPOLLIN|EPOLLRDHUP)
        events = EPOLLIN|EPOLLRDHUP; // epoll_in-> 读, epoll_out -> 写
#endif
    } else {
        e = c->read;
        prev = EPOLLIN|EPOLLRDHUP; // 如果是已经连接的用户，并且收到数据，那么进行读数据
#if (NGX_WRITE_EVENT != EPOLLOUT)
        events = EPOLLOUT;
#endif
    }

    if (e->active) {
        op = EPOLL_CTL_MOD; // 已有链接，读数据，然后写数据
        events |= prev;
    } else {
        op = EPOLL_CTL_ADD; // 监测到 socket 来了一个新连接
    }

#if (NGX_HAVE_EPOLLEXCLUSIVE && NGX_HAVE_EPOLLRDHUP)
    if (flags & NGX_EXCLUSIVE_EVENT) {
        events &= ~EPOLLRDHUP;
    }
#endif

    ee.events = events | (uint32_t) flags;
    ee.data.ptr = (void *) ((uintptr_t) c | ev->instance);
    ngx_log_debug3(NGX_LOG_DEBUG_EVENT, ev->log, 0, "epoll add event: fd:%d op:%d ev:%08XD", c->fd, op, ee.events);
    if (epoll_ctl(ep, op, c->fd, &ee) == -1) {
        ngx_log_error(NGX_LOG_ALERT, ev->log, ngx_errno, "epoll_ctl(%d, %d) failed", op, c->fd);
        return NGX_ERROR;
    }

    ev->active = 1;

    return NGX_OK;
}

static ngx_int_t ngx_epoll_del_event(ngx_event_t *ev, ngx_int_t event, ngx_uint_t flags) {
    int                  op;
    uint32_t             prev;
    ngx_event_t         *e;
    ngx_connection_t    *c;
    struct epoll_event   ee;

    /*
     * when the file descriptor is closed, the epoll automatically deletes
     * it from its queue, so we do not need to delete explicitly the event
     * before the closing the file descriptor
     */
    if (flags & NGX_CLOSE_EVENT) {
        ev->active = 0;
        return NGX_OK;
    }

    c = ev->data;
    if (event == NGX_READ_EVENT) { // 读数据
        e = c->write;
        prev = EPOLLOUT;
    } else {
        e = c->read;
        prev = EPOLLIN|EPOLLRDHUP;
    }

    if (e->active) {
        op = EPOLL_CTL_MOD;
        ee.events = prev | (uint32_t) flags;
        ee.data.ptr = (void *) ((uintptr_t) c | ev->instance);
    } else {
        op = EPOLL_CTL_DEL; // EPOLL_CTL_DEL 从 epfd 中删除一个文件描述符的注册事件 c->fd
        ee.events = 0;
        ee.data.ptr = NULL;
    }

    ngx_log_debug3(NGX_LOG_DEBUG_EVENT, ev->log, 0, "epoll del event: fd:%d op:%d ev:%08XD", c->fd, op, ee.events);
    if (epoll_ctl(ep, op, c->fd, &ee) == -1) {
        ngx_log_error(NGX_LOG_ALERT, ev->log, ngx_errno, "epoll_ctl(%d, %d) failed", op, c->fd);
        return NGX_ERROR;
    }

    ev->active = 0; // 这里注意下

    return NGX_OK;
}

static ngx_int_t ngx_epoll_add_connection(ngx_connection_t *c) {
    struct epoll_event  ee;

    ee.events = EPOLLIN|EPOLLOUT|EPOLLET|EPOLLRDHUP;
    ee.data.ptr = (void *) ((uintptr_t) c | c->read->instance); // 注意下 c->read->instance
    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "epoll add connection: fd:%d ev:%08XD", c->fd, ee.events);
    if (epoll_ctl(ep, EPOLL_CTL_ADD, c->fd, &ee) == -1) {
        ngx_log_error(NGX_LOG_ALERT, c->log, ngx_errno, "epoll_ctl(EPOLL_CTL_ADD, %d) failed", c->fd);
        return NGX_ERROR;
    }

    c->read->active = 1;
    c->write->active = 1;

    return NGX_OK;
}

static ngx_int_t ngx_epoll_del_connection(ngx_connection_t *c, ngx_uint_t flags) {
    int                 op;
    struct epoll_event  ee;

    /*
     * when the file descriptor is closed the epoll automatically deletes
     * it from its queue so we do not need to delete explicitly the event
     * before the closing the file descriptor
     */
    if (flags & NGX_CLOSE_EVENT) {
        c->read->active = 0;
        c->write->active = 0;
        return NGX_OK;
    }

    ngx_log_error(NGX_LOG_STDERR, c->log, 0, "epoll del connection: fd:%d", c->fd);
    op = EPOLL_CTL_DEL;
    ee.events = 0;
    ee.data.ptr = NULL;
    if (epoll_ctl(ep, op, c->fd, &ee) == -1) {
        ngx_log_error(NGX_LOG_ALERT, c->log, ngx_errno, "epoll_ctl(%d, %d) failed", op, c->fd);
        return NGX_ERROR;
    }

    c->read->active = 0;
    c->write->active = 0;

    return NGX_OK;
}


static ngx_int_t
ngx_epoll_init(ngx_cycle_t *cycle, ngx_msec_t timer)
{
    ngx_epoll_conf_t  *epcf;

    epcf = ngx_event_get_conf(cycle->conf_ctx, ngx_epoll_module);

    if (ep == -1) {
        // epoll_create() 创建 epoll_fd
        ep = epoll_create(cycle->connection_n / 2);
        if (ep == -1) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno, "epoll_create() failed");
            return NGX_ERROR;
        }

#if (NGX_HAVE_EVENTFD)
        if (ngx_epoll_notify_init(cycle->log) != NGX_OK) {
            ngx_epoll_module_ctx.actions.notify = NULL;
        }
#endif

// #if (NGX_HAVE_FILE_AIO)
//         ngx_epoll_aio_init(cycle, epcf);
// #endif

// #if (NGX_HAVE_EPOLLRDHUP)
//         ngx_epoll_test_rdhup(cycle);
// #endif
    }

    if (nevents < epcf->events) {
        if (event_list) {
            ngx_free(event_list);
        }
        // event_list[512] 默认值 512，可配置
        event_list = ngx_alloc(sizeof(struct epoll_event) * epcf->events, cycle->log);
        if (event_list == NULL) {
            return NGX_ERROR;
        }
    }

    nevents = epcf->events;
    /* 这里赋值很重要!
    ngx_os_io_t ngx_os_io = {
        ngx_unix_recv,
        ngx_readv_chain,
        ngx_udp_unix_recv,
        ngx_unix_send,
        ngx_udp_unix_send,
        ngx_udp_unix_sendmsg_chain,
        ngx_writev_chain,
        0
    };
    ngx_io 在 ngx_connection.c 里，传递过去
    */
    ngx_io = ngx_os_io;
    /* 很重要
    {
        ngx_epoll_add_event,     
        ngx_epoll_del_event,   
        ngx_epoll_add_event,        
        ngx_epoll_del_event,         
        ngx_epoll_add_connection,     
        ngx_epoll_del_connection,        
        #if (NGX_HAVE_EVENTFD)
                ngx_epoll_notify,            
        #else
                NULL,                        
        #endif
        ngx_epoll_process_events,      
        ngx_epoll_init,             
        ngx_epoll_done,             
    }
    */
    ngx_event_actions = ngx_epoll_module_ctx.actions;

#if (NGX_HAVE_CLEAR_EVENT)
    ngx_event_flags = NGX_USE_CLEAR_EVENT
#else
    ngx_event_flags = NGX_USE_LEVEL_EVENT
#endif
                      |NGX_USE_GREEDY_EVENT
                      |NGX_USE_EPOLL_EVENT;

    return NGX_OK;
}

static void ngx_epoll_done(ngx_cycle_t *cycle) {
    if (close(ep) == -1) {
        ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno, "epoll close() failed");
    }

    ep = -1;

#if (NGX_HAVE_EVENTFD)
    if (close(notify_fd) == -1) {
        ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno, "eventfd close() failed");
    }

    notify_fd = -1;

#endif

// #if (NGX_HAVE_FILE_AIO)

//     if (ngx_eventfd != -1) {

//         if (io_destroy(ngx_aio_ctx) == -1) {
//             ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
//                           "io_destroy() failed");
//         }

//         if (close(ngx_eventfd) == -1) {
//             ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
//                           "eventfd close() failed");
//         }

//         ngx_eventfd = -1;
//     }

//     ngx_aio_ctx = 0;

// #endif

    ngx_free(event_list);

    event_list = NULL;
    nevents = 0;
}

static ngx_int_t ngx_epoll_process_events(ngx_cycle_t *cycle, ngx_msec_t timer, ngx_uint_t flags) {
    int                events;
    uint32_t           revents;
    ngx_int_t          instance, i;
    ngx_uint_t         level;
    ngx_err_t          err;
    ngx_event_t       *rev, *wev;
    ngx_queue_t       *queue;
    ngx_connection_t  *c;

    /* NGX_TIMER_INFINITE == INFTIM */
    ngx_log_error(NGX_LOG_STDERR, cycle->log, 0, "epoll timer: %M", timer);
    // 等待事件发生 https://man7.org/linux/man-pages/man2/epoll_wait.2.html, 阻塞的
    events = epoll_wait(ep, event_list, (int) nevents, timer);
    err = (events == -1) ? ngx_errno : 0;
    if (flags & NGX_UPDATE_TIME || ngx_event_timer_alarm) {
        ngx_time_update();
    }
    if (err) {
        if (err == NGX_EINTR) {
            if (ngx_event_timer_alarm) {
                ngx_event_timer_alarm = 0;
                return NGX_OK;
            }

            level = NGX_LOG_INFO;
        } else {
            level = NGX_LOG_ALERT;
        }

        ngx_log_error(level, cycle->log, err, "epoll_wait() failed");
        return NGX_ERROR;
    }

    if (events == 0) {
        if (timer != NGX_TIMER_INFINITE) {
            return NGX_OK;
        }

        ngx_log_error(NGX_LOG_ALERT, cycle->log, 0, "epoll_wait() returned no events without timeout");
        return NGX_ERROR;
    }
    
    /*
        https://man7.org/linux/man-pages/man3/epoll_event.3type.html
        #include <sys/epoll.h>
       struct epoll_event {
           uint32_t      events;  // Epoll events
           epoll_data_t  data;   // User data variable
       };
       union epoll_data {
           void     *ptr;
           int       fd;
           uint32_t  u32;
           uint64_t  u64;
       };
       typedef union epoll_data  epoll_data_t;
    */
    for (i = 0; i < events; i++) {
        c = event_list[i].data.ptr;
        instance = (uintptr_t) c & 1;
#if (NGX_SSL)
        c = (ngx_connection_t *) ((uintptr_t) c & (uintptr_t) ~3);
#else
        c = (ngx_connection_t *) ((uintptr_t) c & (uintptr_t) ~1);
#endif

        rev = c->read;
        if (c->fd == -1 || rev->instance != instance) {
            /*
             * the stale event from a file descriptor
             * that was just closed in this iteration
             */
            ngx_log_error(NGX_LOG_STDERR, cycle->log, 0, "epoll: stale event %p", c);
            continue;
        }

        revents = event_list[i].events;
        ngx_log_error(NGX_LOG_DEBUG_EVENT, cycle->log, 0, "epoll: fd:%d ev:%04XD d:%p",
                       c->fd, revents, event_list[i].data.ptr);

        if (revents & (EPOLLERR|EPOLLHUP)) {
            ngx_log_error(NGX_LOG_STDERR, cycle->log, 0, "epoll_wait() error on fd:%d ev:%04XD", c->fd, revents);

            /*
             * if the error events were returned, add EPOLLIN and EPOLLOUT
             * to handle the events at least in one active handler
             */
            revents |= EPOLLIN|EPOLLOUT;
        }
        // 1.读数据
        if ((revents & EPOLLIN) && rev->active) {
            rev->ready = 1;
            rev->available = -1;

            if (flags & NGX_POST_EVENTS) {
                queue = rev->accept ? &ngx_posted_accept_events : &ngx_posted_events;
                ngx_post_event(rev, queue);
            } else {
                // 如果是 TCP, handler=ngx_event_accept();
                // 如果是 UDP, handler=ngx_event_recvmsg();
                // 看 ngx_event.c::ngx_event_process_init()
                // UDP 看 ngx_event
                rev->handler(rev);
            }
        }

        // 2.写数据
        wev = c->write;
        if ((revents & EPOLLOUT) && wev->active) {
            if (c->fd == -1 || wev->instance != instance) {
                /*
                 * the stale event from a file descriptor
                 * that was just closed in this iteration
                 */
                ngx_log_error(NGX_LOG_STDERR, cycle->log, 0, "epoll: stale event %p", c);
                continue;
            }

            wev->ready = 1;
#if (NGX_THREADS)
            wev->complete = 1;
#endif

            if (flags & NGX_POST_EVENTS) {
                ngx_post_event(wev, &ngx_posted_events);
            } else {
                // 如果是 TCP, handler=ngx_event_accept(); 这里不对，应该是写数据, send() 才对，也存疑(应该是对的)!!!
                // 如果是 UDP, handler=ngx_event_recvmsg();
                // 看 ngx_event.c::ngx_event_process_init()
                wev->handler(wev);
            }
        }
    }

    return NGX_OK;
}


#if (NGX_HAVE_EVENTFD)

static ngx_int_t
ngx_epoll_notify_init(ngx_log_t *log)
{
    struct epoll_event  ee;

#if (NGX_HAVE_SYS_EVENTFD_H)
    notify_fd = eventfd(0, 0);
#else
    notify_fd = syscall(SYS_eventfd, 0);
#endif

    if (notify_fd == -1) {
        ngx_log_error(NGX_LOG_EMERG, log, ngx_errno, "eventfd() failed");
        return NGX_ERROR;
    }

    ngx_log_error(NGX_LOG_STDERR, log, 0,
                   "notify eventfd: %d", notify_fd);

    notify_event.handler = ngx_epoll_notify_handler;
    notify_event.log = log;
    notify_event.active = 1;

    notify_conn.fd = notify_fd;
    notify_conn.read = &notify_event;
    notify_conn.log = log;

    ee.events = EPOLLIN|EPOLLET;
    ee.data.ptr = &notify_conn;

    if (epoll_ctl(ep, EPOLL_CTL_ADD, notify_fd, &ee) == -1) {
        ngx_log_error(NGX_LOG_EMERG, log, ngx_errno,
                      "epoll_ctl(EPOLL_CTL_ADD, eventfd) failed");

        if (close(notify_fd) == -1) {
            ngx_log_error(NGX_LOG_ALERT, log, ngx_errno,
                            "eventfd close() failed");
        }

        return NGX_ERROR;
    }

    return NGX_OK;
}


static void
ngx_epoll_notify_handler(ngx_event_t *ev)
{
    ssize_t               n;
    uint64_t              count;
    ngx_err_t             err;
    ngx_event_handler_pt  handler;

    if (++ev->index == NGX_MAX_UINT32_VALUE) {
        ev->index = 0;

        n = read(notify_fd, &count, sizeof(uint64_t));

        err = ngx_errno;

        ngx_log_debug3(NGX_LOG_DEBUG_EVENT, ev->log, 0,
                       "read() eventfd %d: %z count:%uL", notify_fd, n, count);

        if ((size_t) n != sizeof(uint64_t)) {
            ngx_log_error(NGX_LOG_ALERT, ev->log, err,
                          "read() eventfd %d failed", notify_fd);
        }
    }

    handler = ev->data;
    handler(ev);
}

#endif

