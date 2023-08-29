



#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_event.h>
#include <ngx_channel.h>


#if (T_PIPES)
#define NGX_PIPE_STILL_NEEDED     0x02
#endif

static void ngx_start_worker_processes(ngx_cycle_t *cycle, ngx_int_t n,
    ngx_int_t type);
static void ngx_start_cache_manager_processes(ngx_cycle_t *cycle,
    ngx_uint_t respawn);
static void ngx_pass_open_channel(ngx_cycle_t *cycle);
static void ngx_signal_worker_processes(ngx_cycle_t *cycle, int signo);
static ngx_uint_t ngx_reap_children(ngx_cycle_t *cycle);
static void ngx_master_process_exit(ngx_cycle_t *cycle);
static void ngx_worker_process_cycle(ngx_cycle_t *cycle, void *data);
static void ngx_worker_process_init(ngx_cycle_t *cycle, ngx_int_t worker);
static void ngx_worker_process_exit(ngx_cycle_t *cycle);
static void ngx_channel_handler(ngx_event_t *ev);
static void ngx_cache_manager_process_cycle(ngx_cycle_t *cycle, void *data);
static void ngx_cache_manager_process_handler(ngx_event_t *ev);
static void ngx_cache_loader_process_handler(ngx_event_t *ev);


ngx_uint_t    ngx_process;
ngx_uint_t    ngx_worker;
ngx_pid_t     ngx_pid;
ngx_pid_t     ngx_parent;

sig_atomic_t  ngx_reap;
sig_atomic_t  ngx_sigio;
sig_atomic_t  ngx_sigalrm;
sig_atomic_t  ngx_terminate;
sig_atomic_t  ngx_quit;
sig_atomic_t  ngx_debug_quit;
ngx_uint_t    ngx_exiting;
sig_atomic_t  ngx_reconfigure;
sig_atomic_t  ngx_reopen;

sig_atomic_t  ngx_change_binary;
ngx_pid_t     ngx_new_binary;
ngx_uint_t    ngx_inherited;
ngx_uint_t    ngx_daemonized;

sig_atomic_t  ngx_noaccept;
ngx_uint_t    ngx_noaccepting;
ngx_uint_t    ngx_restart;

static u_char  master_process[] = "master process";


void ngx_master_process_cycle(ngx_cycle_t *cycle) {
    char              *title;
    u_char            *p;
    size_t             size;
    ngx_int_t          i;
    ngx_uint_t         sigio;
#if (T_PIPES)
    ngx_uint_t         close_old_pipe;
#endif
    sigset_t           set;
    struct itimerval   itv;
    ngx_uint_t         live;
    ngx_msec_t         delay;
    ngx_core_conf_t   *ccf;

    sigemptyset(&set);
    sigaddset(&set, SIGCHLD);
    sigaddset(&set, SIGALRM);
    sigaddset(&set, SIGIO);
    sigaddset(&set, SIGINT);
    sigaddset(&set, ngx_signal_value(NGX_RECONFIGURE_SIGNAL));
    sigaddset(&set, ngx_signal_value(NGX_REOPEN_SIGNAL));
    sigaddset(&set, ngx_signal_value(NGX_NOACCEPT_SIGNAL));
    sigaddset(&set, ngx_signal_value(NGX_TERMINATE_SIGNAL));
    sigaddset(&set, ngx_signal_value(NGX_SHUTDOWN_SIGNAL));
    sigaddset(&set, ngx_signal_value(NGX_CHANGEBIN_SIGNAL));
    if (sigprocmask(SIG_BLOCK, &set, NULL) == -1) {
        ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                      "sigprocmask() failed");
    }
    sigemptyset(&set);

    size = sizeof(master_process);
    for (i = 0; i < ngx_argc; i++) {
        size += ngx_strlen(ngx_argv[i]) + 1;
    }
    title = ngx_pnalloc(cycle->pool, size);
    if (title == NULL) {
        /* fatal */
        exit(2);
    }
    p = ngx_cpymem(title, master_process, sizeof(master_process) - 1);
    for (i = 0; i < ngx_argc; i++) {
        *p++ = ' ';
        p = ngx_cpystrn(p, (u_char *) ngx_argv[i], size);
    }
    ngx_setproctitle(title);

    // ccf 就是 整个 nginx.conf 配置对象
    ccf = (ngx_core_conf_t *) ngx_get_conf(cycle->conf_ctx, ngx_core_module);
    ngx_start_worker_processes(cycle, ccf->worker_processes, NGX_PROCESS_RESPAWN);
    // ngx_start_cache_manager_processes(cycle, 0);

#if (NGX_PROCS)
    // ngx_procs_start(cycle, 0);
#endif

    ngx_new_binary = 0;
#if (T_PIPES)
    close_old_pipe = 0;
#endif
    delay = 0;
    sigio = 0;
    live = 1;
    for ( ;; ){
        if (delay) {
            // todo
        }

        ngx_log_debug0(NGX_LOG_DEBUG_EVENT, cycle->log, 0, "sigsuspend");
        sigsuspend(&set);
        ngx_time_update();
        ngx_log_debug1(NGX_LOG_DEBUG_EVENT, cycle->log, 0, "wake up, sigio %i", sigio);

    }
}

// type=NGX_PROCESS_RESPAWN=-3
static void ngx_start_worker_processes(ngx_cycle_t *cycle, ngx_int_t n, ngx_int_t type) {
    ngx_int_t  i;
    ngx_log_error(NGX_LOG_NOTICE, cycle->log, 0, "start worker processes");

    for (i = 0; i < n; i++) {
        ngx_spawn_process(cycle, ngx_worker_process_cycle, (void *) (intptr_t) i, "worker process", type);
        ngx_pass_open_channel(cycle);
    }
}

ngx_pid_t ngx_spawn_process(ngx_cycle_t *cycle, ngx_spawn_proc_pt proc, void *data,
    char *name, ngx_int_t respawn) {
    u_long     on;
    ngx_pid_t  pid;
    ngx_int_t  s;

    if (respawn >= 0) {
        s = respawn;
    } else {
        for (s = 0; s < ngx_last_process; s++) {
            if (ngx_processes[s].pid == -1) {
                break;
            }
        }
        if (s == NGX_MAX_PROCESSES) {
            ngx_log_error(NGX_LOG_ALERT, cycle->log, 0,
                          "no more than %d processes can be spawned",
                          NGX_MAX_PROCESSES);
            return NGX_INVALID_PID;
        }
    }

    ngx_log_error(NGX_LOG_INFO, cycle->log, ngx_errno, "process index: %d", s);

    // info: 新建进程模板代码
    if (respawn != NGX_PROCESS_DETACHED) {
        /* Solaris 9 still has no AF_LOCAL */
        if (socketpair(AF_UNIX, SOCK_STREAM, 0, ngx_processes[s].channel) == -1){
            ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                          "socketpair() failed while spawning \"%s\"", name);
            return NGX_INVALID_PID;
        }

        ngx_log_debug2(NGX_LOG_DEBUG_CORE, cycle->log, 0,
                       "channel %d:%d",
                       ngx_processes[s].channel[0],
                       ngx_processes[s].channel[1]);

        if (ngx_nonblocking(ngx_processes[s].channel[0]) == -1) {
            ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                          ngx_nonblocking_n " failed while spawning \"%s\"",
                          name);
            ngx_close_channel(ngx_processes[s].channel, cycle->log);
            return NGX_INVALID_PID;
        }

        if (ngx_nonblocking(ngx_processes[s].channel[1]) == -1) {
            ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                          ngx_nonblocking_n " failed while spawning \"%s\"",
                          name);
            ngx_close_channel(ngx_processes[s].channel, cycle->log);
            return NGX_INVALID_PID;
        }

        on = 1;
        if (ioctl(ngx_processes[s].channel[0], FIOASYNC, &on) == -1) {
            ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                          "ioctl(FIOASYNC) failed while spawning \"%s\"", name);
            ngx_close_channel(ngx_processes[s].channel, cycle->log);
            return NGX_INVALID_PID;
        }

        if (fcntl(ngx_processes[s].channel[0], F_SETOWN, ngx_pid) == -1) {
            ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                          "fcntl(F_SETOWN) failed while spawning \"%s\"", name);
            ngx_close_channel(ngx_processes[s].channel, cycle->log);
            return NGX_INVALID_PID;
        }

        if (fcntl(ngx_processes[s].channel[0], F_SETFD, FD_CLOEXEC) == -1) {
            ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                          "fcntl(FD_CLOEXEC) failed while spawning \"%s\"",
                           name);
            ngx_close_channel(ngx_processes[s].channel, cycle->log);
            return NGX_INVALID_PID;
        }

        if (fcntl(ngx_processes[s].channel[1], F_SETFD, FD_CLOEXEC) == -1) {
            ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                          "fcntl(FD_CLOEXEC) failed while spawning \"%s\"",
                           name);
            ngx_close_channel(ngx_processes[s].channel, cycle->log);
            return NGX_INVALID_PID;
        }

        ngx_channel = ngx_processes[s].channel[1];
    } else {
        ngx_processes[s].channel[0] = -1;
        ngx_processes[s].channel[1] = -1;
    }

    ngx_process_slot = s;
    pid = fork();
    switch (pid) {
    case -1:
        ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                      "fork() failed while spawning \"%s\"", name);
        ngx_close_channel(ngx_processes[s].channel, cycle->log);
        return NGX_INVALID_PID;

    case 0:
        ngx_parent = ngx_pid;
        ngx_pid = ngx_getpid();
        proc(cycle, data); // ngx_worker_process_cycle(cycle, data)
        break;

    default:
        break;
    }

    ngx_log_error(NGX_LOG_NOTICE, cycle->log, 0, "start %s %P", name, pid);
    ngx_processes[s].pid = pid;
    ngx_processes[s].exited = 0;
    if (respawn >= 0) {
        return pid;
    }

    switch (respawn) {
    case NGX_PROCESS_NORESPAWN:
        ngx_processes[s].respawn = 0;
        ngx_processes[s].just_spawn = 0;
        ngx_processes[s].detached = 0;
        break;

    case NGX_PROCESS_JUST_SPAWN:
        ngx_processes[s].respawn = 0;
        ngx_processes[s].just_spawn = 1;
        ngx_processes[s].detached = 0;
        break;

    case NGX_PROCESS_RESPAWN:
        ngx_processes[s].respawn = 1;
        ngx_processes[s].just_spawn = 0;
        ngx_processes[s].detached = 0;
        break;

    case NGX_PROCESS_JUST_RESPAWN:
        ngx_processes[s].respawn = 1;
        ngx_processes[s].just_spawn = 1;
        ngx_processes[s].detached = 0;
        break;

    case NGX_PROCESS_DETACHED:
        ngx_processes[s].respawn = 0;
        ngx_processes[s].just_spawn = 0;
        ngx_processes[s].detached = 1;
        break;
    }

    if (s == ngx_last_process) {
        ngx_last_process++;
    }

    return pid;
}

static void ngx_worker_process_cycle(ngx_cycle_t *cycle, void *data){
    ngx_int_t worker = (intptr_t) data;
    ngx_process = NGX_PROCESS_WORKER;
    ngx_worker = worker;
    ngx_worker_process_init(cycle, worker);
    ngx_setproctitle("worker process");

    for ( ;; ) {
        if (ngx_exiting) {
            // todo
        }
        ngx_log_debug0(NGX_LOG_DEBUG_EVENT, cycle->log, 0, "worker cycle");
        ngx_process_events_and_timers(cycle);

        if (ngx_terminate) {
            // todo
        }
        if (ngx_quit) {
            // todo
        }
        if (ngx_reopen) {
            // todo
        }
    }
}

static void ngx_pass_open_channel(ngx_cycle_t *cycle){
    ngx_int_t      i;
    ngx_channel_t  ch;

    ngx_memzero(&ch, sizeof(ngx_channel_t));
    ch.command = NGX_CMD_OPEN_CHANNEL;
    ch.pid = ngx_processes[ngx_process_slot].pid;
    ch.slot = ngx_process_slot;
    ch.fd = ngx_processes[ngx_process_slot].channel[0];
    for (i = 0; i < ngx_last_process; i++) {
        if (i == ngx_process_slot
            || ngx_processes[i].pid == -1
            || ngx_processes[i].channel[0] == -1) {
            continue;
        }

        ngx_log_debug6(NGX_LOG_DEBUG_CORE, cycle->log, 0,
                      "pass channel s:%i pid:%P fd:%d to s:%i pid:%P fd:%d",
                      ch.slot, ch.pid, ch.fd,
                      i, ngx_processes[i].pid,
                      ngx_processes[i].channel[0]);

        /* TODO: NGX_AGAIN */

        ngx_write_channel(ngx_processes[i].channel[0],
                          &ch, sizeof(ngx_channel_t), cycle->log);
    }
}

static void ngx_worker_process_init(ngx_cycle_t *cycle, ngx_int_t worker) {
    ngx_core_conf_t  *ccf;
    ccf = (ngx_core_conf_t *) ngx_get_conf(cycle->conf_ctx, ngx_core_module);

    if (ngx_add_channel_event(cycle, ngx_channel, NGX_READ_EVENT, ngx_channel_handler) == NGX_ERROR) {
        /* fatal */
        exit(2);
    }
}

static void ngx_channel_handler(ngx_event_t *ev) {
    ngx_int_t          n;
    ngx_channel_t      ch;
    ngx_connection_t  *c;

    if (ev->timedout) {
        ev->timedout = 0;
        return;
    }

    c = ev->data;
    ngx_log_debug0(NGX_LOG_DEBUG_CORE, ev->log, 0, "channel handler");

    for ( ;; ) {
        n = ngx_read_channel(c->fd, &ch, sizeof(ngx_channel_t), ev->log);
        ngx_log_debug1(NGX_LOG_DEBUG_CORE, ev->log, 0, "channel: %i", n);
        if (n == NGX_ERROR) {
            if (ngx_event_flags & NGX_USE_EPOLL_EVENT) {
#if (NGX_HTTP_SSL && NGX_SSL_ASYNC)
            if (c->async_enable && ngx_del_async_conn) {
                if (c->num_async_fds) {
                    ngx_del_async_conn(c, NGX_DISABLE_EVENT);
                    c->num_async_fds--;
                }
            }
#endif
                ngx_del_conn(c, 0);
            }

            ngx_close_connection(c);
            return;
        }

        if (ngx_event_flags & NGX_USE_EVENTPORT_EVENT) {
            if (ngx_add_event(ev, NGX_READ_EVENT, 0) == NGX_ERROR) {
                return;
            }
        }

        if (n == NGX_AGAIN) {
            return;
        }

        switch (ch.command) {
        case NGX_CMD_QUIT:
            ngx_quit = 1;
            break;

        case NGX_CMD_TERMINATE:
            ngx_terminate = 1;
            break;

        case NGX_CMD_REOPEN:
            ngx_reopen = 1;
            break;

        case NGX_CMD_OPEN_CHANNEL:
            ngx_log_debug3(NGX_LOG_DEBUG_CORE, ev->log, 0,
                           "get channel s:%i pid:%P fd:%d",
                           ch.slot, ch.pid, ch.fd);
            ngx_processes[ch.slot].pid = ch.pid;
            ngx_processes[ch.slot].channel[0] = ch.fd;
            break;

        case NGX_CMD_CLOSE_CHANNEL:
            ngx_log_debug4(NGX_LOG_DEBUG_CORE, ev->log, 0,
                           "close channel s:%i pid:%P our:%P fd:%d",
                           ch.slot, ch.pid, ngx_processes[ch.slot].pid,
                           ngx_processes[ch.slot].channel[0]);
            if (close(ngx_processes[ch.slot].channel[0]) == -1) {
                ngx_log_error(NGX_LOG_ALERT, ev->log, ngx_errno, "close() channel failed");
            }
            ngx_processes[ch.slot].channel[0] = -1;
            break;
        }
    }
}