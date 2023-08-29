
#include <ngx_config.h>
#include <ngx_core.h>


static ngx_int_t ngx_test_full_name(ngx_str_t *name);


static ngx_atomic_t   temp_number = 0;
ngx_atomic_t         *ngx_temp_number = &temp_number;
ngx_atomic_int_t      ngx_random_number = 123456;



ngx_int_t
ngx_get_full_name(ngx_pool_t *pool, ngx_str_t *prefix, ngx_str_t *name)
{
    size_t      len;
    u_char     *p, *n;
    ngx_int_t   rc;

    rc = ngx_test_full_name(name);

    if (rc == NGX_OK) {
        return rc;
    }

    len = prefix->len;
    n = ngx_pnalloc(pool, len + name->len + 1);
    if (n == NULL) {
        return NGX_ERROR;
    }

    p = ngx_cpymem(n, prefix->data, len);
    ngx_cpystrn(p, name->data, name->len + 1);

    name->len += len;
    name->data = n;

    return NGX_OK;
}

static ngx_int_t
ngx_test_full_name(ngx_str_t *name)
{
    if (name->data[0] == '/') {
        return NGX_OK;
    }

    return NGX_DECLINED;
}

ngx_int_t ngx_create_paths(ngx_cycle_t *cycle, ngx_uid_t user) {
    ngx_err_t         err;
    ngx_uint_t        i;
    ngx_path_t      **path;

    path = cycle->paths.elts;
    for (i = 0; i < cycle->paths.nelts; i++) {
        if (ngx_create_dir(path[i]->name.data, 0700) == NGX_FILE_ERROR) {
            err = ngx_errno;
            if (err != NGX_EEXIST) {
                ngx_log_error(NGX_LOG_EMERG, cycle->log, err, ngx_create_dir_n " \"%s\" failed", path[i]->name.data);
                return NGX_ERROR;
            }
        }

        if (user == (ngx_uid_t) NGX_CONF_UNSET_UINT) {
            continue;
        }

        ngx_file_info_t   fi;
        if (ngx_file_info(path[i]->name.data, &fi) == NGX_FILE_ERROR) {
            ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno, ngx_file_info_n " \"%s\" failed", path[i]->name.data);
            return NGX_ERROR;
        }

        if (fi.st_uid != user) {
            if (chown((const char *) path[i]->name.data, user, -1) == -1) {
                ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno, "chown(\"%s\", %d) failed", path[i]->name.data, user);
                return NGX_ERROR;
            }
        }

        if ((fi.st_mode & (S_IRUSR|S_IWUSR|S_IXUSR)) != (S_IRUSR|S_IWUSR|S_IXUSR)) {
            fi.st_mode |= (S_IRUSR|S_IWUSR|S_IXUSR);
            if (chmod((const char *) path[i]->name.data, fi.st_mode) == -1) {
                ngx_log_error(NGX_LOG_EMERG, cycle->log, ngx_errno,
                              "chmod() \"%s\" failed", path[i]->name.data);
                return NGX_ERROR;
            }
        }
    }

    return NGX_OK;
}


