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

python3 client-port.py
CGO_ENABLED=0 go run . load
CGO_ENABLED=0 go run . bind --label=foo --protocol=tcp --ip-prefix=127.0.0.1/32 --port=4321
CGO_ENABLED=0 go run . bind --label=bar --protocol=tcp --ip-prefix=127.0.0.1/32 --port=4322
CGO_ENABLED=0 go run . register-pid --pid=521347 --label=foo --protocol=tcp --ip=127.0.0.1 --port=9999
CGO_ENABLED=0 go run . register-pid --pid=522598 --label=bar --protocol=tcp --ip=127.0.0.1 --port=10000

echo hello | nc -q 1 127.0.0.1 4322
CGO_ENABLED=0 go run . unload
