
#include <ngx_config.h>
#include <ngx_core.h>


#define NGX_MAX_DYNAMIC_MODULES  128


static ngx_uint_t ngx_module_index(ngx_cycle_t *cycle);
static ngx_uint_t ngx_module_ctx_index(ngx_cycle_t *cycle, ngx_uint_t type,
    ngx_uint_t index);


ngx_uint_t         ngx_max_module;
#if (!T_NGX_SHOW_INFO)
static
#endif
ngx_uint_t  ngx_modules_n;


static ngx_uint_t
ngx_module_index(ngx_cycle_t *cycle)
{
    ngx_uint_t     i, index;
    ngx_module_t  *module;
    index = 0;

again:
    /* find an unused index */
    for (i = 0; cycle->modules[i]; i++) {
        module = cycle->modules[i];
        if (module->index == index) {
            index++;
            goto again;
        }
    }

    /* check previous cycle */
    if (cycle->old_cycle && cycle->old_cycle->modules) {
        for (i = 0; cycle->old_cycle->modules[i]; i++) {
            module = cycle->old_cycle->modules[i];
            if (module->index == index) {
                index++;
                goto again;
            }
        }
    }

    return index;
}

ngx_int_t
ngx_preinit_modules(void)
{
    ngx_uint_t  i;

    for (i = 0; ngx_modules[i]; i++) {
        ngx_modules[i]->index = i;
        ngx_modules[i]->name = ngx_module_names[i];
    }

    ngx_modules_n = i;
    ngx_max_module = ngx_modules_n + NGX_MAX_DYNAMIC_MODULES;

    return NGX_OK;
}

ngx_int_t ngx_init_modules(ngx_cycle_t *cycle) {
    ngx_uint_t  i;
    for (i = 0; cycle->modules[i]; i++) {
        if (cycle->modules[i]->init_module) {
            if (cycle->modules[i]->init_module(cycle) != NGX_OK) {
                return NGX_ERROR;
            }
        }
    }

    return NGX_OK;
}

// copy ngx_modules into cycle->modules
ngx_int_t ngx_cycle_modules(ngx_cycle_t *cycle) {
    /*
     * create a list of modules to be used for this cycle,
     * copy static modules to it
     */
    cycle->modules = ngx_pcalloc(cycle->pool, (ngx_max_module + 1) * sizeof(ngx_module_t *));
    if (cycle->modules == NULL) {
        return NGX_ERROR;
    }

    ngx_memcpy(cycle->modules, ngx_modules, ngx_modules_n * sizeof(ngx_module_t *));
    cycle->modules_n = ngx_modules_n;

    return NGX_OK;
}

ngx_int_t ngx_count_modules(ngx_cycle_t *cycle, ngx_uint_t type) {
    ngx_uint_t     i, next, max;
    ngx_module_t  *module;
    next = 0;
    max = 0;

    /* count appropriate modules, set up their indices */
    for (i = 0; cycle->modules[i]; i++) {
        module = cycle->modules[i];
        if (module->type != type) {
            continue;
        }

        if (module->ctx_index != NGX_MODULE_UNSET_INDEX) {
            /* if ctx_index was assigned, preserve it */

            if (module->ctx_index > max) {
                max = module->ctx_index;
            }

            if (module->ctx_index == next) {
                next++;
            }

            continue;
        }

        /* search for some free index */
        module->ctx_index = ngx_module_ctx_index(cycle, type, next);

        if (module->ctx_index > max) {
            max = module->ctx_index;
        }

        next = module->ctx_index + 1;
    }

    /*
     * make sure the number returned is big enough for previous
     * cycle as well, else there will be problems if the number
     * will be stored in a global variable (as it's used to be)
     * and we'll have to roll back to the previous cycle
     */

    if (cycle->old_cycle && cycle->old_cycle->modules) {
        for (i = 0; cycle->old_cycle->modules[i]; i++) {
            module = cycle->old_cycle->modules[i];

            if (module->type != type) {
                continue;
            }

            if (module->ctx_index > max) {
                max = module->ctx_index;
            }
        }
    }

    /* prevent loading of additional modules */
    cycle->modules_used = 1;

    return max + 1;
}

static ngx_uint_t
ngx_module_ctx_index(ngx_cycle_t *cycle, ngx_uint_t type, ngx_uint_t index)
{
    ngx_uint_t     i;
    ngx_module_t  *module;

again:
    /* find an unused ctx_index */
    for (i = 0; cycle->modules[i]; i++) {
        module = cycle->modules[i];

        if (module->type != type) {
            continue;
        }

        if (module->ctx_index == index) {
            index++;
            goto again;
        }
    }

    /* check previous cycle */
    if (cycle->old_cycle && cycle->old_cycle->modules) {
        for (i = 0; cycle->old_cycle->modules[i]; i++) {
            module = cycle->old_cycle->modules[i];
            if (module->type != type) {
                continue;
            }

            if (module->ctx_index == index) {
                index++;
                goto again;
            }
        }
    }

    return index;
}
