#!/bin/bash

tubectl load
tubectl bind foo tcp 127.0.0.1 4321
tubectl bindings
nc -k -l 127.0.0.1 9999 &
# 当前 server pid 打开的所有 socket_fd 中，127.0.0.1:9999 的那个 socket_fd
echo $!
tubectl register-pid $! foo tcp 127.0.0.1 9999
# redirect tcp://127.0.0.1:4321 > tcp://127.0.0.1:9999
echo hello | nc -q 1 127.0.0.1 4321

