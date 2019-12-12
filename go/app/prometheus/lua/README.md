
# Install OpenResty
```shell script
brew install openresty
ln -s /usr/local/Cellar/openresty/1.15.8.2/nginx/sbin/nginx /usr/local/bin/nginx
nginx -c /Users/liuxiang/Code/lx1036/code/go/app/prometheus/lua/nginx.conf
```

# Docs
https://moonbingbing.gitbooks.io/openresty-best-practices/
https://github.com/moonbingbing/openresty-best-practices
https://github.com/knyar/nginx-lua-prometheus
https://github.com/mpeterv/luacheck

## OpenResty

### 官方文档
**[OpenResty 执行阶段](https://moonbingbing.gitbooks.io/openresty-best-practices/ngx_lua/phase.html)**
**[openresty/lua-nginx-module](https://github.com/openresty/lua-nginx-module)**


## Lua 单元测试
**[luaunit](https://github.com/bluebird75/luaunit)**
```shell script
brew install luarocks
brew install luaunit
luarocks install luacheck
luacheck --globals ngx -- prometheus.lua
lua prometheus_test.lua
```
