

#include <ngx_config.h>
#include <ngx_core.h>


static ngx_pool_t             *ngx_regex_pool;
static ngx_list_t             *ngx_regex_studies;
static ngx_uint_t              ngx_regex_direct_alloc;


static void * ngx_libc_cdecl ngx_regex_malloc(size_t size);
static void ngx_libc_cdecl ngx_regex_free(void *p);

void
ngx_regex_init(void)
{
#if !(NGX_PCRE2)    
    pcre_malloc = ngx_regex_malloc;
    pcre_free = ngx_regex_free;
#endif
}

static void * ngx_libc_cdecl
ngx_regex_malloc(size_t size)
{
    if (ngx_regex_pool) {
        return ngx_palloc(ngx_regex_pool, size);
    }

    return NULL;
}


static void ngx_libc_cdecl
ngx_regex_free(void *p)
{
    return;
}