package main

/**
/proc/sys/net/ipv4/tcp_syncookies 值为 2 的情况下，服务器可以在某些情况下发送带有时间戳的 SYN cookies
/proc/sys/net/ipv4/tcp_fastopen 值为7，表示内核将支持 TFO 功能，并且在同一时间内，最多可以有 65535 个并发的 TFO 连接
*/
