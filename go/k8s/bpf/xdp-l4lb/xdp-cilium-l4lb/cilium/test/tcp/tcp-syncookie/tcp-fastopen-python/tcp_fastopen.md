
# TCP FastOpen
> 作用：首次 client 在 syn TCP Option 里加上 tfo cookie request(tcp.options.tfo.request)，然后 server 回的 synack 包里
> tfo option 里加上 Fast Open Cookie(69bd7321bdbafb15)，然后后面 client 再次发起新的链接时，新的 syn 报文带上这个 cookie 和
> 请求报文如 HTTP/TCP 请求，即 HTTP/TCP 请求在第二次之后直接在 syn 包里发送请求，这样无需等待 synack + ack 这一个 rtt 时间。
>
> 结论：首次 syn 获取 tfo cookie 之后，后续每次连接 syn 直接带上数据报文，减少了一个 rtt 时间，提高了性能。 

首先需要客户端和服务端 linux 内核开启 tcp fastopen:
```
# number 3 will add support for both TFO client and server
echo 3 > /proc/sys/net/ipv4/tcp_fastopen
sysctl -w net.ipv4.tcp_fastopen=3 # 或者

net.ipv4.tcp_fastopen = 3
```


## nginx 演示
HTTP 开启 TFO

```
server {
    listen 8099 fastopen=256;
    root /var/www/html;
    client_max_body_size 100g; 
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}

```

然后 curl 客户端访问：
```
curl --tcp-fastopen localhost:8099
```



## python 代码演示

服务端 server：
```
python3 server.py
```

客户端 client:
```
python3 client.py
```

