


#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_event.h>


#if (T_NGX_SHOW_INFO)
extern ngx_uint_t  ngx_modules_n;
#endif

static void ngx_destroy_cycle_pools(ngx_conf_t *conf);
static ngx_int_t ngx_init_zone_pool(ngx_cycle_t *cycle,
    ngx_shm_zone_t *shm_zone);
static ngx_int_t ngx_test_lockfile(u_char *file, ngx_log_t *log);
static void ngx_clean_old_cycles(ngx_event_t *ev);
static void ngx_shutdown_timer_handler(ngx_event_t *ev);


volatile ngx_cycle_t  *ngx_cycle;
ngx_array_t            ngx_old_cycles;

static ngx_pool_t     *ngx_temp_pool;
static ngx_event_t     ngx_cleaner_event;
static ngx_event_t     ngx_shutdown_event;

ngx_uint_t             ngx_test_config;
ngx_uint_t             ngx_dump_config;
ngx_uint_t             ngx_quiet_mode;
#if (T_NGX_SHOW_INFO)
ngx_uint_t             ngx_show_modules;
ngx_uint_t             ngx_show_directives;
#endif


/* STUB NAME */
static ngx_connection_t  dumb;
/* STUB */


ngx_cycle_t *
ngx_init_cycle(ngx_cycle_t *old_cycle) {
    void                *rv;
    char               **senv;
    ngx_uint_t           i, n;
    ngx_log_t           *log;
    ngx_time_t          *tp;
    ngx_conf_t           conf;
    ngx_pool_t          *pool;
    ngx_cycle_t         *cycle, **old;
    ngx_shm_zone_t      *shm_zone, *oshm_zone;
    ngx_list_part_t     *part, *opart;
    ngx_open_file_t     *file;
    ngx_listening_t     *ls, *nls;
    ngx_core_conf_t     *ccf, *old_ccf;
    ngx_core_module_t   *module;
    char                 hostname[NGX_MAXHOSTNAMELEN];

    
    log = old_cycle->log;
    pool = ngx_create_pool(NGX_CYCLE_POOL_SIZE, log);
    if (pool == NULL) {
        return NULL;
    }
    pool->log = log;

    cycle = ngx_pcalloc(pool, sizeof(ngx_cycle_t)); // info: 实例化 cycle 对象
    if (cycle == NULL) {
        ngx_destroy_pool(pool);
        return NULL;
    }
    cycle->pool = pool;
    cycle->log = log;
    cycle->old_cycle = old_cycle;
#if (NGX_SSL && NGX_SSL_ASYNC)
    cycle->no_ssl_init = old_cycle->no_ssl_init;
#endif

    




    return NULL;
}
