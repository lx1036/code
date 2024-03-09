
/*
 * Copyright (C) Roman Arutyunyan
 * Copyright (C) Nginx, Inc.
 */


#ifndef _NGX_PROXY_PROTOCOL_H_INCLUDED_
#define _NGX_PROXY_PROTOCOL_H_INCLUDED_


#include <ngx_config.h>
#include <ngx_core.h>


#define NGX_PROXY_PROTOCOL_V1_MAX_HEADER  107
#define NGX_PROXY_PROTOCOL_MAX_HEADER     4096

#define NGX_MINI_PROXY_PROTOCOL_V2_VERSION 8 // 只有 client_ip:client_port 的 PPv2 报文

#define NGX_PROXY_PROTOCOL_V2_TRANS_STREAM     0x01
#define NGX_PROXY_PROTOCOL_V2_TRANS_DGRAM      0x02
#define NGX_PROXY_PROTOCOL_V2_CMD_PROXY        (0x20 | 0x01)
#define NGX_PROXY_PROTOCOL_V2_HDR_LEN          16
#define NGX_PROXY_PROTOCOL_V2_HDR_LEN_INET \
                (NGX_PROXY_PROTOCOL_V2_HDR_LEN + (4 + 4 + 2 + 2))
#define NGX_PROXY_PROTOCOL_V2_HDR_LEN_INET6 \
                (NGX_PROXY_PROTOCOL_V2_HDR_LEN + (16 + 16 + 2 + 2))
#define NGX_PROXY_PROTOCOL_V2_FAM_UNSPEC       0x00
#define NGX_PROXY_PROTOCOL_V2_FAM_INET         0x10
#define NGX_PROXY_PROTOCOL_V2_FAM_INET6        0x20

struct ngx_proxy_protocol_s {
    ngx_str_t           src_addr;
    ngx_str_t           dst_addr;
    in_port_t           src_port;
    in_port_t           dst_port;
    ngx_str_t           tlvs;
};


u_char *ngx_proxy_protocol_read(ngx_connection_t *c, u_char *buf,
    u_char *last);
u_char *ngx_proxy_protocol_write(ngx_connection_t *c, u_char *buf,
    u_char *last, ngx_uint_t version);

u_char *ngx_proxy_protocol_v2_write(ngx_connection_t *c, u_char *buf, u_char *last);

ngx_int_t ngx_proxy_protocol_get_tlv(ngx_connection_t *c, ngx_str_t *name,
    ngx_str_t *value);


#endif /* _NGX_PROXY_PROTOCOL_H_INCLUDED_ */
