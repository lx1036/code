



#include <ngx_config.h>
#include <ngx_core.h>





ngx_chain_t *
ngx_alloc_chain_link(ngx_pool_t *pool)
{
    ngx_chain_t  *cl;

    cl = pool->chain;

    if (cl) {
        pool->chain = cl->next;
        return cl;
    }

    cl = ngx_palloc(pool, sizeof(ngx_chain_t));
    if (cl == NULL) {
        return NULL;
    }

    return cl;
}

off_t
ngx_chain_coalesce_file(ngx_chain_t **in, off_t limit)
{
    off_t         total, size, aligned, fprev;
    ngx_fd_t      fd;
    ngx_chain_t  *cl;

    total = 0;

    cl = *in;
    fd = cl->buf->file->fd;

    do {
        size = cl->buf->file_last - cl->buf->file_pos;

        if (size > limit - total) {
            size = limit - total;

            aligned = (cl->buf->file_pos + size + ngx_pagesize - 1)
                       & ~((off_t) ngx_pagesize - 1);

            if (aligned <= cl->buf->file_last) {
                size = aligned - cl->buf->file_pos;
            }

            total += size;
            break;
        }

        total += size;
        fprev = cl->buf->file_pos + size;
        cl = cl->next;

    } while (cl
             && cl->buf->in_file
             && total < limit
             && fd == cl->buf->file->fd
             && fprev == cl->buf->file_pos);

    *in = cl;

    return total;
}

ngx_chain_t *
ngx_chain_update_sent(ngx_chain_t *in, off_t sent)
{
    off_t  size;

    for ( /* void */ ; in; in = in->next) {

        if (ngx_buf_special(in->buf)) {
            continue;
        }

        if (sent == 0) {
            break;
        }

        size = ngx_buf_size(in->buf);

        if (sent >= size) {
            sent -= size;

            if (ngx_buf_in_memory(in->buf)) {
                in->buf->pos = in->buf->last;
            }

            if (in->buf->in_file) {
                in->buf->file_pos = in->buf->file_last;
            }

            continue;
        }

        if (ngx_buf_in_memory(in->buf)) {
            in->buf->pos += (size_t) sent;
        }

        if (in->buf->in_file) {
            in->buf->file_pos += sent;
        }

        break;
    }

    return in;
}
