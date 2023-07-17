


# proxy-pass-module
官网：
http: https://nginx.org/en/docs/http/ngx_http_proxy_module.html
tcp/udp: https://nginx.org/en/docs/stream/ngx_stream_proxy_module.html
代码：
https://github.com/lx1036/tengine/blob/feature/memory-leak/src/http/modules/ngx_http_proxy_module.c
https://github.com/lx1036/tengine/blob/feature/memory-leak/src/stream/ngx_stream_proxy_module.c

```conf
location / {
    proxy_pass       http://localhost:8000;
    proxy_set_header Host      $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```


