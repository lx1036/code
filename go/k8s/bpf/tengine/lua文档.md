

# 编译 http_lua_module

```shell
# 先安装 luajit
brew install openresty/brew/openresty

# 编译 http_lua_module
./configure --prefix=./bin --with-stream --with-debug --add-module=./modules/ngx_http_echo_module/ \
    --add-module=./modules/ngx_http_curl_module/ --with-http_lua_module \
    --with-luajit-inc=/usr/local/Cellar/openresty/1.21.4.2_1/luajit/include/luajit-2.1 \
    --with-luajit-lib=/usr/local/Cellar/openresty/1.21.4.2_1/luajit/lib

make && make install

```

