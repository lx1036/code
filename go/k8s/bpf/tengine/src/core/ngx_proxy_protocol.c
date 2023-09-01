




#include <ngx_config.h>
#include <ngx_core.h>

#define NGX_PROXY_PROTOCOL_AF_INET          1
#define NGX_PROXY_PROTOCOL_AF_INET6         2


#define ngx_proxy_protocol_parse_uint16(p)                                    \
    ( ((uint16_t) (p)[0] << 8)                                                \
    + (           (p)[1]) )

#define ngx_proxy_protocol_parse_uint32(p)                                    \
    ( ((uint32_t) (p)[0] << 24)                                               \
    + (           (p)[1] << 16)                                               \
    + (           (p)[2] << 8)                                                \
    + (           (p)[3]) )

typedef struct {
    u_char                                  signature[12];
    u_char                                  version_command; // The next byte (the 13th one) is the protocol version and command
    u_char                                  family_transport;
    u_char                                  len[2];
} ngx_proxy_protocol_header_t;

typedef struct {
    u_char                                  src_addr[4];
    u_char                                  dst_addr[4];
    u_char                                  src_port[2];
    u_char                                  dst_port[2];
} ngx_proxy_protocol_inet_addrs_t;

typedef struct {
    // ngx_proxy_protocol_header_t header;
    u_char                                  signature[12];
    u_char                                  version_command;
    u_char                                  family_transport;
    uint16_t                                  len;
    union {
        struct {
            uint32_t                                  src_addr;
            uint32_t                                  dst_addr;
            uint16_t                                  src_port;
            uint16_t                                  dst_port;
        } ipv4;
        struct {
            uint8_t                                  src_addr[16];
            uint8_t                                  dst_addr[16];
            uint16_t                                  src_port;
            uint16_t                                  dst_port;
        } ipv6;
        // ngx_proxy_protocol_inet_addrs_t ipv4;
        // ngx_proxy_protocol_inet6_addrs_t ipv6;
    };
} ngx_proxy_protocol_v2_t;

typedef struct {
    u_char                                  type;
    u_char                                  len[2];
} ngx_proxy_protocol_tlv_t;

typedef struct {
    u_char                                  client;
    u_char                                  verify[4];
} ngx_proxy_protocol_tlv_ssl_t;


typedef struct {
    ngx_str_t                               name;
    ngx_uint_t                              type;
} ngx_proxy_protocol_tlv_entry_t;

static ngx_proxy_protocol_tlv_entry_t  ngx_proxy_protocol_tlv_entries[] = {
    { ngx_string("alpn"),       0x01 },
    { ngx_string("authority"),  0x02 },
    { ngx_string("unique_id"),  0x05 },
    { ngx_string("ssl"),        0x20 },
    { ngx_string("netns"),      0x30 },
    { ngx_null_string,          0x00 }
};

static ngx_proxy_protocol_tlv_entry_t  ngx_proxy_protocol_tlv_ssl_entries[] = {
    { ngx_string("version"),    0x21 },
    { ngx_string("cn"),         0x22 },
    { ngx_string("cipher"),     0x23 },
    { ngx_string("sig_alg"),    0x24 },
    { ngx_string("key_alg"),    0x25 },
    { ngx_null_string,          0x00 }
};

static u_char *ngx_proxy_protocol_read_addr(ngx_connection_t *c, u_char *p,
    u_char *last, ngx_str_t *addr);
static u_char *ngx_proxy_protocol_read_port(u_char *p, u_char *last,
    in_port_t *port, u_char sep);
static u_char *ngx_proxy_protocol_v2_read(ngx_connection_t *c, u_char *buf,
    u_char *last);
static ngx_int_t ngx_proxy_protocol_lookup_tlv(ngx_connection_t *c,
    ngx_str_t *tlvs, ngx_uint_t type, ngx_str_t *value);
u_char *ngx_proxy_protocol_v2_write(ngx_connection_t *c, u_char *buf, u_char *last);


ngx_int_t
ngx_proxy_protocol_get_tlv(ngx_connection_t *c, ngx_str_t *name,
    ngx_str_t *value)
{
    u_char                          *p;
    size_t                           n;
    uint32_t                         verify;
    ngx_str_t                        ssl, *tlvs;
    ngx_int_t                        rc, type;
    ngx_proxy_protocol_tlv_ssl_t    *tlv_ssl;
    ngx_proxy_protocol_tlv_entry_t  *te;

    if (c->proxy_protocol == NULL) {
        return NGX_DECLINED;
    }

    ngx_log_debug1(NGX_LOG_DEBUG_CORE, c->log, 0,
                   "PROXY protocol v2 get tlv \"%V\"", name);

    te = ngx_proxy_protocol_tlv_entries;
    tlvs = &c->proxy_protocol->tlvs;

    p = name->data;
    n = name->len;

    if (n >= 4 && p[0] == 's' && p[1] == 's' && p[2] == 'l' && p[3] == '_') {

        rc = ngx_proxy_protocol_lookup_tlv(c, tlvs, 0x20, &ssl);
        if (rc != NGX_OK) {
            return rc;
        }

        if (ssl.len < sizeof(ngx_proxy_protocol_tlv_ssl_t)) {
            return NGX_ERROR;
        }

        p += 4;
        n -= 4;

        if (n == 6 && ngx_strncmp(p, "verify", 6) == 0) {

            tlv_ssl = (ngx_proxy_protocol_tlv_ssl_t *) ssl.data;
            verify = ngx_proxy_protocol_parse_uint32(tlv_ssl->verify);

            value->data = ngx_pnalloc(c->pool, NGX_INT32_LEN);
            if (value->data == NULL) {
                return NGX_ERROR;
            }

            value->len = ngx_sprintf(value->data, "%uD", verify)
                         - value->data;
            return NGX_OK;
        }

        ssl.data += sizeof(ngx_proxy_protocol_tlv_ssl_t);
        ssl.len -= sizeof(ngx_proxy_protocol_tlv_ssl_t);

        te = ngx_proxy_protocol_tlv_ssl_entries;
        tlvs = &ssl;
    }

    if (n >= 2 && p[0] == '0' && p[1] == 'x') {
        type = ngx_hextoi(p + 2, n - 2);
        if (type == NGX_ERROR) {
            ngx_log_error(NGX_LOG_ERR, c->log, 0,
                          "invalid PROXY protocol TLV \"%V\"", name);
            return NGX_ERROR;
        }

        return ngx_proxy_protocol_lookup_tlv(c, tlvs, type, value);
    }

    for ( /* void */ ; te->type; te++) {
        if (te->name.len == n && ngx_strncmp(te->name.data, p, n) == 0) {
            return ngx_proxy_protocol_lookup_tlv(c, tlvs, te->type, value);
        }
    }

    ngx_log_error(NGX_LOG_ERR, c->log, 0,
                  "unknown PROXY protocol TLV \"%V\"", name);

    return NGX_DECLINED;
}

static ngx_int_t
ngx_proxy_protocol_lookup_tlv(ngx_connection_t *c, ngx_str_t *tlvs,
    ngx_uint_t type, ngx_str_t *value) {
    u_char                    *p;
    size_t                     n, len;
    ngx_proxy_protocol_tlv_t  *tlv;

    ngx_log_debug1(NGX_LOG_DEBUG_CORE, c->log, 0,
                   "PROXY protocol v2 lookup tlv:%02xi", type);

    p = tlvs->data;
    n = tlvs->len;

    while (n) {
        if (n < sizeof(ngx_proxy_protocol_tlv_t)) {
            ngx_log_error(NGX_LOG_ERR, c->log, 0, "broken PROXY protocol TLV");
            return NGX_ERROR;
        }

        tlv = (ngx_proxy_protocol_tlv_t *) p;
        len = ngx_proxy_protocol_parse_uint16(tlv->len);

        p += sizeof(ngx_proxy_protocol_tlv_t);
        n -= sizeof(ngx_proxy_protocol_tlv_t);

        if (n < len) {
            ngx_log_error(NGX_LOG_ERR, c->log, 0, "broken PROXY protocol TLV");
            return NGX_ERROR;
        }

        if (tlv->type == type) {
            value->data = p;
            value->len = len;
            return NGX_OK;
        }

        p += len;
        n -= len;
    }

    return NGX_DECLINED;
}

u_char *
ngx_proxy_protocol_read(ngx_connection_t *c, u_char *buf, u_char *last)
{
    size_t                 len;
    u_char                *p;
    ngx_proxy_protocol_t  *pp;

    static const u_char signature[] = "\r\n\r\n\0\r\nQUIT\n"; // \x0D \x0A \x0D \x0A \x00 \x0D \x0A \x51 \x55 \x49 \x54 \x0A

    p = buf; // 指针拷贝，修改 p 不影响 buf: "PROXY TCP4 127.0.0.1 127.0.0.1 56935 12345"
    len = last - buf;

    if (len >= sizeof(ngx_proxy_protocol_header_t)
        && ngx_memcmp(p, signature, sizeof(signature) - 1) == 0)
    {
        return ngx_proxy_protocol_v2_read(c, buf, last);
    }

    if (len < 8 || ngx_strncmp(p, "PROXY ", 6) != 0) {
        goto invalid;
    }

    p += 6;
    len -= 6;

    if (len >= 7 && ngx_strncmp(p, "UNKNOWN", 7) == 0) {
        ngx_log_debug0(NGX_LOG_DEBUG_CORE, c->log, 0,
                       "PROXY protocol unknown protocol");
        p += 7;
        goto skip;
    }

    if (len < 5 || ngx_strncmp(p, "TCP", 3) != 0
        || (p[3] != '4' && p[3] != '6') || p[4] != ' ')
    {
        goto invalid;
    }

    p += 5;

    pp = ngx_pcalloc(c->pool, sizeof(ngx_proxy_protocol_t));
    if (pp == NULL) {
        return NULL;
    }

    p = ngx_proxy_protocol_read_addr(c, p, last, &pp->src_addr); // 因为可能是 ipv4/ipv6 地址，所以只能这么取值
    if (p == NULL) {
        goto invalid;
    }

    p = ngx_proxy_protocol_read_addr(c, p, last, &pp->dst_addr);
    if (p == NULL) {
        goto invalid;
    }

    p = ngx_proxy_protocol_read_port(p, last, &pp->src_port, ' ');
    if (p == NULL) {
        goto invalid;
    }

    p = ngx_proxy_protocol_read_port(p, last, &pp->dst_port, CR);
    if (p == NULL) {
        goto invalid;
    }

    if (p == last) {
        goto invalid;
    }

    if (*p++ != LF) {
        goto invalid;
    }

    ngx_log_debug4(NGX_LOG_DEBUG_CORE, c->log, 0,
                   "PROXY protocol src: %V %d, dst: %V %d",
                   &pp->src_addr, pp->src_port, &pp->dst_addr, pp->dst_port);

    c->proxy_protocol = pp;

    return p;

skip:

    for ( /* void */ ; p < last - 1; p++) {
        if (p[0] == CR && p[1] == LF) {
            return p + 2;
        }
    }

invalid:

    for (p = buf; p < last; p++) {
        if (*p == CR || *p == LF) {
            break;
        }
    }

    ngx_log_error(NGX_LOG_ERR, c->log, 0,
                  "broken header: \"%*s\"", (size_t) (p - buf), buf);

    return NULL;
}


static u_char * ngx_proxy_protocol_read_addr(ngx_connection_t *c, u_char *p, u_char *last, ngx_str_t *addr) {
    size_t  len;
    u_char  ch, *pos;

    pos = p;

    for ( ;; ) {
        if (p == last) {
            return NULL;
        }

        ch = *p++;

        if (ch == ' ') {
            break;
        }

        if (ch != ':' && ch != '.'
            && (ch < 'a' || ch > 'f')
            && (ch < 'A' || ch > 'F')
            && (ch < '0' || ch > '9'))
        {
            return NULL;
        }
    }

    len = p - pos - 1;
    addr->data = ngx_pnalloc(c->pool, len);
    if (addr->data == NULL) {
        return NULL;
    }

    ngx_memcpy(addr->data, pos, len);
    addr->len = len;

    return p;
}


static u_char * ngx_proxy_protocol_read_port(u_char *p, u_char *last, in_port_t *port, u_char sep) {
    size_t      len;
    u_char     *pos;
    ngx_int_t   n;

    pos = p;

    for ( ;; ) {
        if (p == last) {
            return NULL;
        }

        if (*p++ == sep) {
            break;
        }
    }

    len = p - pos - 1;

    n = ngx_atoi(pos, len);
    if (n < 0 || n > 65535) {
        return NULL;
    }

    *port = (in_port_t) n;

    return p;
}

// ./configure --prefix=./bin --with-stream --with-debug --add-module=./modules/ngx_http_echo_module/ && make && make install
// ./bin/sbin/nginx -c conf/proxy-pass-module.conf  -p .
// https://github.com/dedok/nginx-stream-proxy-protocol-v2/blob/main/stream-proxy-protocol-v2-release-1.19.8.patch
u_char *
ngx_proxy_protocol_v2_write(ngx_connection_t *c, u_char *buf, u_char *last) {
    size_t                          len = 0;
    struct sockaddr                 *src, *dst;
    ngx_proxy_protocol_v2_t         *header;
    // ngx_uint_t family = c->sockaddr->sa_family;
    ngx_int_t                       v6_used = 0;

    if (c->sockaddr->sa_family == AF_INET && c->local_sockaddr->sa_family == AF_INET) { // ipv4
        if (last - buf < NGX_PROXY_PROTOCOL_V2_HDR_LEN_INET) {
            ngx_log_error(NGX_LOG_ERR, c->log, 0,  "too small buffer for PROXY protocol");
            return NULL;
        }
    } else { // ipv6
        if (last - buf < NGX_PROXY_PROTOCOL_V2_HDR_LEN_INET6) {
            ngx_log_error(NGX_LOG_ERR, c->log, 0,  "too small buffer for PROXY protocol");
            return NULL;
        }
    }

    if (ngx_connection_local_sockaddr(c, NULL, 0) != NGX_OK) {
        return NULL;
    }

    header = (ngx_proxy_protocol_v2_t *) buf;
    header->len = 0;
    static const u_char signature[] = "\r\n\r\n\0\r\nQUIT\n";
    ngx_memcpy(header->signature, signature, sizeof(signature)-1);
    header->family_transport = NGX_PROXY_PROTOCOL_V2_TRANS_STREAM;
    header->version_command = NGX_PROXY_PROTOCOL_V2_CMD_PROXY;
    
    // data->header.len = 0x00;
    // ngx_memcpy(&data->header.len, 0, 2);

    src = c->sockaddr;
    dst = c->local_sockaddr;

    switch (src->sa_family) {
    case AF_INET:
        // data->ipv4.src_addr = ((struct sockaddr_in *) src)->sin_addr.s_addr;
        // data->ipv4.src_port = ((struct sockaddr_in *) src)->sin_port; 
        header->ipv4.src_addr = ((struct sockaddr_in *) src)->sin_addr.s_addr;
        header->ipv4.src_port = ((struct sockaddr_in *) src)->sin_port;
        // ngx_memcpy(&data->ipv4.src_addr, ((struct sockaddr_in *) src)->sin_addr.s_addr, 4);
        // ngx_memcpy(&data->ipv4.src_port, ((struct sockaddr_in *) src)->sin_port, 2);

        break;
#if (NGX_HAVE_INET6)
    case AF_INET6:
        v6_used = 1;
        // ngx_memcpy(data->addrs.ip6.src_addr, &((struct sockaddr_in6 *) src)->sin6_addr, 16);
        // data->addrs.ip6.src_port = ((struct sockaddr_in6 *) src)->sin6_port;
        break;
#endif    
    default: 
        ngx_log_error(NGX_LOG_ERR, c->log, 0,  "PPv2 unsupported src address family %ui", src->sa_family); 
        goto unspec;
    }

    switch (dst->sa_family) {
    case AF_INET: 
        // data->addrs.ip4.dst_addr = ((struct sockaddr_in *) dst)->sin_addr.s_addr;
        // data->addrs.ip4.dst_port = ((struct sockaddr_in *) dst)->sin_port;
        header->ipv4.dst_addr = ((struct sockaddr_in *) dst)->sin_addr.s_addr;
        header->ipv4.dst_port = ((struct sockaddr_in *) dst)->sin_port;
        // ngx_memcpy(&data->ipv4.dst_addr, ((struct sockaddr_in *) dst)->sin_addr.s_addr, 4);
        // ngx_memcpy(&data->ipv4.dst_port, ((struct sockaddr_in *) dst)->sin_port, 2);
        break;
#if (NGX_HAVE_INET6)
    case AF_INET6:
        v6_used = 1;
        // ngx_memcpy(data->addrs.ip6.dst_addr, &((struct sockaddr_in6 *) dst)->sin6_addr, 16);
        // data->addrs.ip6.dst_port = ((struct sockaddr_in6 *) dst)->sin6_port;
        break;
#endif    
    default: 
        ngx_log_error(NGX_LOG_ERR, c->log, 0,  "PPv2 unsupported dst address family %ui", dst->sa_family); 
        goto unspec;
    }

    if (v6_used) {
        header->family_transport |= NGX_PROXY_PROTOCOL_V2_FAM_INET6;
        // data->header.family_transport |= NGX_PROXY_PROTOCOL_V2_FAM_INET6;
        len = NGX_PROXY_PROTOCOL_V2_HDR_LEN_INET6;
    } else {
        header->family_transport |= NGX_PROXY_PROTOCOL_V2_FAM_INET;
        // data->header.family_transport |= NGX_PROXY_PROTOCOL_V2_FAM_INET;
        len = NGX_PROXY_PROTOCOL_V2_HDR_LEN_INET;
    }

    // ngx_memcpy(&data->header.len, len - NGX_PROXY_PROTOCOL_V2_HDR_LEN, 2);
    header->len = htons(len - NGX_PROXY_PROTOCOL_V2_HDR_LEN);
    // data->header.len = htons(len - NGX_PROXY_PROTOCOL_V2_HDR_LEN);
    return buf + len;

unspec:
    header->family_transport |= NGX_PROXY_PROTOCOL_V2_FAM_UNSPEC;
    header->len = 0;
    // data->header.len = 0;
    // ngx_memcpy(&data->header.len, 0, 2);
    return buf + NGX_PROXY_PROTOCOL_V2_HDR_LEN;
}

static u_char *
ngx_proxy_protocol_v2_read(ngx_connection_t *c, u_char *buf, u_char *last)
{
    u_char                             *end;
    size_t                              len;
    socklen_t                           socklen;
    ngx_uint_t                          version, command, family, transport;
    ngx_sockaddr_t                      src_sockaddr, dst_sockaddr;
    ngx_proxy_protocol_t               *pp;
    ngx_proxy_protocol_header_t        *header;
    ngx_proxy_protocol_inet_addrs_t    *in;
    // C 语言可以类型转换 字节->结构体，Go 需要一个个字段获取，类似 msg.Len = binary.BigEndian.Uint16(data[16:18]) msg.Type = data[18]
    header = (ngx_proxy_protocol_header_t *) buf;

    buf += sizeof(ngx_proxy_protocol_header_t);

    version = header->version_command >> 4; // 必须格式为 \x2{command}

    if (version != 2) {
        ngx_log_error(NGX_LOG_ERR, c->log, 0,
                      "unknown PROXY protocol version: %ui", version);
        return NULL;
    }

    len = ngx_proxy_protocol_parse_uint16(header->len);

    if ((size_t) (last - buf) < len) {
        ngx_log_error(NGX_LOG_ERR, c->log, 0, "header is too large");
        return NULL;
    }

    end = buf + len;

    command = header->version_command & 0x0f;

    /* only PROXY is supported */
    if (command != 1) {
        ngx_log_debug1(NGX_LOG_DEBUG_CORE, c->log, 0,
                       "PROXY protocol v2 unsupported command %ui", command);
        return end;
    }

    transport = header->family_transport & 0x0f;

    /* only STREAM is supported */
    if (transport != 1) { // TODO: 还需要支持 UDP PPv2, TCP or UNIX_STREAM, 不支持 UDP or UNIX_DGRAM
        ngx_log_debug1(NGX_LOG_DEBUG_CORE, c->log, 0,
                       "PROXY protocol v2 unsupported transport %ui",
                       transport);
        return end;
    }

    pp = ngx_pcalloc(c->pool, sizeof(ngx_proxy_protocol_t));
    if (pp == NULL) {
        return NULL;
    }

    family = header->family_transport >> 4;

    switch (family) {

    case NGX_PROXY_PROTOCOL_AF_INET:

        if ((size_t) (end - buf) < sizeof(ngx_proxy_protocol_inet_addrs_t)) {
            return NULL;
        }

        in = (ngx_proxy_protocol_inet_addrs_t *) buf;

        src_sockaddr.sockaddr_in.sin_family = AF_INET;
        src_sockaddr.sockaddr_in.sin_port = 0;
        ngx_memcpy(&src_sockaddr.sockaddr_in.sin_addr, in->src_addr, 4);

        dst_sockaddr.sockaddr_in.sin_family = AF_INET;
        dst_sockaddr.sockaddr_in.sin_port = 0;
        ngx_memcpy(&dst_sockaddr.sockaddr_in.sin_addr, in->dst_addr, 4);

        pp->src_port = ngx_proxy_protocol_parse_uint16(in->src_port);
        pp->dst_port = ngx_proxy_protocol_parse_uint16(in->dst_port);

        socklen = sizeof(struct sockaddr_in);

        buf += sizeof(ngx_proxy_protocol_inet_addrs_t);

        break;

    default:
        ngx_log_debug1(NGX_LOG_DEBUG_CORE, c->log, 0,
                       "PROXY protocol v2 unsupported address family %ui",
                       family);
        return end;
    }

    pp->src_addr.data = ngx_pnalloc(c->pool, NGX_SOCKADDR_STRLEN);
    if (pp->src_addr.data == NULL) {
        return NULL;
    }

    pp->src_addr.len = ngx_sock_ntop(&src_sockaddr.sockaddr, socklen,
                                     pp->src_addr.data, NGX_SOCKADDR_STRLEN, 0); // copy ip -> pp->src_addr.data

    pp->dst_addr.data = ngx_pnalloc(c->pool, NGX_SOCKADDR_STRLEN);
    if (pp->dst_addr.data == NULL) {
        return NULL;
    }

    pp->dst_addr.len = ngx_sock_ntop(&dst_sockaddr.sockaddr, socklen,
                                     pp->dst_addr.data, NGX_SOCKADDR_STRLEN, 0);

    ngx_log_debug4(NGX_LOG_DEBUG_CORE, c->log, 0,
                   "PROXY protocol v2 src: %V %d, dst: %V %d",
                   &pp->src_addr, pp->src_port, &pp->dst_addr, pp->dst_port);

    if (buf < end) {
        pp->tlvs.data = ngx_pnalloc(c->pool, end - buf);
        if (pp->tlvs.data == NULL) {
            return NULL;
        }

        ngx_memcpy(pp->tlvs.data, buf, end - buf);
        pp->tlvs.len = end - buf;
    }

    c->proxy_protocol = pp;

    return end;
}


