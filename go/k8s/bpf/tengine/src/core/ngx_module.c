
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


ngx_int_t
ngx_add_module(ngx_conf_t *cf, ngx_str_t *file, ngx_module_t *module,
    char **order)
{
    void               *rv;
    ngx_uint_t          i, m, before;
    ngx_core_module_t  *core_module;

    if (cf->cycle->modules_n >= ngx_max_module) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                           "too many modules loaded");
        return NGX_ERROR;
    }

    if (module->version != nginx_version) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                           "module \"%V\" version %ui instead of %ui",
                           file, module->version, (ngx_uint_t) nginx_version);
        return NGX_ERROR;
    }

    if (ngx_strcmp(module->signature, NGX_MODULE_SIGNATURE) != 0) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                           "module \"%V\" is not binary compatible",
                           file);
        return NGX_ERROR;
    }

    for (m = 0; cf->cycle->modules[m]; m++) {
        if (ngx_strcmp(cf->cycle->modules[m]->name, module->name) == 0) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                               "module \"%s\" is already loaded",
                               module->name);
            return NGX_ERROR;
        }
    }

    /*
     * if the module wasn't previously loaded, assign an index
     */

    if (module->index == NGX_MODULE_UNSET_INDEX) {
        module->index = ngx_module_index(cf->cycle);

        if (module->index >= ngx_max_module) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                               "too many modules loaded");
            return NGX_ERROR;
        }
    }

    /*
     * put the module into the cycle->modules array
     */

    before = cf->cycle->modules_n;

    if (order) {
        for (i = 0; order[i]; i++) {
            if (ngx_strcmp(order[i], module->name) == 0) {
                i++;
                break;
            }
        }

        for ( /* void */ ; order[i]; i++) {

#if 0
            ngx_log_debug2(NGX_LOG_DEBUG_CORE, cf->log, 0,
                           "module: %s before %s",
                           module->name, order[i]);
#endif

            for (m = 0; m < before; m++) {
                if (ngx_strcmp(cf->cycle->modules[m]->name, order[i]) == 0) {

                    ngx_log_debug3(NGX_LOG_DEBUG_CORE, cf->log, 0,
                                   "module: %s before %s:%i",
                                   module->name, order[i], m);

                    before = m;
                    break;
                }
            }
        }
    }

    /* put the module before modules[before] */

    if (before != cf->cycle->modules_n) {
        ngx_memmove(&cf->cycle->modules[before + 1],
                    &cf->cycle->modules[before],
                    (cf->cycle->modules_n - before) * sizeof(ngx_module_t *));
    }

    cf->cycle->modules[before] = module;
    cf->cycle->modules_n++;

    if (module->type == NGX_CORE_MODULE) {

        /*
         * we are smart enough to initialize core modules;
         * other modules are expected to be loaded before
         * initialization - e.g., http modules must be loaded
         * before http{} block
         */

        core_module = module->ctx;

        if (core_module->create_conf) {
            rv = core_module->create_conf(cf->cycle);
            if (rv == NULL) {
                return NGX_ERROR;
            }

            cf->cycle->conf_ctx[module->index] = rv;
        }
    }

    return NGX_OK;
}

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
