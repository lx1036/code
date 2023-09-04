


#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_channel.h>


static ngx_uint_t       ngx_pipe_generation;
static ngx_uint_t       ngx_last_pipe;
static ngx_open_pipe_t  ngx_pipes[NGX_MAX_PROCESSES];

#define MAX_BACKUP_NUM          128
#define NGX_PIPE_DIR_ACCESS     S_IRWXU | S_IRWXG | S_IROTH | S_IXOTH
#define NGX_PIPE_FILE_ACCESS    S_IRUSR | S_IWUSR | S_IRGRP | S_IWGRP | S_IROTH

typedef struct {
    ngx_int_t       time_now;
    ngx_int_t       last_open_time;
    ngx_int_t       log_size;
    ngx_int_t       last_suit_time;

    char           *logname;
    char           *backup[MAX_BACKUP_NUM];

    ngx_int_t       backup_num;
    ngx_int_t       log_max_size;
    ngx_int_t       interval;
    char           *suitpath;
    ngx_int_t       adjust_time;
    ngx_int_t       adjust_time_raw;
} ngx_pipe_rollback_conf_t;

static void ngx_signal_pipe_broken(ngx_log_t *log, ngx_pid_t pid);
static ngx_int_t ngx_open_pipe(ngx_cycle_t *cycle, ngx_open_pipe_t *op);
static void ngx_close_pipe(ngx_open_pipe_t *pipe);

static void ngx_pipe_log(ngx_cycle_t *cycle, ngx_open_pipe_t *op);
void ngx_pipe_get_last_rollback_time(ngx_pipe_rollback_conf_t *rbcf);
static void ngx_pipe_do_rollback(ngx_cycle_t *cycle, ngx_pipe_rollback_conf_t *rbcf);
static ngx_int_t ngx_pipe_rollback_parse_args(ngx_cycle_t *cycle,
    ngx_open_pipe_t *op, ngx_pipe_rollback_conf_t *rbcf);

ngx_str_t ngx_log_error_backup = ngx_string(NGX_ERROR_LOG_PATH);
ngx_str_t ngx_log_access_backup = ngx_string(NGX_HTTP_LOG_PATH);

ngx_str_t ngx_pipe_dev_null_file = ngx_string("/dev/null");


ngx_int_t
ngx_open_pipes(ngx_cycle_t *cycle)
{
    ngx_int_t          stat;
    ngx_uint_t         i;
    ngx_core_conf_t   *ccf;

    ccf = (ngx_core_conf_t *) ngx_get_conf(cycle->conf_ctx, ngx_core_module);

    for (i = 0; i < ngx_last_pipe; i++) {

        if (!ngx_pipes[i].configured) {
            continue;
        }

        if (ngx_pipes[i].generation != ngx_pipe_generation) {
            continue;
        }

        ngx_pipes[i].backup = ngx_pipes[i].open_fd->name;
        ngx_pipes[i].user = ccf->user;

        stat = ngx_open_pipe(cycle, &ngx_pipes[i]);

        ngx_log_debug4(NGX_LOG_DEBUG_CORE, cycle->log, 0,
                       "pipe: %ui(%d, %d) \"%s\"",
                       i, ngx_pipes[i].pfd[0],
                       ngx_pipes[i].pfd[1], ngx_pipes[i].cmd);

        if (stat == NGX_ERROR) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                          "open pipe \"%s\" failed",
                          ngx_pipes[i].cmd);
            return NGX_ERROR;
        }

        if (fcntl(ngx_pipes[i].open_fd->fd, F_SETFD, FD_CLOEXEC) == -1) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                          "fcntl(FD_CLOEXEC) \"%s\" failed",
                          ngx_pipes[i].cmd);
            return NGX_ERROR;
        }

        if (ngx_nonblocking(ngx_pipes[i].open_fd->fd) == -1) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                          "nonblock \"%s\" failed",
                          ngx_pipes[i].cmd);
            return NGX_ERROR;
        }

        ngx_pipes[i].open_fd->name.len = 0;
        ngx_pipes[i].open_fd->name.data = NULL;
    }

    return NGX_OK;
}

static ngx_int_t ngx_open_pipe(ngx_cycle_t *cycle, ngx_open_pipe_t *op) {
    int               fd;
    u_char          **argv;
    ngx_pid_t         pid;
    sigset_t          set;
#ifdef T_PIPE_USE_USER
    ngx_core_conf_t  *ccf;
    ccf = (ngx_core_conf_t *) ngx_get_conf(cycle->conf_ctx, ngx_core_module);
#endif

    if (pipe(op->pfd) < 0) {
        return NGX_ERROR;
    }

    argv = op->argv->elts;
    pid = fork();
    if (pid < 0) {
        goto err;
    } else if (pid > 0) {
        op->pid = pid;

        if (op->open_fd->fd != NGX_INVALID_FILE) {
            if (close(op->open_fd->fd) == NGX_FILE_ERROR) {
                ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                              "close \"%s\" failed",
                              op->open_fd->name.data);
            }
        }

        if (op->type == NGX_PIPE_WRITE) {
            op->open_fd->fd = op->pfd[1];
            close(op->pfd[0]);
        } else {
            op->open_fd->fd = op->pfd[0];
            close(op->pfd[1]);
        }
    } else {

       /*
        * Set correct process type since closing listening Unix domain socket
        * in a master process also removes the Unix domain socket file.
        */
        ngx_process = NGX_PROCESS_PIPE;
        ngx_close_listening_sockets(cycle);

        if (op->type == 1) {
            close(op->pfd[1]);
            if (op->pfd[0] != STDIN_FILENO) {
                dup2(op->pfd[0], STDIN_FILENO);
                close(op->pfd[0]);
            }
        } else {
            close(op->pfd[0]);
            if (op->pfd[1] != STDOUT_FILENO) {
                dup2(op->pfd[1], STDOUT_FILENO);
                close(op->pfd[1]);
            }
        }
#ifdef T_PIPE_USE_USER
        if (geteuid() == 0) {
            if (setgid(ccf->group) == -1) {
                ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                              "setgid(%d) failed", ccf->group);
                exit(2);
            }

            if (initgroups(ccf->username, ccf->group) == -1) {
                ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                              "initgroups(%s, %d) failed",
                              ccf->username, ccf->group);
            }

            if (setuid(ccf->user) == -1) {
                ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                              "setuid(%d) failed", ccf->user);
                exit(2);
            }
        }
#endif

        /*
         * redirect stderr to /dev/null, because stderr will be connected with
         * fd used by the last pipe when error log is configured using pipe,
         * that will cause it no close
         */

        fd = ngx_open_file("/dev/null", NGX_FILE_WRONLY, NGX_FILE_OPEN, 0);
        if (fd == -1) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                          "open(\"/dev/null\") failed");
            exit(2);
        }

        if (dup2(fd, STDERR_FILENO) == -1) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                          "dup2(STDERR) failed");
            exit(2);
        }

        if (fd > STDERR_FILENO && close(fd) == -1) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                          "close() failed");
            exit(2);
        }

        sigemptyset(&set);

        if (sigprocmask(SIG_SETMASK, &set, NULL) == -1) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                          "sigprocmask() failed");
            exit(2);
        }

        if (ngx_strncmp(argv[0], "rollback", sizeof("rollback") - 1) == 0) {
            ngx_pipe_log(cycle, op);
            exit(0);

        } else {
            execv((const char *) argv[0], (char *const *) op->argv->elts);
            exit(0);
        }
    }

    return NGX_OK;

err:

    close(op->pfd[0]);
    close(op->pfd[1]);

    return NGX_ERROR;
}

void ngx_increase_pipe_generation(void) {
    ngx_pipe_generation++;
}

void ngx_close_pipes(void) {
    ngx_uint_t i, last;

    for (i = 0, last = -1; i < ngx_last_pipe; i++) {

        if (!ngx_pipes[i].configured) {
            continue;
        }

        if (ngx_pipes[i].generation == ngx_pipe_generation) {
            ngx_close_pipe(&ngx_pipes[i]);
        } else {
            last = i;
        }
    }

    ngx_last_pipe = last + 1;
}

void ngx_close_old_pipes(void) {
    ngx_uint_t i, last;
    for (i = 0, last = -1; i < ngx_last_pipe; i++) {
        if (!ngx_pipes[i].configured) {
            continue;
        }

        if (ngx_pipes[i].generation < ngx_pipe_generation) {
            ngx_close_pipe(&ngx_pipes[i]);
        } else {
            last = i;
        }
    }

    ngx_last_pipe = last + 1;
}

