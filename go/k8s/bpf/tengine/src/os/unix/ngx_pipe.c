


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

